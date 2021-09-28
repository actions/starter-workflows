package worker

import (
	"time"

	gocontext "context"

	"github.com/mitchellh/multistep"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/config"
	"github.com/travis-ci/worker/context"
	"go.opencensus.io/trace"
)

// A Processor gets jobs off the job queue and coordinates running it with other
// components.
type Processor struct {
	ID       string
	hostname string
	config   *config.Config

	ctx                     gocontext.Context
	buildJobsChan           <-chan Job
	provider                backend.Provider
	generator               BuildScriptGenerator
	persister               BuildTracePersister
	logWriterFactory        LogWriterFactory
	cancellationBroadcaster *CancellationBroadcaster

	graceful   chan struct{}
	terminate  gocontext.CancelFunc
	shutdownAt time.Time

	// ProcessedCount contains the number of jobs that has been processed
	// by this Processor. This value should not be modified outside of the
	// Processor.
	ProcessedCount int

	// CurrentStatus contains the current status of the processor, and can
	// be one of "new", "waiting", "processing" or "done".
	CurrentStatus string

	// LastJobID contains the ID of the last job the processor processed.
	LastJobID uint64
}

type ProcessorConfig struct {
	Config *config.Config
}

// NewProcessor creates a new processor that will run the build jobs on the
// given channel using the given provider and getting build scripts from the
// generator.
func NewProcessor(ctx gocontext.Context, hostname string, queue JobQueue,
	logWriterFactory LogWriterFactory, provider backend.Provider, generator BuildScriptGenerator, persister BuildTracePersister, cancellationBroadcaster *CancellationBroadcaster,
	config ProcessorConfig) (*Processor, error) {

	processorID, _ := context.ProcessorFromContext(ctx)

	ctx, cancel := gocontext.WithCancel(ctx)

	buildJobsChan, err := queue.Jobs(ctx)
	if err != nil {
		context.LoggerFromContext(ctx).WithField("err", err).Error("couldn't create jobs channel")
		cancel()
		return nil, err
	}

	return &Processor{
		ID:       processorID,
		hostname: hostname,
		config:   config.Config,

		ctx:                     ctx,
		buildJobsChan:           buildJobsChan,
		provider:                provider,
		generator:               generator,
		persister:               persister,
		cancellationBroadcaster: cancellationBroadcaster,
		logWriterFactory:        logWriterFactory,

		graceful:  make(chan struct{}),
		terminate: cancel,

		CurrentStatus: "new",
	}, nil
}

// Run starts the processor. This method will not return until the processor is
// terminated, either by calling the GracefulShutdown or Terminate methods, or
// if the build jobs channel is closed.
func (p *Processor) Run() {
	logger := context.LoggerFromContext(p.ctx).WithField("self", "processor")
	logger.Info("starting processor")
	defer logger.Info("processor done")
	defer func() { p.CurrentStatus = "done" }()

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("processor is done, terminating")
			return
		case <-p.graceful:
			logger.WithField("shutdown_duration_s", time.Since(p.shutdownAt).Seconds()).Info("processor is done, terminating")
			p.terminate()
			return
		default:
		}

		select {
		case <-p.ctx.Done():
			logger.Info("processor is done, terminating")
			return
		case <-p.graceful:
			logger.WithField("shutdown_duration_s", time.Since(p.shutdownAt).Seconds()).Info("processor is done, terminating")
			p.terminate()
			return
		case buildJob, ok := <-p.buildJobsChan:
			if !ok {
				p.terminate()
				return
			}

			buildJob.StartAttributes().ProgressType = p.config.ProgressType

			jobID := buildJob.Payload().Job.ID

			hardTimeout := p.config.HardTimeout
			if buildJob.Payload().Timeouts.HardLimit != 0 {
				hardTimeout = time.Duration(buildJob.Payload().Timeouts.HardLimit) * time.Second
			}
			logger.WithFields(logrus.Fields{
				"hard_timeout": hardTimeout,
				"job_id":       jobID,
			}).Debug("setting hard timeout")
			buildJob.StartAttributes().HardTimeout = hardTimeout

			ctx := context.FromJobID(context.FromRepository(p.ctx, buildJob.Payload().Repository.Slug), buildJob.Payload().Job.ID)
			if buildJob.Payload().UUID != "" {
				ctx = context.FromUUID(ctx, buildJob.Payload().UUID)
			} else {
				ctx = context.FromUUID(ctx, uuid.NewRandom().String())
			}
			logger.WithFields(logrus.Fields{
				"job_id": jobID,
				"status": "processing",
			}).Debug("updating processor status and last id")
			p.LastJobID = jobID
			p.CurrentStatus = "processing"

			p.process(ctx, buildJob)

			logger.WithFields(logrus.Fields{
				"job_id": jobID,
				"status": "waiting",
			}).Debug("updating processor status")
			p.CurrentStatus = "waiting"
		case <-time.After(10 * time.Second):
			logger.Debug("timeout waiting for job, shutdown, or context done")
		}
	}
}

