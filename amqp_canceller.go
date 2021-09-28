package worker

import (
	"encoding/json"
	"fmt"

	gocontext "context"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/travis-ci/worker/context"
)

type cancelCommand struct {
	Type   string `json:"type"`
	JobID  uint64 `json:"job_id"`
	Source string `json:"source"`
	Reason string `json:"reason"`
}

// AMQPCanceller is responsible for listening to a command queue on AMQP and
// dispatching the commands to the right place. Currently the only valid command
// is the 'cancel job' command.
type AMQPCanceller struct {
	conn                    *amqp.Connection
	ctx                     gocontext.Context
	cancellationBroadcaster *CancellationBroadcaster
}

// NewAMQPCanceller creates a new AMQPCanceller. No network traffic
// occurs until you call Run()
func NewAMQPCanceller(ctx gocontext.Context, conn *amqp.Connection, cancellationBroadcaster *CancellationBroadcaster) *AMQPCanceller {
	ctx = context.FromComponent(ctx, "canceller")

	return &AMQPCanceller{
		ctx:  ctx,
		conn: conn,

		cancellationBroadcaster: cancellationBroadcaster,
	}
}

// Run will make the AMQPCanceller listen to the worker command queue and
// start dispatching any incoming commands.
func (d *AMQPCanceller) Run() {
	amqpChan, err := d.conn.Channel()
	logger := context.LoggerFromContext(d.ctx).WithFields(logrus.Fields{
		"self": "amqp_canceller",
		"inst": fmt.Sprintf("%p", d),
	})
	if err != nil {
		logger.WithField("err", err).Error("couldn't open channel")
		return
	}
	defer amqpChan.Close()

	err = amqpChan.Qos(1, 0, false)
	if err != nil {
		logger.WithField("err", err).Error("couldn't set prefetch")
		return
	}

	err = amqpChan.ExchangeDeclare("worker.commands", "fanout", false, false, false, false, nil)
	if err != nil {
		logger.WithField("err", err).Error("couldn't declare exchange")
		return
	}

	queue, err := amqpChan.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logger.WithField("err", err).Error("couldn't declare queue")
		return
	}

	err = amqpChan.QueueBind(queue.Name, "", "worker.commands", false, nil)
	if err != nil {
		logger.WithField("err", err).Error("couldn't bind queue to exchange")
		return
	}

	deliveries, err := amqpChan.Consume(queue.Name, "commands", false, true, false, false, nil)
	if err != nil {
		logger.WithField("err", err).Error("couldn't consume queue")
		return
	}

	for delivery := range deliveries {
		err := d.processCommand(delivery)
		if err != nil {
			logger.WithField("err", err).WithField("delivery", delivery).Error("couldn't process delivery")
		}

		err = delivery.Ack(false)
		if err != nil {
			logger.WithField("err", err).WithField("delivery", delivery).Error("couldn't ack delivery")
		}
	}
}

func (d *AMQPCanceller) processCommand(delivery amqp.Delivery) error {
	command := &cancelCommand{}
	logger := context.LoggerFromContext(d.ctx).WithFields(logrus.Fields{
		"self": "amqp_canceller",
		"inst": fmt.Sprintf("%p", d),
	})
	err := json.Unmarshal(delivery.Body, command)
	if err != nil {
		logger.WithField("err", err).Error("unable to parse JSON")
		return err
	}

	if command.Type != "cancel_job" {
		logger.WithField("command", command.Type).Error("unknown worker command")
		return nil
	}

	d.cancellationBroadcaster.Broadcast(CancellationCommand{JobID: command.JobID, Reason: command.Reason})

	return nil
}

func tryClose(ch chan<- struct{}) (closedNow bool) {
	closedNow = true
	defer func() {
		if x := recover(); x != nil {
			closedNow = false
		}
	}()

	close(ch)

	return
}
