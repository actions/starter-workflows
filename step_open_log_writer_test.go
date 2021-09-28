package worker

import (
	"testing"

	gocontext "context"

	"github.com/mitchellh/multistep"
	"github.com/stretchr/testify/assert"
)

func setupStepOpenLogWriter() (*stepOpenLogWriter, multistep.StateBag) {
	s := &stepOpenLogWriter{}

	ctx := gocontext.TODO()

	job := &fakeJob{
		payload: &JobPayload{
			Type: "job:test",
			Job: JobJobPayload{
				ID:     2,
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
	}

	state := &multistep.BasicStateBag{}
	state.Put("ctx", ctx)
	state.Put("buildJob", job)

	return s, state
}

func TestStepOpenLogWriter_Run(t *testing.T) {
	s, state := setupStepOpenLogWriter()
	action := s.Run(state)
	assert.Equal(t, multistep.ActionContinue, action)
	assert.NotNil(t, state.Get("logWriter"))
}
