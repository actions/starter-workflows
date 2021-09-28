package worker

import (
	gocontext "context"
	"time"

	"github.com/streadway/amqp"
)

type AMQPLogWriterFactory struct {
	conn            *amqp.Connection
	withLogSharding bool
	logWriterChan   *amqp.Channel
}

func NewAMQPLogWriterFactory(conn *amqp.Connection, sharded bool) (*AMQPLogWriterFactory, error) {
	channel, err := conn.Channel()
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

	return &AMQPLogWriterFactory{
		conn:            conn,
		withLogSharding: sharded,
		logWriterChan:   channel,
	}, nil
}

func (l *AMQPLogWriterFactory) LogWriter(ctx gocontext.Context, defaultLogTimeout time.Duration, job Job) (LogWriter, error) {
	logTimeout := time.Duration(job.Payload().Timeouts.LogSilence) * time.Second
	if logTimeout == 0 {
		logTimeout = defaultLogTimeout
	}

	return newAMQPLogWriter(ctx, l.logWriterChan, job.Payload().Job.ID, logTimeout, l.withLogSharding)
}

func (l *AMQPLogWriterFactory) Cleanup() error {
	l.logWriterChan.Close()
	return l.conn.Close()
}
