package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	gocontext "context"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/travis-ci/worker/context"
)

type amqpLogPart struct {
	JobID   uint64          `json:"id"`
	Content string          `json:"log"`
	Number  int             `json:"number"`
	UUID    string          `json:"uuid"`
	Final   bool            `json:"final"`
	Meta    *JobStartedMeta `json:"meta,omitempty"`
}

type amqpLogWriter struct {
	ctx     gocontext.Context
	cancel  gocontext.CancelFunc
	jobID   uint64
	sharded bool

	closeChan chan struct{}

	bufferMutex      sync.Mutex
	buffer           *bytes.Buffer
	logPartNumber    int
	jobStarted       bool
	jobStartedMeta   *JobStartedMeta
	maxLengthReached bool

	bytesWritten int
	maxLength    int

	amqpChanMutex sync.RWMutex
	amqpChan      *amqp.Channel

	timer   *time.Timer
	timeout time.Duration
}

func newAMQPLogWriter(ctx gocontext.Context, logWriterChan *amqp.Channel, jobID uint64, timeout time.Duration, sharded bool) (*amqpLogWriter, error) {
	writer := &amqpLogWriter{
		ctx:       context.FromComponent(ctx, "log_writer"),
		amqpChan:  logWriterChan,
		jobID:     jobID,
		closeChan: make(chan struct{}),
		buffer:    new(bytes.Buffer),
		timer:     time.NewTimer(time.Hour),
		timeout:   timeout,
		sharded:   sharded,
	}

	context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"writer": writer,
		"job_id": jobID,
	}).Debug("created new log writer")

	go writer.flushRegularly(ctx)

	return writer, nil
}

func (w *amqpLogWriter) Write(p []byte) (int, error) {
	if w.closed() {
		return 0, fmt.Errorf("attempted write to closed log")
	}

	logger := context.LoggerFromContext(w.ctx).WithFields(logrus.Fields{
		"self": "amqp_log_writer",
		"inst": fmt.Sprintf("%p", w),
	})

	logger.WithFields(logrus.Fields{
		"length": len(p),
		"bytes":  string(p),
	}).Debug("writing bytes")

	w.timer.Reset(w.timeout)

	w.bytesWritten += len(p)
	if w.bytesWritten > w.maxLength {
		logger.Info("wrote past maximum log length - cancelling context")
		w.maxLengthReached = true
		if w.cancel == nil {
			logger.Error("cancel function does not exist")
		} else {
			w.cancel()
		}
		return 0, nil
	}

	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()
	return w.buffer.Write(p)
}

func (w *amqpLogWriter) Close() error {
	if w.closed() {
		return nil
	}

	w.timer.Stop()

	close(w.closeChan)
	w.flush()

	part := amqpLogPart{
		JobID:  w.jobID,
		Number: w.logPartNumber,
		Final:  true,
	}
	w.logPartNumber++

	err := w.publishLogPart(part)
	return err
}

func (w *amqpLogWriter) Timeout() <-chan time.Time {
	return w.timer.C
}

func (w *amqpLogWriter) SetMaxLogLength(bytes int) {
	w.maxLength = bytes
}

func (w *amqpLogWriter) SetJobStarted(meta *JobStartedMeta) {
	w.jobStarted = true
	w.jobStartedMeta = meta
}

func (w *amqpLogWriter) SetCancelFunc(cancel gocontext.CancelFunc) {
	w.cancel = cancel
}

func (w *amqpLogWriter) MaxLengthReached() bool {
	return w.maxLengthReached
}

// WriteAndClose works like a Write followed by a Close, but ensures that no
// other Writes are allowed in between.
func (w *amqpLogWriter) WriteAndClose(p []byte) (int, error) {
	if w.closed() {
		return 0, fmt.Errorf("log already closed")
	}

	w.timer.Stop()

	close(w.closeChan)

	w.bufferMutex.Lock()
	n, err := w.buffer.Write(p)
	w.bufferMutex.Unlock()
	if err != nil {
		return n, err
	}

	w.flush()

	part := amqpLogPart{
		JobID:  w.jobID,
		Number: w.logPartNumber,
		Final:  true,
	}
	w.logPartNumber++

	err = w.publishLogPart(part)
	return n, err
}

func (w *amqpLogWriter) closed() bool {
	select {
	case <-w.closeChan:
		return true
	default:
		return false
	}
}

func (w *amqpLogWriter) flushRegularly(ctx gocontext.Context) {
	ticker := time.NewTicker(LogWriterTick)
	defer ticker.Stop()
	for {
		select {
		case <-w.closeChan:
			return
		case <-ticker.C:
			w.flush()
		case <-ctx.Done():
			return
		}
	}
}

func (w *amqpLogWriter) flush() {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()
	if w.buffer.Len() <= 0 {
		return
	}

	buf := make([]byte, LogChunkSize)
	logger := context.LoggerFromContext(w.ctx).WithFields(logrus.Fields{
		"self": "amqp_log_writer",
		"inst": fmt.Sprintf("%p", w),
	})

	for w.buffer.Len() > 0 {
		n, err := w.buffer.Read(buf)
		if err != nil {
			// According to documentation, err should only be non-nil if
			// there's no data in the buffer. We've checked for this, so
			// this means that err should never be nil. Something is very
			// wrong if this happens, so let's abort!
			panic("non-empty buffer shouldn't return an error on Read")
		}

		part := amqpLogPart{
			JobID:   w.jobID,
			Content: string(buf[0:n]),
			Number:  w.logPartNumber,
		}
		w.logPartNumber++

		err = w.publishLogPart(part)
		if err != nil {
			logger.WithField("err", err).Error("couldn't publish log part")
		}
	}
}

func (w *amqpLogWriter) publishLogPart(part amqpLogPart) error {
	part.UUID, _ = context.UUIDFromContext(w.ctx)

	// we emit the queued_at field on the log part to indicate that
	// this is when the job started running. downstream consumers of
	// the log parts (travis-logs) can then use the timestamp to compute
	// a "time to first log line" metric.
	if w.jobStarted {
		part.Meta = w.jobStartedMeta
		w.jobStarted = false
	}

	partBody, err := json.Marshal(part)
	if err != nil {
		return err
	}

	w.amqpChanMutex.RLock()
	var exchange string
	var routingKey string
	if w.sharded {
		exchange = "reporting.jobs.logs_sharded"
		routingKey = strconv.FormatUint(w.jobID, 10)
	} else {
		exchange = "reporting"
		routingKey = "reporting.jobs.logs"
	}
	err = w.amqpChan.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Type:         "job:test:log",
		Body:         partBody,
	})
	w.amqpChanMutex.RUnlock()

	return err
}
