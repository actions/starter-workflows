package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gocontext "context"

	"github.com/bitly/go-simplejson"
	"github.com/travis-ci/worker/backend"
)

func newTestHTTPJob(t *testing.T) *httpJob {
	jobPayload := &JobPayload{
		Type: "job:test",
		Job: JobJobPayload{
			ID:     uint64(123),
			Number: "1",
		},
		Build: BuildPayload{
			ID:     uint64(456),
			Number: "1",
		},
		UUID:   "870f986d-a88f-4801-86cc-3d2dbc6c80da",
		Config: map[string]interface{}{},
		Timeouts: TimeoutsPayload{
			HardLimit:  uint64(9000),
			LogSilence: uint64(8001),
		},
	}
	jsp := jobScriptPayload{
		Name:     "main",
		Encoding: "base64",
		Content:  "IyEvdXNyL2Jpbi9lbnYgYmFzaAplY2hvIHd1dAo=",
	}
	startAttributes := &backend.StartAttributes{
		Language: "go",
		Dist:     "trusty",
	}

	body, err := json.Marshal(jobPayload)
	if err != nil {
		t.Error(err)
	}

	rawPayload, err := simplejson.NewJson(body)
	if err != nil {
		t.Error(err)
	}

	return &httpJob{
		payload: &httpJobPayload{
			Data:      jobPayload,
			JobScript: jsp,
			JWT:       "huh",
			ImageName: "yeap",
		},
		rawPayload:      rawPayload,
		startAttributes: startAttributes,
		deleteSelf:      func(_ gocontext.Context) error { return nil },
	}
}

func TestHTTPJob(t *testing.T) {
	job := newTestHTTPJob(t)

	if job.Payload() == nil {
		t.Fatalf("payload not set")
	}

	if job.RawPayload() == nil {
		t.Fatalf("raw payload not set")
	}

	if job.StartAttributes() == nil {
		t.Fatalf("start attributes not set")
	}

	if job.GoString() == "" {
		t.Fatalf("go string is empty")
	}
}

func TestHTTPJob_GoString(t *testing.T) {
	job := newTestHTTPJob(t)

	str := job.GoString()

	if !strings.HasPrefix(str, "&httpJob{") && !strings.HasSuffix(str, "}") {
		t.Fatalf("go string has unexpected format: %q", str)
	}
}

func TestHTTPJob_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" || r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()
	job := newTestHTTPJob(t)
	job.payload.JobStateURL = ts.URL
	job.payload.JobPartsURL = ts.URL

	err := job.Error(gocontext.TODO(), "wat")
	if err != nil {
		t.Error(err)
	}
}

func TestHTTPJob_Requeue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hai")
	}))
	defer ts.Close()

	job := newTestHTTPJob(t)
	job.payload.JobStateURL = ts.URL
	job.payload.JobPartsURL = ts.URL

	ctx := gocontext.TODO()

	err := job.Requeue(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestHTTPJob_Received(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hai")
	}))
	defer ts.Close()
	job := newTestHTTPJob(t)
	job.payload.JobStateURL = ts.URL
	job.payload.JobPartsURL = ts.URL

	err := job.Received(gocontext.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestHTTPJob_Started(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hai")
	}))
	defer ts.Close()
	job := newTestHTTPJob(t)
	job.payload.JobStateURL = ts.URL
	job.payload.JobPartsURL = ts.URL

	err := job.Started(gocontext.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestHTTPJob_Finish(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
		fmt.Fprintln(w, "hai")
	}))
	defer ts.Close()

	job := newTestHTTPJob(t)
	job.payload.JobStateURL = ts.URL
	job.payload.JobPartsURL = ts.URL

	ctx := gocontext.TODO()

	err := job.Finish(ctx, FinishStatePassed)
	if err != nil {
		t.Error(err)
	}
}
