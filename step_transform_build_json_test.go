package worker

import (
	"testing"

	gocontext "golang.org/x/net/context"

	"github.com/bitly/go-simplejson"
	"github.com/mitchellh/multistep"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/config"
)

func setupStepTransformBuildJSON(cfg *config.ProviderConfig) (*stepTransformBuildJSON, multistep.StateBag) {
	s := &stepTransformBuildJSON{}

	bp, _ := backend.NewBackendProvider("fake", cfg)

	ctx := gocontext.TODO()
	instance, _ := bp.Start(ctx, nil)

	rawPayload, _ := simplejson.NewJson([]byte("{}"))

	job := &fakeJob{
		rawPayload: rawPayload,
	}

	state := &multistep.BasicStateBag{}
	state.Put("ctx", ctx)
	state.Put("buildJob", job)
	state.Put("instance", instance)

	return s, state
}

func TestStepTransformBuildJSON_Run(t *testing.T) {

	cfg := config.ProviderConfigFromMap(map[string]string{
		"PAYLOAD_FILTER_EXECUTABLE": "",
	})

	s, state := setupStepTransformBuildJSON(cfg)
	action := s.Run(state)
	assert.Equal(t, multistep.ActionContinue, action)
}

func TestStepTransformBuildJSON_RunWithExecutableConfigured(t *testing.T) {

	cfg := config.ProviderConfigFromMap(map[string]string{
		"PAYLOAD_FILTER_EXECUTABLE": "/usr/local/bin/filter.py",
	})

	s, state := setupStepTransformBuildJSON(cfg)
	action := s.Run(state)
	assert.Equal(t, multistep.ActionContinue, action)
}
