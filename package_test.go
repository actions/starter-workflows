package worker

import (
	"os"
	"time"

	gocontext "context"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
	if os.Getenv("TRAVIS_WORKER_TEST_DEBUG") == "1" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

type fakeJobQueue struct {
	c chan Job

	cleanedUp bool
}

func (jq *fakeJobQueue) Jobs(ctx gocontext.Context) (<-chan Job, error) {
	return (<-chan Job)(jq.c), nil
}

func (jq *fakeJobQueue) Name() string { return "fake" }

func (jq *fakeJobQueue) Cleanup() error {
	jq.cleanedUp = true
	return nil
}

type fakeJob struct {
	payload         *JobPayload
	rawPayload      *simplejson.Json
	startAttributes *backend.StartAttributes
	finishState     FinishState
	requeued        bool

	events []string

	hasBrokenLogWriter bool
}

func (fj *fakeJob) Payload() *JobPayload {
	return fj.payload
}

func (fj *fakeJob) RawPayload() *simplejson.Json {
	return fj.rawPayload
}

func (fj *fakeJob) StartAttributes() *backend.StartAttributes {
	return fj.startAttributes
}

func (fj *fakeJob) FinishState() FinishState {
	return fj.finishState
}

func (fj *fakeJob) Requeued() bool {
	return fj.requeued
}

func (fj *fakeJob) Received(ctx gocontext.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fj.events = append(fj.events, "received")
	return nil
}

func (fj *fakeJob) Started(ctx gocontext.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fj.events = append(fj.events, "started")
	return nil
}

func (fj *fakeJob) Error(ctx gocontext.Context, msg string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fj.events = append(fj.events, "errored")
	return nil
}

func (fj *fakeJob) Requeue(ctx gocontext.Context) error {
	fj.requeued = true
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fj.events = append(fj.events, "requeued")
	return nil
}

func (fj *fakeJob) Finish(ctx gocontext.Context, state FinishState) error {
	fj.finishState = state
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fj.events = append(fj.events, string(state))
	return nil
}

func (fj *fakeJob) LogWriter(_ gocontext.Context, _ time.Duration) (LogWriter, error) {
	return &fakeLogWriter{broken: fj.hasBrokenLogWriter}, nil
}

func (j *fakeJob) SetupContext(ctx gocontext.Context) gocontext.Context { return ctx }

func (fj *fakeJob) Name() string { return "fake" }

type fakeLogWriter struct {
	broken bool
}

func (flw *fakeLogWriter) Write(_ []byte) (int, error) {
	if flw.broken {
		return 0, errors.New("failed to write")
	}
	return 0, nil
}

func (flw *fakeLogWriter) Close() error {
	if flw.broken {
		return errors.New("failed to close")
	}
	return nil
}

func (flw *fakeLogWriter) WriteAndClose(_ []byte) (int, error) {
	if flw.broken {
		return 0, errors.New("failed to write and close")
	}
	return 0, nil
}

func (flw *fakeLogWriter) Timeout() <-chan time.Time {
	return make(chan time.Time)
}

func (flw *fakeLogWriter) SetMaxLogLength(_ int) {}

func (flw *fakeLogWriter) SetJobStarted(meta *JobStartedMeta) {}

func (flw *fakeLogWriter) SetCancelFunc(_ gocontext.CancelFunc) {}

func (flw *fakeLogWriter) MaxLengthReached() bool {
	return false
}
