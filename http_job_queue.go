package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/cenk/backoff"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"

	gocontext "context"
)

var (
	httpJobQueueNoJobsErr  = fmt.Errorf("no jobs available")
	httpJobRefreshClaimErr = fmt.Errorf("failed to refresh claim")
)

// HTTPJobQueue is a JobQueue that uses http
type HTTPJobQueue struct {
	jobBoardURL          *url.URL
	site                 string
	providerName         string
	queue                string
	pollInterval         time.Duration
	refreshClaimInterval time.Duration
	cb                   *CancellationBroadcaster

	DefaultLanguage, DefaultDist, DefaultArch, DefaultGroup, DefaultOS string
}

type jobBoardErrorResponse struct {
	Type          string `json:"@type"`
	Error         string `json:"error"`
	UpstreamError string `json:"upstream_error,omitempty"`
}

// NewHTTPJobQueue creates a new http job queue
func NewHTTPJobQueue(jobBoardURL *url.URL, site, providerName, queue string,
	cb *CancellationBroadcaster) (*HTTPJobQueue, error) {

	return &HTTPJobQueue{
		jobBoardURL:          jobBoardURL,
		site:                 site,
		providerName:         providerName,
		queue:                queue,
		pollInterval:         3 * time.Second,
		refreshClaimInterval: 5 * time.Second,
		cb:                   cb,
	}, nil
}

// NewHTTPJobQueueWithIntervals creates a new http job queue with the specified
// poll and refresh claim intervals
func NewHTTPJobQueueWithIntervals(jobBoardURL *url.URL, site, providerName, queue string,
	pollInterval, refreshClaimInterval time.Duration,
	cb *CancellationBroadcaster) (*HTTPJobQueue, error) {

	return &HTTPJobQueue{
		jobBoardURL:          jobBoardURL,
		site:                 site,
		providerName:         providerName,
		queue:                queue,
		pollInterval:         pollInterval,
		refreshClaimInterval: refreshClaimInterval,
		cb:                   cb,
	}, nil
}

// Jobs consumes new jobs from job-board
func (q *HTTPJobQueue) Jobs(ctx gocontext.Context) (outChan <-chan Job, err error) {
	buildJobChan := make(chan Job)
	outChan = buildJobChan
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "http_job_queue",
		"inst": fmt.Sprintf("%p", q),
	})

	go func() {
		defer close(buildJobChan)

		for {
			logger.Debug("polling for job tick")
			pollInterval, keepPolling, readyChan := q.pollForJob(ctx, buildJobChan)
			if readyChan != nil {
				readyWaitBegin := time.Now()
				logger.Debug("blocking on ready channel recv")
				<-readyChan
				metrics.TimeSince("travis.worker.job_queue.http.ready_wait_time", readyWaitBegin)
			}
			if !keepPolling {
				return
			}
			time.Sleep(pollInterval)
		}
	}()

	return outChan, nil
}

// pollForJob is responsible for first fetching a job ID, if available, and then
// fetching the complete job representation and sending it into the
// `buildJobChan` that is passed in from the `Jobs` method.  The *httpJob that
// is constructed and sent into the `buildJobChan` is assigned a `refreshClaim`
// func that has a reference to a "ready" `chan struct{}` used to indicate when
// the polling loop may resume.
func (q *HTTPJobQueue) pollForJob(ctx gocontext.Context, buildJobChan chan Job) (time.Duration, bool, <-chan struct{}) {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "http_job_queue",
		"inst": fmt.Sprintf("%p", q),
	})

	logger.Debug("fetching job id")
	pollInterval, jobID, err := q.fetchJobID(ctx)
	if err != nil {
		logger.WithField("err", err).Debug("continuing after failing to get job id")
		return pollInterval, true, nil
	}
	logger.WithField("job_id", jobID).Debug("fetching complete job")
	buildJob, readyChan, err := q.fetchJob(ctx, jobID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
			"id":  jobID,
		}).Warn("failed to get complete job")
		return pollInterval, true, nil
	}

	logger.WithField("job_id", jobID).Debug("sending job to output channel")
	jobSendBegin := time.Now()
	select {
	case buildJobChan <- buildJob:
		metrics.TimeSince("travis.worker.job_queue.http.blocking_time", jobSendBegin)
		logger.WithFields(logrus.Fields{
			"source":           "http",
			"send_duration_ms": time.Since(jobSendBegin).Seconds() * 1e3,
		}).Info("sent job to output channel")
		return pollInterval, true, readyChan
	case <-ctx.Done():
		if j, ok := buildJob.(*httpJob); ok {
			if processorID, ok := context.ProcessorFromContext(ctx); ok {
				// best-effort delete
				delCtx := context.FromProcessor(
					context.FromJWT(gocontext.TODO(), j.payload.JWT),
					processorID)
				logger.WithField("job_id", jobID).Warn("context done; deleting job")
				_ = q.deleteJob(delCtx, jobID)
			}
		}
		logger.WithField("err", ctx.Err()).Warn("returning from jobs loop due to context done")
		return pollInterval, false, nil
	}
}

