package worker

import (
	"os"
	"testing"

	"github.com/streadway/amqp"
)

func setupAMQPConn(t *testing.T) (*amqp.Connection, *amqp.Channel) {
	if os.Getenv("AMQP_URI") == "" {
		t.Skip("skipping amqp test since there is no AMQP_URI")
	}

	amqpConn, err := amqp.Dial(os.Getenv("AMQP_URI"))
	if err != nil {
		t.Fatal(err)
	}

	logChan, err := amqpConn.Channel()
	if err != nil {
		t.Fatal(err)
	}

	err = logChan.ExchangeDeclare("reporting", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, nil
	}
	_, err = logChan.QueueDeclare("reporting.jobs.logs", true, false, false, false, nil)
	if err != nil {
		t.Error(err)
	}

	err = logChan.QueueBind("reporting.jobs.logs", "reporting.jobs.logs", "reporting", false, nil)
	if err != nil {
		return nil, nil
	}

	_, err = logChan.QueuePurge("reporting.jobs.logs", false)
	if err != nil {
		t.Error(err)
	}

	return amqpConn, logChan
}
