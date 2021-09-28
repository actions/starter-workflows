package worker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gocontext "context"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/context"
)

func TestNewCLI(t *testing.T) {
	c := NewCLI(nil)
	assert.NotNil(t, c)
	assert.Nil(t, c.c)
	assert.True(t, c.bootTime.String() < time.Now().UTC().String())
}

func TestCLI_heartbeatHandler(t *testing.T) {
	i := NewCLI(nil)
	i.heartbeatSleep = time.Duration(0)
	i.heartbeatErrSleep = time.Duration(0)

	ctx, cancel := gocontext.WithCancel(gocontext.Background())
	logger := context.LoggerFromContext(ctx).WithField("self", "cli_test")
	i.ctx = ctx
	i.cancel = cancel
	i.logger = logger

	logrus.SetLevel(logrus.FatalLevel)

	i.ProcessorPool = NewProcessorPool(&ProcessorPoolConfig{
		Context: ctx,
	}, nil, nil, nil, nil)

	n := 0
	done := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, func(w http.ResponseWriter, req *http.Request) {
		n++
		switch n {
		case 1:
			t.Logf("responding 404")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "no\n")
			return
		case 2:
			t.Logf("responding 200 with busted JSON")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "{bork\n")
			return
		case 3:
			t.Logf("responding 200 with unexpected JSON")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"etat": "sous"}`)
			return
		case 4:
			t.Logf("responding 200 with JSON of mismatched type")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"state": true}`)
			return
		case 5:
			t.Logf("responding 200 with state=up")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"state": "up"}`)
			return
		default:
			t.Logf("responding 200 with state=down")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"state": "down"}`)
			done <- struct{}{}
			return
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	go i.heartbeatHandler(ts.URL, "")

	for {
		select {
		case <-done:
			cancel()
			return
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
