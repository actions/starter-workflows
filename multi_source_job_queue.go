package worker

import (
	"fmt"
	"strings"
	"time"

	gocontext "context"

	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"
)

type MultiSourceJobQueue struct {
	queues []JobQueue
}

func NewMultiSourceJobQueue(queues ...JobQueue) *MultiSourceJobQueue {
	return &MultiSourceJobQueue{queues: queues}
}

// Jobs returns a Job channel that selects over each source queue Job channel
func (msjq *MultiSourceJobQueue) Jobs(ctx gocontext.Context) (outChan <-chan Job, err error) {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "multi_source_job_queue",
		"inst": fmt.Sprintf("%p", msjq),
	})

	buildJobChan := make(chan Job)
	outChan = buildJobChan

	buildJobChans := map[string]<-chan Job{}

	for i, queue := range msjq.queues {
		jc, err := queue.Jobs(ctx)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err":  err,
				"name": queue.Name(),
			}).Error("failed to get job chan from queue")
			return nil, err
		}
		qName := fmt.Sprintf("%s.%d", queue.Name(), i)
		buildJobChans[qName] = jc
	}

	go func() {
		for {
			for queueName, bjc := range buildJobChans {
				var job Job = nil
				jobSendBegin := time.Now()
				logger = logger.WithField("queue_name", queueName)

				logger.Debug("about to receive job")
				select {
				case job = <-bjc:
					if job == nil {
						logger.Debug("skipping nil job")
						continue
					}
					jobID := uint64(0)
					if job.Payload() != nil {
						jobID = job.Payload().Job.ID
					}

					logger.WithField("job_id", jobID).Debug("about to send job to multi source output channel")
					buildJobChan <- job

					metrics.TimeSince("travis.worker.job_queue.multi.blocking_time", jobSendBegin)
					logger.WithFields(logrus.Fields{
						"job_id":           jobID,
						"source":           queueName,
						"send_duration_ms": time.Since(jobSendBegin).Seconds() * 1e3,
					}).Info("sent job to multi source output channel")
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}
		}
	}()

	return outChan, nil
}

// Name builds a name from each source queue name
func (msjq *MultiSourceJobQueue) Name() string {
	s := []string{}
	for _, queue := range msjq.queues {
		s = append(s, queue.Name())
	}

	return strings.Join(s, ",")
}

// Cleanup runs cleanup for each source queue
func (msjq *MultiSourceJobQueue) Cleanup() error {
	for _, queue := range msjq.queues {
		err := queue.Cleanup()
		if err != nil {
			return err
		}
	}
	return nil
}
