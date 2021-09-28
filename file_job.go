package worker

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	gocontext "context"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"
)

type fileJob struct {
	createdFile     string
	receivedFile    string
	startedFile     string
	finishedFile    string
	logFile         string
	bytes           []byte
	payload         *JobPayload
	rawPayload      *simplejson.Json
	startAttributes *backend.StartAttributes
	finishState     FinishState
	requeued        bool
}

func (j *fileJob) Payload() *JobPayload {
	return j.payload
}

func (j *fileJob) RawPayload() *simplejson.Json {
	return j.rawPayload
}

func (j *fileJob) StartAttributes() *backend.StartAttributes {
	return j.startAttributes
}

func (j *fileJob) FinishState() FinishState {
	return j.finishState
}

func (j *fileJob) Requeued() bool {
	return j.requeued
}

func (j *fileJob) Received(_ gocontext.Context) error {
	return os.Rename(j.createdFile, j.receivedFile)
}

func (j *fileJob) Started(_ gocontext.Context) error {
	return os.Rename(j.receivedFile, j.startedFile)
}

func (j *fileJob) Error(ctx gocontext.Context, errMessage string) error {
	log, err := j.LogWriter(ctx, time.Minute)
	if err != nil {
		return err
	}

	_, err = log.WriteAndClose([]byte(errMessage))
	if err != nil {
		return err
	}

	return j.Finish(ctx, FinishStateErrored)
}

func (j *fileJob) Requeue(ctx gocontext.Context) error {
	context.LoggerFromContext(ctx).WithField("self", "file_job").Info("requeueing job")

	metrics.Mark("worker.job.requeue")

	j.requeued = true

	var err error

	for _, fname := range []string{
		j.receivedFile,
		j.startedFile,
		j.finishedFile,
	} {
		err = os.Rename(fname, j.createdFile)
		if err == nil {
			return nil
		}
	}

	return err
}

func (j *fileJob) Finish(ctx gocontext.Context, state FinishState) error {
	context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"state": state,
		"self":  "file_job",
	}).Info("finishing job")

	metrics.Mark(fmt.Sprintf("travis.worker.job.finish.%s", state))

	err := os.Rename(j.startedFile, j.finishedFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(strings.Replace(j.finishedFile, ".json", ".state", -1),
		[]byte(state), os.FileMode(0644))
}

func (j *fileJob) LogWriter(ctx gocontext.Context, defaultLogTimeout time.Duration) (LogWriter, error) {
	logTimeout := time.Duration(j.payload.Timeouts.LogSilence) * time.Second
	if logTimeout == 0 {
		logTimeout = defaultLogTimeout
	}

	return newFileLogWriter(ctx, j.logFile, logTimeout)
}

func (j *fileJob) SetupContext(ctx gocontext.Context) gocontext.Context { return ctx }

func (j *fileJob) Name() string { return "file" }
