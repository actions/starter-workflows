package worker

import (
	"time"

	gocontext "context"

	"github.com/mitchellh/multistep"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/metrics"
	"go.opencensus.io/trace"
)

type stepUploadScript struct {
	uploadTimeout time.Duration
}

func (s *stepUploadScript) Run(state multistep.StateBag) multistep.StepAction {
	ctx := state.Get("ctx").(gocontext.Context)
	buildJob := state.Get("buildJob").(Job)
	logWriter := state.Get("logWriter").(LogWriter)
	processedAt := state.Get("processedAt").(time.Time)

	instance := state.Get("instance").(backend.Instance)
	script := state.Get("script").([]byte)

	logger := context.LoggerFromContext(ctx).WithField("self", "step_upload_script")

	defer context.TimeSince(ctx, "step_upload_script_run", time.Now())

	ctx, span := trace.StartSpan(ctx, "UploadScript.Run")
	defer span.End()

	preTimeoutCtx := ctx

	ctx, cancel := gocontext.WithTimeout(ctx, s.uploadTimeout)
	defer cancel()

	if instance.SupportsProgress() && buildJob.StartAttributes().ProgressType == "text" {
		_, _ = writeFoldStart(logWriter, "step_upload_script", []byte("\033[33;1mUploading script\033[0m\r\n"))
		defer func() {
			_, err := writeFoldEnd(logWriter, "step_upload_script", []byte(""))
			if err != nil {
			    logger.WithFields(logrus.Fields{
					"err":            err,
				}).Error("couldn't write fold end")
			}
		}()
	}

	err := instance.UploadScript(ctx, script)
	if err != nil {
		state.Put("err", err)

		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnavailable,
			Message: err.Error(),
		})

		errMetric := "worker.job.upload.error"
		if errors.Cause(err) == backend.ErrStaleVM {
			errMetric += ".stalevm"
		}
		metrics.Mark(errMetric)

		logger.WithFields(logrus.Fields{
			"err":            err,
			"upload_timeout": s.uploadTimeout,
		}).Error("couldn't upload script, attempting requeue")
		context.CaptureError(ctx, err)

		err := buildJob.Requeue(preTimeoutCtx)
		if err != nil {
			logger.WithField("err", err).Error("couldn't requeue job")
		}

		return multistep.ActionHalt
	}

	logger.WithFields(logrus.Fields{
		"since_processed_ms": time.Since(processedAt).Seconds() * 1e3,
	}).Info("uploaded script")

	return multistep.ActionContinue
}

func (s *stepUploadScript) Cleanup(state multistep.StateBag) {
	// Nothing to clean up
}
