package worker

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gocontext "context"

	"github.com/cenk/backoff"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/context"
)

var (
	httpLogPartSinksByURL      = map[string]*httpLogPartSink{}
	httpLogPartSinksByURLMutex = &sync.Mutex{}
)

const (
	// defaultHTTPLogPartSinkMaxBufferSize of 150 should be roughly 250MB of
	// buffer on the high end of possible LogChunkSize, which is somewhat
	// arbitrary, but is an amount of memory per worker process that we can
	// tolerate on all hosted infrastructures and should allow for enough wiggle
	// room that we don't hit log sink buffer backpressure unless something is
	// seriously broken with log parts publishing. ~@meatballhat
	defaultHTTPLogPartSinkMaxBufferSize = 150
)

type httpLogPartEncodedPayload struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Final    bool   `json:"final"`
	JobID    uint64 `json:"job_id"`
	Number   uint64 `json:"number"`
	Token    string `json:"tok"`
	Type     string `json:"@type"`
}

type httpLogPartSink struct {
	httpClient *http.Client
	baseURL    string

	partsBuffer      []*httpLogPart
	partsBufferMutex *sync.Mutex
	flushChan        chan struct{}

	maxBufferSize uint64
}

func getHTTPLogPartSinkByURL(url string) *httpLogPartSink {
	httpLogPartSinksByURLMutex.Lock()
	defer httpLogPartSinksByURLMutex.Unlock()

	var (
		lps *httpLogPartSink
		ok  bool
	)

	if lps, ok = httpLogPartSinksByURL[url]; !ok {
		lps = newHTTPLogPartSink(context.FromComponent(rootContext, "log_part_sink"),
			url, defaultHTTPLogPartSinkMaxBufferSize)
		httpLogPartSinksByURL[url] = lps
	}

	return lps
}

func newHTTPLogPartSink(ctx gocontext.Context, url string, maxBufferSize uint64) *httpLogPartSink {
	lps := &httpLogPartSink{
		httpClient:       &http.Client{},
		baseURL:          url,
		partsBuffer:      []*httpLogPart{},
		partsBufferMutex: &sync.Mutex{},
		flushChan:        make(chan struct{}),
		maxBufferSize:    maxBufferSize,
	}

	go lps.flushRegularly(ctx)

	return lps
}

func (lps *httpLogPartSink) Add(ctx gocontext.Context, part *httpLogPart) error {
	logger := context.LoggerFromContext(ctx).WithField("self", "http_log_part_sink")

	lps.partsBufferMutex.Lock()
	bufLen := uint64(len(lps.partsBuffer))
	lps.partsBufferMutex.Unlock()

	if bufLen >= lps.maxBufferSize {
		return fmt.Errorf("log sink buffer has reached max size %d", lps.maxBufferSize)
	} else if (bufLen + (lps.maxBufferSize / uint64(10))) >= lps.maxBufferSize {
		logger.WithField("size", bufLen).Debug("triggering flush because of large buffer size")
		lps.flushChan <- struct{}{}
	}

	lps.partsBufferMutex.Lock()
	lps.partsBuffer = append(lps.partsBuffer, part)
	lps.partsBufferMutex.Unlock()

	logger.WithField("size", bufLen+1).Debug("appended to log parts buffer")

	return nil
}

func (lps *httpLogPartSink) flushRegularly(ctx gocontext.Context) {
	logger := context.LoggerFromContext(ctx).WithField("self", "http_log_part_sink")
	ticker := time.NewTicker(LogWriterTick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logger.Debug("ticker flushing")
			lps.flush(ctx)
		case <-lps.flushChan:
			logger.Debug("event-based flushing")
			lps.flush(ctx)
		case <-ctx.Done():
			logger.Debug("exiting from context done")
			return
		}
	}
}

