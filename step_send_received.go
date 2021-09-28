package worker

import (
	gocontext "context"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

type stepSendReceived struct{}

func (s *stepSendReceived) Run(state multistep.StateBag) multistep.StepAction {
	buildJob := state.Get("buildJob").(Job)
	ctx := state.Get("ctx").(gocontext.Context)

	defer context.TimeSince(ctx, "step_send_received_run", time.Now())

	ctx, span := trace.StartSpan(ctx, "SendReceived.Run")
	defer span.End()

	err := buildJob.Received(ctx)
	if err != nil {
		context.LoggerFromContext(ctx).WithFields(logrus.Fields{
			"err":  err,
			"self": "step_send_received",
		}).Error("couldn't send received event")
	}

	return multistep.ActionContinue
}

func (s *stepSendReceived) Cleanup(state multistep.StateBag) {}
