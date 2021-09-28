package worker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gocontext "context"

	"github.com/bitly/go-simplejson"
	"github.com/jtacoma/uritemplates"
	"github.com/pkg/errors"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"
)

type httpJob struct {
	payload         *httpJobPayload
	rawPayload      *simplejson.Json
	startAttributes *backend.StartAttributes
	received        time.Time
	started         time.Time
	finished        time.Time
	finishState     FinishState
	requeued        bool
	stateCount      uint

	refreshClaim func(gocontext.Context)
	deleteSelf   func(gocontext.Context) error
	cancelSelf   func(gocontext.Context)
}

type jobScriptPayload struct {
	Name     string `json:"name"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

type httpJobPayload struct {
	Data        *JobPayload      `json:"data"`
	JobScript   jobScriptPayload `json:"job_script"`
	JobStateURL string           `json:"job_state_url"`
	JobPartsURL string           `json:"log_parts_url"`
	JWT         string           `json:"jwt"`
	ImageName   string           `json:"image_name"`
}

func (j *httpJob) GoString() string {
	return fmt.Sprintf("&httpJob{payload: %#v, startAttributes: %#v}",
		j.payload, j.startAttributes)
}

func (j *httpJob) Payload() *JobPayload {
	return j.payload.Data
}

func (j *httpJob) RawPayload() *simplejson.Json {
	return j.rawPayload
}

func (j *httpJob) StartAttributes() *backend.StartAttributes {
	return j.startAttributes
}

func (j *httpJob) FinishState() FinishState {
	return j.finishState
}

func (j *httpJob) Requeued() bool {
	return j.requeued
}

func (j *httpJob) Error(ctx gocontext.Context, errMessage string) error {
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

func (j *httpJob) Requeue(ctx gocontext.Context) error {
	context.LoggerFromContext(ctx).WithField("self", "http_job").Info("requeueing job")

	metrics.Mark("worker.job.requeue")

	j.requeued = true

	j.received = time.Time{}
	j.started = time.Time{}

	return j.sendStateUpdate(ctx, j.currentState(), "created")
}

func (j *httpJob) Received(ctx gocontext.Context) error {
	j.received = time.Now()
	if j.refreshClaim != nil {
		context.LoggerFromContext(ctx).WithField("self", "http_job").Debug("starting claim refresh goroutine")
		go j.refreshClaim(context.FromJWT(ctx, j.payload.JWT))
	}
	return j.sendStateUpdate(ctx, "queued", "received")
}

func (j *httpJob) Started(ctx gocontext.Context) error {
	j.started = time.Now()

	metrics.TimeSince("travis.worker.job.start_time", j.received)

	return j.sendStateUpdate(ctx, "received", "started")
}

func (j *httpJob) currentState() string {
	currentState := "queued"

	if !j.received.IsZero() {
		currentState = "received"
	}

	if !j.started.IsZero() {
		currentState = "started"
	}

	return currentState
}

func (j *httpJob) Finish(ctx gocontext.Context, state FinishState) error {
	err := j.deleteSelf(ctx)
	if err != nil {
		return err
	}

	j.finished = time.Now()
	if j.received.IsZero() {
		j.received = j.finished
	}

	if j.started.IsZero() {
		j.started = j.finished
	}

	return j.sendStateUpdate(ctx, j.currentState(), string(state))
}

func (j *httpJob) LogWriter(ctx gocontext.Context, defaultLogTimeout time.Duration) (LogWriter, error) {
	logTimeout := time.Duration(j.payload.Data.Timeouts.LogSilence) * time.Second
	if logTimeout == 0 {
		logTimeout = defaultLogTimeout
	}

	return newHTTPLogWriter(ctx, j.payload.JobPartsURL, j.payload.JWT, j.payload.Data.Job.ID, logTimeout)
}

func (j *httpJob) Generate(ctx gocontext.Context, job Job) ([]byte, error) {
	if j.payload.JobScript.Encoding != "base64" {
		return nil, errors.Errorf("unknown job script encoding: %s", j.payload.JobScript.Encoding)
	}

	script, err := base64.StdEncoding.DecodeString(j.payload.JobScript.Content)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't base64 decode job script")
	}

	return script, nil
}

func (j *httpJob) createStateUpdateBody(curState, newState string) map[string]interface{} {
	body := map[string]interface{}{
		"id":    j.Payload().Job.ID,
		"state": newState,
		"cur":   curState,
		"new":   newState,
		"meta": map[string]interface{}{
			"state_update_count": j.stateCount,
		},
	}

	if j.Payload().Job.QueuedAt != nil {
		body["queued_at"] = j.Payload().Job.QueuedAt.UTC().Format(time.RFC3339)
	}
	if !j.received.IsZero() {
		body["received_at"] = j.received.UTC().Format(time.RFC3339)
	}
	if !j.started.IsZero() {
		body["started_at"] = j.started.UTC().Format(time.RFC3339)
	}
	if !j.finished.IsZero() {
		body["finished_at"] = j.finished.UTC().Format(time.RFC3339)
	}

	if j.Payload().Trace {
		body["trace"] = true
	}

	return body
}

func (j *httpJob) sendStateUpdate(ctx gocontext.Context, curState, newState string) error {
	j.stateCount++
	payload := j.createStateUpdateBody(curState, newState)

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error encoding json")
	}

	template, err := uritemplates.Parse(j.payload.JobStateURL)
	if err != nil {
		return errors.Wrap(err, "couldn't parse base URL template")
	}

	u, err := template.Expand(map[string]interface{}{
		"job_id": j.payload.Data.Job.ID,
	})
	if err != nil {
		return errors.Wrap(err, "couldn't expand base URL template")
	}

	req, err := http.NewRequest("PATCH", u, bytes.NewReader(encodedPayload))
	if err != nil {
		return errors.Wrap(err, "couldn't create request")
	}
	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", j.payload.JWT))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error making state update request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		j.handleStateUpdateError(ctx, resp.StatusCode, newState)
		return errors.Errorf("expected %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	return nil
}

func (j *httpJob) handleStateUpdateError(ctx gocontext.Context, status int, newState string) {
	if status != http.StatusConflict {
		return
	}

	if newState == "received" || newState == "started" {
		// NOTE: receiving a conflict response when attempting to transition to
		// 'received' or 'started' means that the job is potentially being run
		// by multiple workers.  Assume the worst and cancel self.
		j.cancelSelf(ctx)
	}
}

func (j *httpJob) SetupContext(ctx gocontext.Context) gocontext.Context {
	return context.FromJWT(ctx, j.payload.JWT)
}

func (j *httpJob) Name() string { return "http" }