func (lps *httpLogPartSink) flush(ctx gocontext.Context) error {
	logger := context.LoggerFromContext(ctx).WithField("self", "http_log_part_sink")

	lps.partsBufferMutex.Lock()
	bufLen := len(lps.partsBuffer)
	if bufLen == 0 {
		logger.WithField("size", bufLen).Debug("not flushing empty log parts buffer")
		lps.partsBufferMutex.Unlock()
		return nil
	}

	logger.WithField("size", bufLen).Debug("flushing log parts buffer")
	bufferSample := make([]*httpLogPart, bufLen)
	copy(bufferSample, lps.partsBuffer)
	lps.partsBuffer = []*httpLogPart{}

	lps.partsBufferMutex.Unlock()

	payload := []*httpLogPartEncodedPayload{}

	for _, part := range bufferSample {
		logger.WithFields(logrus.Fields{
			"job_id": part.JobID,
			"number": part.Number,
		}).Debug("appending encoded log part to payload")

		payload = append(payload, &httpLogPartEncodedPayload{
			Content:  base64.StdEncoding.EncodeToString([]byte(part.Content)),
			Encoding: "base64",
			Final:    part.Final,
			JobID:    part.JobID,
			Number:   part.Number,
			Token:    part.Token,
			Type:     "log_part",
		})
	}

	err := lps.publishLogParts(ctx, payload)
	if err != nil {
		// NOTE: This is the point of origin for log parts backpressure, in
		// combination with the error returned by `.Add` when maxBufferSize is
		// reached.  Because running jobs will not be able to send their log parts
		// anywhere, it remains to be determined whether we should cancel (and
		// reset) running jobs or allow them to complete without capturing output.
		for _, part := range bufferSample {
			addErr := lps.Add(ctx, part)
			if addErr != nil {
				logger.WithField("err", addErr).Error("failed to re-add buffer sample log part")
			}
		}
		logger.WithField("err", err).Error("failed to publish buffered parts")
		return err
	}
	logger.Debug("successfully published buffered parts")
	return nil
}

func (lps *httpLogPartSink) publishLogParts(ctx gocontext.Context, payload []*httpLogPartEncodedPayload) error {
	publishURL, err := url.Parse(lps.baseURL)
	if err != nil {
		return errors.Wrap(err, "couldn't parse base URL")
	}

	payloadBody, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "couldn't marshal JSON")
	}

	query := publishURL.Query()
	query.Set("source", "worker")
	publishURL.RawQuery = query.Encode()

	httpBackOff := backoff.NewExponentialBackOff()
	// TODO: make this configurable?
	httpBackOff.MaxInterval = 10 * time.Second
	httpBackOff.MaxElapsedTime = 3 * time.Minute

	logger := context.LoggerFromContext(ctx).WithField("self", "http_log_part_sink")

	var resp *http.Response
	err = backoff.Retry(func() (err error) {
		var req *http.Request
		req, err = http.NewRequest("POST", publishURL.String(), bytes.NewReader(payloadBody))
		if err != nil {
			return
		}

		req.Header.Set("Authorization", fmt.Sprintf("token sig:%s", lps.generatePayloadSignature(payload)))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		logger.WithField("req", req).Debug("attempting to publish log parts")
		resp, err = lps.httpClient.Do(req)
		if resp != nil && resp.StatusCode != http.StatusNoContent {
			logger.WithFields(logrus.Fields{
				"expected_status": http.StatusNoContent,
				"actual_status":   resp.StatusCode,
			}).Debug("publish failed")

			if resp.Body != nil {
				resp.Body.Close()
			}

			return errors.Errorf("expected %d but got %d", http.StatusNoContent, resp.StatusCode)
		}
		return
	}, httpBackOff)

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return errors.Wrap(err, "failed to send log parts with retries")
	}

	return nil
}

func (lps *httpLogPartSink) generatePayloadSignature(payload []*httpLogPartEncodedPayload) string {
	authTokens := []string{}
	for _, logPart := range payload {
		authTokens = append(authTokens, logPart.Token)
	}

	sig := sha1.Sum([]byte(strings.Join(authTokens, "")))
	return hex.EncodeToString(sig[:])
}
