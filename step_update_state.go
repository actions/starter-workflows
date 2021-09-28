package worker

import (
	gocontext "context"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

type stepUpdateState struct{}

func (s *stepUpdateState) Run(state multistep.StateBag) multistep.StepAction {
	ctx := state.Get("ctx").(gocontext.Context)
	buildJob := state.Get("buildJob").(Job)
	instance := state.Get("instance").(backend.Instance)
	processedAt := state.Get("processedAt").(time.Time)
	logWriter := state.Get("logWriter").(LogWriter)

	logger := context.LoggerFromContext(ctx).WithField("self", "step_update_state")

	instanceID := instance.ID()
	if instanceID != "" {
		ctx = context.FromInstanceID(ctx, instanceID)
		state.Put("ctx", ctx)
	}

	defer context.TimeSince(ctx, "step_update_state_run", time.Now())

	ctx, span := trace.StartSpan(ctx, "UpdateState.Run")
	defer span.End()

	logWriter.SetJobStarted(&JobStartedMeta{
		QueuedAt: buildJob.Payload().Job.QueuedAt,
		Repo:     buildJob.Payload().Repository.Slug,
		Queue:    buildJob.Payload().Queue,
		Infra:    state.Get("infra").(string),
	})

	err := buildJob.Started(ctx)
	if err != nil {
		context.LoggerFromContext(ctx).WithFields(logrus.Fields{
			"err":         err,
			"self":        "step_update_state",
			"instance_id": instanceID,
		}).Error("couldn't mark job as started")
	}

	logger.WithFields(logrus.Fields{
		"since_processed_ms": time.Since(processedAt).Seconds() * 1e3,
		"action":             "run",
	}).Info("marked job as started")

	return multistep.ActionContinue
}

func (s *stepUpdateState) Cleanup(state multistep.StateBag) {
	buildJob := state.Get("buildJob").(Job)
	ctx := state.Get("ctx").(gocontext.Context)
	processedAt := state.Get("processedAt").(time.Time)

	instance := state.Get("instance").(backend.Instance)
	instanceID := instance.ID()
	if instanceID != "" {
		ctx = context.FromInstanceID(ctx, instanceID)
		state.Put("ctx", ctx)
	}

	defer context.TimeSince(ctx, "step_update_state_cleanup", time.Now())

	ctx, span := trace.StartSpan(ctx, "UpdateState.Cleanup")
	defer span.End()

	logger := context.LoggerFromContext(ctx).WithField("self", "step_update_state")
	logger.WithFields(logrus.Fields{
		"since_processed_ms": time.Since(processedAt).Seconds() * 1e3,
		"action":             "cleanup",
	}).Info("cleaning up")

	mresult, ok := state.GetOk("scriptResult")

	if ok {
		result := mresult.(*backend.RunResult)

		var err error

		switch result.ExitCode {
		case 0:
			err = buildJob.Finish(ctx, FinishStatePassed)
		case 1:
			err = buildJob.Finish(ctx, FinishStateFailed)
		default:
			err = buildJob.Finish(ctx, FinishStateErrored)
		}

		if err != nil {
			context.LoggerFromContext(ctx).WithFields(logrus.Fields{
				"err":         err,
				"self":        "step_update_state",
				"instance_id": instanceID,
			}).Error("couldn't mark job as finished")
		}
	}
}
