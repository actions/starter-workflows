package worker

import (
	"encoding/json"
	"fmt"
	"time"

	gocontext "context"

	"github.com/Jeffail/tunny"
	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"
)

// AMQPJobQueue is a JobQueue that uses AMQP
type AMQPJobQueue struct {
	conn            *amqp.Connection
	queue           string
	priority        int
	withLogSharding bool

	stateUpdatePool *tunny.Pool

	DefaultLanguage, DefaultDist, DefaultArch, DefaultGroup, DefaultOS string
}

// NewAMQPJobQueue creates a AMQPJobQueue backed by the given AMQP connections and
// connects to the AMQP queue with the given name. The queue will be declared
// in AMQP when this function is called, so an error could be raised if the
// queue already exists, but with different attributes than we expect.
func NewAMQPJobQueue(conn *amqp.Connection, queue string, stateUpdatePoolSize int, sharded bool) (*AMQPJobQueue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare("reporting", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	_, err = channel.QueueDeclare("reporting.jobs.builds", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if sharded {
		// This exchange should be declared as sharded using a policy that matches its name.
		err = channel.ExchangeDeclare("reporting.jobs.logs_sharded", "x-modulus-hash", true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = channel.QueueDeclare("reporting.jobs.logs", true, false, false, false, nil)
		if err != nil {
			return nil, err
		}

		err = channel.QueueBind("reporting.jobs.logs", "reporting.jobs.logs", "reporting", false, nil)
		if err != nil {
			return nil, err
		}
	}

	err = channel.Close()
	if err != nil {
		return nil, err
	}

	stateUpdatePool := newStateUpdatePool(conn, stateUpdatePoolSize)

	go reportPoolMetrics("state_update_pool", stateUpdatePool)

	return &AMQPJobQueue{
		conn:            conn,
		queue:           queue,
		withLogSharding: sharded,

		stateUpdatePool: stateUpdatePool,
	}, nil
}

func newStateUpdatePool(conn *amqp.Connection, poolSize int) *tunny.Pool {
	return tunny.New(poolSize, func() tunny.Worker {
		stateUpdateChan, err := conn.Channel()
		if err != nil {
			logrus.WithField("err", err).Panic("could not create state update amqp channel")
		}
		return &amqpStateUpdateWorker{
			stateUpdateChan: stateUpdateChan,
		}
	})
}

func reportPoolMetrics(poolName string, pool *tunny.Pool) {
	for {
		metrics.Gauge(fmt.Sprintf("travis.worker.%s.queue_length", poolName), pool.QueueLength())
		metrics.Gauge(fmt.Sprintf("travis.worker.%s.pool_size", poolName), int64(pool.GetSize()))
		time.Sleep(10 * time.Second)
	}
}

// Jobs creates a new consumer on the queue, and returns three channels. The
// first channel gets sent every BuildJob that we receive from AMQP. The
// stopChan is a channel that can be closed in order to stop the consumer.
func (q *AMQPJobQueue) Jobs(ctx gocontext.Context) (outChan <-chan Job, err error) {
	jobsChannel, err := q.conn.Channel()
	if err != nil {
		return
	}

	err = jobsChannel.Qos(1, 0, false)
	if err != nil {
		return
	}

	deliveries, err := jobsChannel.Consume(
		q.queue,              // queue
		"build-job-consumer", // consumer

		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		amqp.Table{"x-priority": int64(q.priority)}) // args

	if err != nil {
		return
	}

	logWriterChannel, err := q.conn.Channel()
	if err != nil {
		return
	}

	buildJobChan := make(chan Job)
	outChan = buildJobChan

	go func() {
		defer jobsChannel.Close()
		defer logWriterChannel.Close()
		defer close(buildJobChan)

		logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
			"self": "amqp_job_queue",
			"inst": fmt.Sprintf("%p", q),
		})

		for {
			if ctx.Err() != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			case delivery, ok := <-deliveries:
				if !ok {
					logger.Info("job queue channel closed")
					return
				}

				buildJob := &amqpJob{
					payload:         &JobPayload{},
					startAttributes: &backend.StartAttributes{},
					stateUpdatePool: q.stateUpdatePool,
					withLogSharding: q.withLogSharding,
					logWriterChan:   logWriterChannel,
					delivery:        delivery,
				}
				startAttrs := &jobPayloadStartAttrs{Config: &backend.StartAttributes{}}

				err := json.Unmarshal(delivery.Body, buildJob.payload)
				if err != nil {
					logger.WithField("err", err).Error("payload JSON parse error")
					continue
				}

				logger.WithField("job_id", buildJob.payload.Job.ID).Info("received amqp delivery")

				err = json.Unmarshal(delivery.Body, &startAttrs)
				if err != nil {
					logger.WithField("err", err).Error("start attributes JSON parse error")

					err = buildJob.Error(ctx, "An error occured while parsing the job config. Please consider enabling the build config validation feature for the repository: https://docs.travis-ci.com/user/build-config-validation")
					if err != nil {
						logger.WithField("err", err).Error("couldn't error the job")
					}

					continue
				}

				buildJob.rawPayload, err = simplejson.NewJson(delivery.Body)
				if err != nil {
					logger.WithField("err", err).Error("raw payload JSON parse error, attempting to ack+drop delivery")
					err := delivery.Ack(false)
					if err != nil {
						logger.WithField("err", err).WithField("delivery", delivery).Error("couldn't ack+drop delivery")
					}
					continue
				}

				buildJob.startAttributes = startAttrs.Config
				buildJob.startAttributes.VMType = buildJob.payload.VMType
				buildJob.startAttributes.VMSize = buildJob.payload.VMSize
				buildJob.startAttributes.VMConfig = buildJob.payload.VMConfig
				buildJob.startAttributes.Warmer = buildJob.payload.Warmer
				buildJob.startAttributes.SetDefaults(q.DefaultLanguage, q.DefaultDist, q.DefaultArch, q.DefaultGroup, q.DefaultOS, VMTypeDefault, VMConfigDefault)
				buildJob.conn = q.conn
				buildJob.stateCount = buildJob.payload.Meta.StateUpdateCount

				jobSendBegin := time.Now()
				select {
				case buildJobChan <- buildJob:
					metrics.TimeSince("travis.worker.job_queue.amqp.blocking_time", jobSendBegin)
					logger.WithFields(logrus.Fields{
						"source":           "amqp",
						"send_duration_ms": time.Since(jobSendBegin).Seconds() * 1e3,
					}).Info("sent job to output channel")
				case <-ctx.Done():
					_ = delivery.Nack(false, true)
					return
				}
			}
		}
	}()

	return
}

// Name returns the name of this queue type, wow!
func (q *AMQPJobQueue) Name() string {
	return "amqp"
}

// Cleanup closes the underlying AMQP connection
func (q *AMQPJobQueue) Cleanup() error {
	q.stateUpdatePool.Close()
	return q.conn.Close()
}