func (q *HTTPJobQueue) fetchJobID(ctx gocontext.Context) (time.Duration, uint64, error) {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "http_job_queue",
		"inst": fmt.Sprintf("%p", q),
	})

	processorID, ok := context.ProcessorFromContext(ctx)
	if !ok {
		processorID = "unknown-processor"
	}

	u := *q.jobBoardURL

	query := u.Query()
	query.Add("queue", q.queue)

	u.Path = "/jobs/pop"
	u.RawQuery = query.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return q.pollInterval, 0, errors.Wrap(err, "failed to create job-board job pop request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Travis-Site", q.site)
	req.Header.Add("From", processorID)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return q.pollInterval, 0, errors.Wrap(err, "failed to make job-board job pop request")
	}

	defer resp.Body.Close()

	pollInterval := q.pollInterval
	if v, err := strconv.ParseUint(resp.Header.Get("Travis-Pop-Interval"), 10, 64); err == nil {
		pollInterval = time.Duration(v) * time.Second
	}

	if resp.StatusCode == http.StatusNoContent {
		return pollInterval, 0, httpJobQueueNoJobsErr
	}

	fetchResponsePayload := map[string]string{"job_id": ""}
	err = json.NewDecoder(resp.Body).Decode(&fetchResponsePayload)
	if err != nil {
		return pollInterval, 0, errors.Wrap(err, "failed to decode job-board job pop response")
	}

	fetchedJobID, err := strconv.ParseUint(fetchResponsePayload["job_id"], 10, 64)
	if err != nil {
		return pollInterval, 0, errors.Wrap(err, "failed to parse job ID")
	}

	logger.WithField("job_id", fetchedJobID).Debug("fetched")
	return pollInterval, fetchedJobID, nil
}

func (q *HTTPJobQueue) deleteJob(ctx gocontext.Context, jobID uint64) error {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "http_job_queue",
	})

	logger.Info("deleting job")

	jwt, ok := context.JWTFromContext(ctx)
	if !ok {
		return errors.New("failed to delete job; no jwt in context")
	}

	processorID, ok := context.ProcessorFromContext(ctx)
	if !ok {
		processorID = "unknown-processor"
	}

	u := *q.jobBoardURL
	u.Path = fmt.Sprintf("/jobs/%d", jobID)
	u.User = nil

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Travis-Site", q.site)
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("From", processorID)

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = 1 * time.Minute

	logger.WithField("url", u.String()).Debug("performing DELETE request")

	var resp *http.Response
	err = backoff.Retry(func() (err error) {
		resp, err = http.DefaultClient.Do(req)
		if resp != nil && resp.StatusCode != http.StatusNoContent {
			logger.WithFields(logrus.Fields{
				"expected_status": http.StatusNoContent,
				"actual_status":   resp.StatusCode,
			}).Debug("delete failed")

			if resp.Body != nil {
				resp.Body.Close()
			}
			return errors.Errorf("expected %d but got %d", http.StatusNoContent, resp.StatusCode)
		}

		return
	}, bo)

	if err != nil {
		return errors.Wrap(err, "failed to delete job with retries")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var errorResp jobBoardErrorResponse
	err = json.Unmarshal(body, &errorResp)
	if err != nil {
		return errors.Wrapf(err, "job board job delete request errored with status %d and didn't send an error response", resp.StatusCode)
	}

	return errors.Errorf("job board job delete request errored with status %d: %s", resp.StatusCode, errorResp.Error)
}

func (q *HTTPJobQueue) refreshJobClaim(ctx gocontext.Context, jobID uint64) (time.Duration, error) {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self":   "http_job_queue",
		"job_id": jobID,
		"inst":   fmt.Sprintf("%p", q),
	})

	jwt, ok := context.JWTFromContext(ctx)
	if !ok {
		return q.refreshClaimInterval, errors.New("failed to refresh claim; no jwt in context")
	}

	processorID, ok := context.ProcessorFromContext(ctx)
	if !ok {
		processorID = "unknown-processor"
	}

	u := *q.jobBoardURL
	u.User = nil

	query := u.Query()
	query.Add("queue", q.queue)

	u.Path = fmt.Sprintf("/jobs/%v/claim", jobID)
	u.RawQuery = query.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return q.refreshClaimInterval, errors.Wrap(err, "failed to create job-board job claim request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Travis-Site", q.site)
	req.Header.Add("From", processorID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return q.refreshClaimInterval, errors.Wrap(err, "failed to make job-board job claim request")
	}

	resp.Body.Close()

	refreshClaimInterval := q.refreshClaimInterval
	if v, err := strconv.ParseUint(resp.Header.Get("Travis-Refresh-Claim-Interval"), 10, 64); err == nil {
		refreshClaimInterval = time.Duration(v) * time.Second
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithField("response_code", resp.StatusCode).Debug("non-200 response code")
		return refreshClaimInterval, httpJobRefreshClaimErr
	}

	logger.Debug("refreshed claim")
	return refreshClaimInterval, nil
}

