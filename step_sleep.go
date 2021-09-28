package worker

import (
	gocontext "context"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

type stepSleep struct {
	duration time.Duration
}

func (s *stepSleep) Run(state multistep.StateBag) multistep.StepAction {
	ctx := state.Get("ctx").(gocontext.Context)

	defer context.TimeSince(ctx, "step_sleep_run", time.Now())

	_, span := trace.StartSpan(ctx, "Sleep.Run")
	defer span.End()

	time.Sleep(s.duration)

	return multistep.ActionContinue
}

func (s *stepSleep) Cleanup(state multistep.StateBag) {}
