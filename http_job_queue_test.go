package worker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	gocontext "context"

	"github.com/stretchr/testify/assert"
)

func TestHTTPJobQueue(t *testing.T) {
	hjq, err := NewHTTPJobQueue(nil, "test", "fake", "fake", nil)
	assert.Nil(t, err)
	assert.NotNil(t, hjq)
}

func TestHTTPJobQueue_Jobs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(`/jobs/pop`, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"job_id":"100001"}`)
	})
	mux.HandleFunc(`/jobs/100001/claim`, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc(`/jobs/100001`, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, strings.Replace(`{
			"data": {
				"type": "job",
				"job": {
					"id": 100001,
					"number": "42.1",
					"queued_at": "2011-04-01T11:05:55Z"
				},
				"source": {
					"id": 100001,
					"number": "42"
				},
				"repository": {
					"id": 8490324,
					"slug": "travis-ci/nonexistent-repository"
				},
				"uuid": "fafafaf",
				"config": {},
				"vm_type": "test",
				"meta": {
					"state_update_count": 0
				}
			}
		}`, "\t", "  ", -1))
	})

	mux.HandleFunc(`/`, func(w http.ResponseWriter, req *http.Request) {
		t.Fatalf("unknown URL requested: %#v", req.URL.Path)
	})
	jobBoardServer := httptest.NewServer(mux)
	defer jobBoardServer.Close()

	jobBoardURL, _ := url.Parse(jobBoardServer.URL)
	hjq, err := NewHTTPJobQueue(jobBoardURL, "test", "fake", "fake", nil)
	assert.Nil(t, err)
	assert.NotNil(t, hjq)

	ctx := gocontext.TODO()
	buildJobChan, err := hjq.Jobs(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, buildJobChan)

	select {
	case job := <-buildJobChan:
		assert.NotNil(t, job)
	case <-time.After(time.Second):
		t.Fatalf("failed to recv job")
	}
}

func TestHTTPJobQueue_Name(t *testing.T) {
	hjq, err := NewHTTPJobQueue(nil, "test", "fake", "fake", nil)
	assert.Nil(t, err)
	assert.Equal(t, "http", hjq.Name())
}

func TestHTTPJobQueue_Cleanup(t *testing.T) {
	hjq, err := NewHTTPJobQueue(nil, "test", "fake", "fake", nil)
	assert.Nil(t, err)
	assert.Nil(t, hjq.Cleanup())
}