func (q *HTTPJobQueue) fetchJob(ctx gocontext.Context, jobID uint64) (Job, <-chan struct{}, error) {
	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"self": "http_job_queue",
		"inst": fmt.Sprintf("%p", q),
	})

	processorID, ok := context.ProcessorFromContext(ctx)
	if !ok {
		processorID = "unknown-processor"
	}

	refreshClaimFunc, readyChan := q.generateJobRefreshClaimFunc(jobID)

	buildJob := &httpJob{
		payload: &httpJobPayload{
			Data: &JobPayload{},
		},
		startAttributes: &backend.StartAttributes{},

		refreshClaim: refreshClaimFunc,
		deleteSelf: func(ctx gocontext.Context) error {
			return q.deleteJob(ctx, jobID)
		},
		cancelSelf: func(ctx gocontext.Context) {
			q.cb.Broadcast(CancellationCommand{JobID: jobID})
		},
	}
	startAttrs := &httpJobPayloadStartAttrs{
		Data: &jobPayloadStartAttrs{
			Config: &backend.StartAttributes{},
		},
	}

	u := *q.jobBoardURL
	u.Path = fmt.Sprintf("/jobs/%d", jobID)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't make job-board job request")
	}

	// TODO: ensure infrastructure is not synonymous with providerName since
	// there's the possibility that a provider has multiple infrastructures, which
	// is expected to be the case with the future cloudbrain provider.
	req.Header.Add("Travis-Infrastructure", q.providerName)
	req.Header.Add("Travis-Site", q.site)
	req.Header.Add("From", processorID)
	req = req.WithContext(ctx)

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = 1 * time.Minute

	var resp *http.Response
	err = backoff.Retry(func() (err error) {
		resp, err = (&http.Client{}).Do(req)
		if resp != nil && resp.StatusCode != http.StatusOK {
			logger.WithFields(logrus.Fields{
				"expected_status": http.StatusOK,
				"actual_status":   resp.StatusCode,
			}).Debug("job fetch failed")

			if resp.Body != nil {
				resp.Body.Close()
			}

			return errors.Errorf("expected %d but got %d", http.StatusOK, resp.StatusCode)
		}
		return
	}, bo)

	if err != nil {
		return nil, nil, errors.Wrap(err, "error making job-board job request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error reading body from job-board job request")
	}

	err = json.Unmarshal(body, buildJob.payload)
	if err != nil {
		logger.WithField("err", err).Error("payload JSON parse error, attempting to delete job")
		err := q.deleteJob(ctx, jobID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't delete job")
		}
		return nil, nil, errors.Wrap(err, "payload JSON parse error")
	}

	err = json.Unmarshal(body, &startAttrs)
	if err != nil {
		logger.WithField("err", err).Error("start attributes JSON parse error, attempting to delete job")
		err := q.deleteJob(ctx, jobID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't delete job")
		}
		return nil, nil, errors.Wrap(err, "start attributes JSON parse error")
	}

	rawPayload, err := simplejson.NewJson(body)
	if err != nil {
		logger.WithField("err", err).Error("raw payload JSON parse error, attempting to delete job")
		err := q.deleteJob(ctx, jobID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't delete job")
		}
		return nil, nil, errors.Wrap(err, "raw payload JSON parse error")
	}
	buildJob.rawPayload = rawPayload.Get("data")

	buildJob.startAttributes = startAttrs.Data.Config
	buildJob.startAttributes.VMConfig = buildJob.payload.Data.VMConfig
	buildJob.startAttributes.VMType = buildJob.payload.Data.VMType
	buildJob.startAttributes.SetDefaults(q.DefaultLanguage, q.DefaultDist, q.DefaultArch, q.DefaultGroup, q.DefaultOS, VMTypeDefault, VMConfigDefault)

	return buildJob, readyChan, nil
}

func (q *HTTPJobQueue) generateJobRefreshClaimFunc(jobID uint64) (func(gocontext.Context), <-chan struct{}) {
	readyChan := make(chan struct{})

	return func(ctx gocontext.Context) {
		defer func() { close(readyChan) }()

		for {
			refreshClaimInterval, err := q.refreshJobClaim(ctx, jobID)
			if err == httpJobRefreshClaimErr && ctx.Err() == nil {
				// NOTE: indicates an error while context is not yet done
				context.LoggerFromContext(ctx).WithFields(logrus.Fields{
					"err":    err,
					"job_id": jobID,
				}).Error("cancelling")
				q.cb.Broadcast(CancellationCommand{JobID: jobID})
				return
			}

			if err != nil && ctx.Err() == nil {
				context.LoggerFromContext(ctx).WithFields(logrus.Fields{
					"err":    err,
					"job_id": jobID,
				}).Error("failed to refresh claim; continuing")
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(refreshClaimInterval):
			}
		}
	}, (<-chan struct{})(readyChan)
}

// Name returns the name of this queue type, wow!
func (q *HTTPJobQueue) Name() string {
	return "http"
}

// Cleanup does not do anything!
func (q *HTTPJobQueue) Cleanup() error {
	return nil
}
