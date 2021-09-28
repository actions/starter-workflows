package worker

import (
	"context"
	"reflect"
	"testing"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/pborman/uuid"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/config"
	workerctx "github.com/travis-ci/worker/context"
)

type buildScriptGeneratorFunction func(context.Context, Job) ([]byte, error)

func (bsg buildScriptGeneratorFunction) Generate(ctx context.Context, job Job) ([]byte, error) {
	return bsg(ctx, job)
}

func TestProcessor(t *testing.T) {
	t.Skip("brittle test is brittle :scream_cat:")
	for i, tc := range []struct {
		runSleep           time.Duration
		hardTimeout        time.Duration
		stateEvents        []string
		isCancelled        bool
		hasBrokenLogWriter bool
	}{
		{
			runSleep:    0 * time.Second,
			hardTimeout: 5 * time.Second,
			stateEvents: []string{"received", "started", string(FinishStatePassed)},
		},
		{
			runSleep:    5 * time.Second,
			hardTimeout: 6 * time.Second,
			stateEvents: []string{"received", "started", string(FinishStateCancelled)},
			isCancelled: true,
		},
		{
			runSleep:           0 * time.Second,
			hardTimeout:        5 * time.Second,
			stateEvents:        []string{"received", "started", "requeued"},
			hasBrokenLogWriter: true,
		},
		{
			runSleep:    5 * time.Second,
			hardTimeout: 4 * time.Second,
			stateEvents: []string{"received", "started", string(FinishStateErrored)},
		},
	} {
		jobID := uint64(100 + i)
		uuid := uuid.NewRandom()
		ctx := workerctx.FromProcessor(context.TODO(), uuid.String())

		provider, err := backend.NewBackendProvider("fake", config.ProviderConfigFromMap(map[string]string{
			"RUN_SLEEP":  tc.runSleep.String(),
			"LOG_OUTPUT": "hello, world",
		}))
		if err != nil {
			t.Error(err)
		}

		generator := buildScriptGeneratorFunction(func(ctx context.Context, job Job) ([]byte, error) {
			return []byte("hello, world"), nil
		})

		jobChan := make(chan Job)
		jobQueue := &fakeJobQueue{c: jobChan}
		cancellationBroadcaster := NewCancellationBroadcaster()

		processor, err := NewProcessor(ctx, "test-hostname", jobQueue, nil, provider, generator, nil, cancellationBroadcaster, ProcessorConfig{
			Config: &config.Config{
				HardTimeout:             tc.hardTimeout,
				LogTimeout:              time.Second,
				ScriptUploadTimeout:     3 * time.Second,
				StartupTimeout:          4 * time.Second,
				MaxLogLength:            4500000,
				PayloadFilterExecutable: "filter.py",
			},
		})
		if err != nil {
			t.Error(err)
		}

		doneChan := make(chan struct{})
		go func() {
			processor.Run()
			doneChan <- struct{}{}
		}()

		rawPayload, _ := simplejson.NewJson([]byte("{}"))

		job := &fakeJob{
			rawPayload: rawPayload,
			payload: &JobPayload{
				Type: "job:test",
				Job: JobJobPayload{
					ID:     jobID,
					Number: "3.1",
				},
				Build: BuildPayload{
					ID:     1,
					Number: "3",
				},
				Repository: RepositoryPayload{
					ID:   4,
					Slug: "green-eggs/ham",
				},
				UUID:     "foo-bar",
				Config:   map[string]interface{}{},
				Timeouts: TimeoutsPayload{},
			},
			startAttributes:    &backend.StartAttributes{},
			hasBrokenLogWriter: tc.hasBrokenLogWriter,
		}

		if tc.isCancelled {
			go func(sl time.Duration, i uint64) {
				time.Sleep(sl)
				cancellationBroadcaster.Broadcast(CancellationCommand{JobID: i})
			}(tc.runSleep-1, jobID)
		}

		jobChan <- job

		processor.GracefulShutdown()
		<-doneChan

		if processor.ProcessedCount != 1 {
			t.Errorf("processor.ProcessedCount = %d, expected %d", processor.ProcessedCount, 1)
		}

		if !reflect.DeepEqual(tc.stateEvents, job.events) {
			t.Errorf("job.events = %#v, expected %#v", job.events, tc.stateEvents)
		}
	}
}