// GracefulShutdown tells the processor to finish the job it is currently
// processing, but not pick up any new jobs. This method will return
// immediately, the processor is done when Run() returns.
func (p *Processor) GracefulShutdown() {
	logger := context.LoggerFromContext(p.ctx).WithField("self", "processor")
	defer func() {
		err := recover()
		if err != nil {
			logger.WithField("err", err).Error("recovered from panic")
		}
	}()
	logger.Info("processor initiating graceful shutdown")
	p.shutdownAt = time.Now()
	tryClose(p.graceful)
}

// Terminate tells the processor to stop working on the current job as soon as
// possible.
func (p *Processor) Terminate() {
	p.terminate()
}

func (p *Processor) process(ctx gocontext.Context, buildJob Job) {
	ctx = buildJob.SetupContext(ctx)
	ctx = context.WithTimings(ctx)

	ctx, span := trace.StartSpan(ctx, "ProcessorRun")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", "worker"),
		trace.Int64Attribute("job_id", int64(buildJob.Payload().Job.ID)),
		trace.StringAttribute("repo", buildJob.Payload().Repository.Slug),
		trace.StringAttribute("infra", p.config.ProviderName),
		trace.StringAttribute("site", p.config.TravisSite),
	)

	state := new(multistep.BasicStateBag)
	state.Put("hostname", p.ID)
	state.Put("buildJob", buildJob)
	state.Put("logWriterFactory", p.logWriterFactory)
	state.Put("ctx", ctx)
	state.Put("processedAt", time.Now().UTC())
	state.Put("infra", p.config.Infra)

	logger := context.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"job_id": buildJob.Payload().Job.ID,
		"self":   "processor",
	})

	logTimeout := p.config.LogTimeout
	if buildJob.Payload().Timeouts.LogSilence != 0 {
		logTimeout = time.Duration(buildJob.Payload().Timeouts.LogSilence) * time.Second
	}

	steps := []multistep.Step{
		&stepSubscribeCancellation{
			cancellationBroadcaster: p.cancellationBroadcaster,
		},
		&stepTransformBuildJSON{
			payloadFilterExecutable: p.config.PayloadFilterExecutable,
		},
		&stepGenerateScript{
			generator: p.generator,
		},
		&stepSendReceived{},
		&stepSleep{duration: p.config.InitialSleep},
		&stepCheckCancellation{},
		&stepOpenLogWriter{
			maxLogLength:      p.config.MaxLogLength,
			defaultLogTimeout: p.config.LogTimeout,
		},
		&stepCheckCancellation{},
		&stepStartInstance{
			provider:     p.provider,
			startTimeout: p.config.StartupTimeout,
		},
		&stepCheckCancellation{},
		&stepUploadScript{
			uploadTimeout: p.config.ScriptUploadTimeout,
		},
		&stepCheckCancellation{},
		&stepUpdateState{},
		&stepWriteWorkerInfo{},
		&stepCheckCancellation{},
		&stepRunScript{
			logTimeout:               logTimeout,
			hardTimeout:              buildJob.StartAttributes().HardTimeout,
			skipShutdownOnLogTimeout: p.config.SkipShutdownOnLogTimeout,
		},
		&stepDownloadTrace{
			persister: p.persister,
		},
	}

	runner := &multistep.BasicRunner{Steps: steps}

	logger.Info("starting job")
	runner.Run(state)

	fields := context.LoggerTimingsFromContext(ctx)
	instance, ok := state.Get("instance").(backend.Instance)
	if ok {
		fields["instance_id"] = instance.ID()
		fields["image_name"] = instance.ImageName()
	}
	err, ok := state.Get("err").(error)
	if ok {
		fields["err"] = err
	}
	if buildJob.FinishState() != "" {
		fields["state"] = buildJob.FinishState()
	}
	if buildJob.Requeued() {
		fields["requeued"] = 1
	}
	logger.WithFields(fields).Info("finished job")

	p.ProcessedCount++
}

func (p *Processor) processorInfo() processorInfo {
	return processorInfo{
		ID:        p.ID,
		Processed: p.ProcessedCount,
		Status:    p.CurrentStatus,
		LastJobID: p.LastJobID,
	}
}
