package worker

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
	gocontext "golang.org/x/net/context"
)

type stepTransformBuildJSON struct {
	payloadFilterExecutable string
}

type EnvVar struct {
	Name   string
	Public bool
	Value  string
}

func (s *stepTransformBuildJSON) Run(state multistep.StateBag) multistep.StepAction {
	buildJob := state.Get("buildJob").(Job)
	ctx := state.Get("ctx").(gocontext.Context)

	ctx, span := trace.StartSpan(ctx, "TransformBuildJSON.Run")
	defer span.End()

	if s.payloadFilterExecutable == "" {
		context.LoggerFromContext(ctx).Info("skipping json transformation, no filter executable defined")
		return multistep.ActionContinue
	}

	context.LoggerFromContext(ctx).Info(fmt.Sprintf("calling filter executable: %s", s.payloadFilterExecutable))

	payload := buildJob.RawPayload()

	cmd := exec.Command(s.payloadFilterExecutable)
	rawJson, err := payload.MarshalJSON()
	if err != nil {
		context.LoggerFromContext(ctx).Info(fmt.Sprintf("failed to marshal json: %v", err))
		return multistep.ActionContinue
	}

	cmd.Stdin = strings.NewReader(string(rawJson))

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		context.LoggerFromContext(ctx).Info(fmt.Sprintf("failed to run filter executable: %v", err))
		return multistep.ActionContinue
	}

	err = payload.UnmarshalJSON(out.Bytes())
	if err != nil {
		context.LoggerFromContext(ctx).Info(fmt.Sprintf("failed to unmarshal json: %v", err))
		return multistep.ActionContinue
	}

	context.LoggerFromContext(ctx).Info("replaced the build json")

	return multistep.ActionContinue
}

func (s *stepTransformBuildJSON) Cleanup(multistep.StateBag) {
	// Nothing to clean up
}
