package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pborman/uuid"
	workerctx "github.com/travis-ci/worker/context"
)

func TestAMQPLogWriterWrite(t *testing.T) {
	amqpConn, amqpChan := setupAMQPConn(t)
	defer amqpConn.Close()
	defer amqpChan.Close()

	uuid := uuid.NewRandom()
	ctx := workerctx.FromUUID(context.TODO(), uuid.String())

	logWriter, err := newAMQPLogWriter(ctx, amqpChan, 4, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	logWriter.SetMaxLogLength(1000)

	_, err = fmt.Fprintf(logWriter, "Hello, ")
	if err != nil {
		t.Error(err)
	}
	_, err = fmt.Fprintf(logWriter, "world!")
	if err != nil {
		t.Error(err)
	}

	// Close the log writer to force it to flush out the buffer
	err = logWriter.Close()
	if err != nil {
		t.Error(err)
	}

	delivery, ok, err := amqpChan.Get("reporting.jobs.logs", true)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("expected log message, but there was none")
	}

	var lp amqpLogPart

	err = json.Unmarshal(delivery.Body, &lp)
	if err != nil {
		t.Error(err)
	}

	expected := amqpLogPart{
		JobID:   4,
		Content: "Hello, world!",
		Number:  0,
		UUID:    uuid.String(),
		Final:   false,
	}

	if expected != lp {
		t.Errorf("log part is %#v, expected %#v", lp, expected)
	}
}

func TestAMQPLogWriterClose(t *testing.T) {
	amqpConn, amqpChan := setupAMQPConn(t)
	defer amqpConn.Close()
	defer amqpChan.Close()

	uuid := uuid.NewRandom()
	ctx := workerctx.FromUUID(context.TODO(), uuid.String())

	logWriter, err := newAMQPLogWriter(ctx, amqpChan, 4, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	logWriter.SetMaxLogLength(1000)

	// Close the log writer to force it to flush out the buffer
	err = logWriter.Close()
	if err != nil {
		t.Error(err)
	}

	delivery, ok, err := amqpChan.Get("reporting.jobs.logs", true)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("expected log message, but there was none")
	}

	var lp amqpLogPart

	err = json.Unmarshal(delivery.Body, &lp)
	if err != nil {
		t.Error(err)
	}

	expected := amqpLogPart{
		JobID:   4,
		Content: "",
		Number:  0,
		UUID:    uuid.String(),
		Final:   true,
	}

	if expected != lp {
		t.Errorf("log part is %#v, expected %#v", lp, expected)
	}
}

func noCancel() {}

func TestAMQPMaxLogLength(t *testing.T) {
	amqpConn, amqpChan := setupAMQPConn(t)
	defer amqpConn.Close()
	defer amqpChan.Close()

	uuid := uuid.NewRandom()
	ctx := workerctx.FromUUID(context.TODO(), uuid.String())

	logWriter, err := newAMQPLogWriter(ctx, amqpChan, 4, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	logWriter.SetMaxLogLength(4)
	logWriter.SetCancelFunc(noCancel)

	_, err = fmt.Fprintf(logWriter, "1234")
	if err != nil {
		t.Error(err)
	}
	if logWriter.MaxLengthReached() {
		t.Error("max length should not be reached yet")
	}

	_, _ = fmt.Fprintf(logWriter, "5")
	if !logWriter.MaxLengthReached() {
		t.Error("expected MaxLengthReached to be true")
	}
}
