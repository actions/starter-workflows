package worker

import (
	gocontext "context"
	"errors"
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

var JobCancelledError = errors.New("job cancelled")

type stepCheckCancellation struct{}

func (s *stepCheckCancellation) Run(state multistep.StateBag) multistep.StepAction {
	cancelChan := state.Get("cancelChan").(<-chan CancellationCommand)

	ctx := state.Get("ctx").(gocontext.Context)

	_, span := trace.StartSpan(ctx, "CheckCancellation.Run")
	defer span.End()

	select {
	case command := <-cancelChan:
		ctx := state.Get("ctx").(gocontext.Context)
		buildJob := state.Get("buildJob").(Job)
		if _, ok := state.GetOk("logWriter"); ok {
			logWriter := state.Get("logWriter").(LogWriter)
			s.writeLogAndFinishWithState(ctx, logWriter, buildJob, FinishStateCancelled, fmt.Sprintf("\n\nDone: Job Cancelled\n\n%s", command.Reason))
		} else {
			err := buildJob.Finish(ctx, FinishStateCancelled)
			if err != nil {
				context.LoggerFromContext(ctx).WithField("err", err).WithField("state", FinishStateCancelled).Error("couldn't update job state")
			}
		}
		state.Put("err", JobCancelledError)
		return multistep.ActionHalt
	default:
	}

	return multistep.ActionContinue
}

func (s *stepCheckCancellation) Cleanup(state multistep.StateBag) {}

func (s *stepCheckCancellation) writeLogAndFinishWithState(ctx gocontext.Context, logWriter LogWriter, buildJob Job, state FinishState, logMessage string) {
	ctx, span := trace.StartSpan(ctx, "WriteLogAndFinishWithState.CheckCancellation")
	defer span.End()

	_, err := logWriter.WriteAndClose([]byte(logMessage))
	if err != nil {
		context.LoggerFromContext(ctx).WithField("err", err).Error("couldn't write final log message")
	}

	err = buildJob.Finish(ctx, state)
	if err != nil {
		context.LoggerFromContext(ctx).WithField("err", err).WithField("state", state).Error("couldn't update job state")
	}
}
