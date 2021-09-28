package worker

import (
	"time"

	gocontext "context"

	"github.com/mitchellh/multistep"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

type stepOpenLogWriter struct {
	maxLogLength      int
	defaultLogTimeout time.Duration
}

func (s *stepOpenLogWriter) Run(state multistep.StateBag) multistep.StepAction {
	ctx := state.Get("ctx").(gocontext.Context)
	buildJob := state.Get("buildJob").(Job)
	logWriterFactory := state.Get("logWriterFactory")
	logger := context.LoggerFromContext(ctx).WithField("self", "step_open_log_writer")

	ctx, span := trace.StartSpan(ctx, "OpenLogWriter.Run")
	defer span.End()

	var logWriter LogWriter
	var err error

	if logWriterFactory != nil {
		logWriter, err = logWriterFactory.(LogWriterFactory).LogWriter(ctx, s.defaultLogTimeout, buildJob)
	} else {
		logWriter, err = buildJob.LogWriter(ctx, s.defaultLogTimeout)
	}
	if err != nil {
		state.Put("err", err)

		logger.WithFields(logrus.Fields{
			"err":         err,
			"log_timeout": s.defaultLogTimeout,
		}).Error("couldn't open a log writer, attempting requeue")
		context.CaptureError(ctx, err)

		err := buildJob.Requeue(ctx)
		if err != nil {
			logger.WithField("err", err).Error("couldn't requeue job")
		}

		return multistep.ActionHalt
	}
	logWriter.SetMaxLogLength(s.maxLogLength)

	state.Put("logWriter", logWriter)

	return multistep.ActionContinue
}

func (s *stepOpenLogWriter) Cleanup(state multistep.StateBag) {
	ctx := state.Get("ctx").(gocontext.Context)

	_, span := trace.StartSpan(ctx, "OpenLogWriter.Cleanup")
	defer span.End()

	logWriter, ok := state.Get("logWriter").(LogWriter)
	if ok {
		logWriter.Close()
	}
}
