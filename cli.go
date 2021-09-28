package worker

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// include for conditional pprof HTTP server
	_ "net/http/pprof"

	gocontext "context"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/cenk/backoff"
	"github.com/getsentry/raven-go"
	librato "github.com/mihasya/go-metrics-librato"
	"github.com/pkg/errors"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/config"
	"github.com/travis-ci/worker/context"
	"github.com/travis-ci/worker/image"
	travismetrics "github.com/travis-ci/worker/metrics"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	scopeTraceAppend = "https://www.googleapis.com/auth/trace.append"
)

var (
	rootContext = gocontext.TODO()
)

// CLI is the top level of execution for the whole shebang
type CLI struct {
	c *cli.Context

	bootTime time.Time

	ctx    gocontext.Context
	cancel gocontext.CancelFunc
	logger *logrus.Entry

	Config                  *config.Config
	BuildScriptGenerator    BuildScriptGenerator
	BuildTracePersister     BuildTracePersister
	BackendProvider         backend.Provider
	ProcessorPool           *ProcessorPool
	CancellationBroadcaster *CancellationBroadcaster
	JobQueue                JobQueue
	LogWriterFactory        LogWriterFactory

	heartbeatErrSleep time.Duration
	heartbeatSleep    time.Duration
}

// NewCLI creates a new *CLI from a *cli.Context
func NewCLI(c *cli.Context) *CLI {
	return &CLI{
		c:        c,
		bootTime: time.Now().UTC(),

		heartbeatSleep:    5 * time.Minute,
		heartbeatErrSleep: 30 * time.Second,

		CancellationBroadcaster: NewCancellationBroadcaster(),
	}
}

// Setup runs one-time preparatory actions and returns a boolean success value
// that is used to determine if it is safe to invoke the Run func
func (i *CLI) Setup() (bool, error) {
	if i.c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel := gocontext.WithCancel(gocontext.Background())
	logger := context.LoggerFromContext(ctx).WithField("self", "cli")

	i.ctx = ctx
	rootContext = ctx
	i.cancel = cancel
	i.logger = logger

	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	i.Config = config.FromCLIContext(i.c)

	if i.c.Bool("echo-config") {
		config.WriteEnvConfig(i.Config, os.Stdout)
		return false, nil
	}

	if i.c.Bool("list-backend-providers") {
		backend.EachBackend(func(b *backend.Backend) {
			fmt.Println(b.Alias)
		})
		return false, nil
	}

	if i.c.Bool("update-images") {
		baseURL, err := url.Parse(i.Config.ProviderConfig.Get("IMAGE_SELECTOR_URL"))
		if err != nil {
			return false, err
		}

		imageBaseURL, err := url.Parse(i.Config.ProviderConfig.Get("IMAGE_SERVER_URL"))
		if err != nil {
			return false, err
		}

		selector := image.NewAPISelector(baseURL)
		manager, err := image.NewManager(ctx, selector, imageBaseURL)
		if err != nil {
			logger.WithField("err", err).Error("failed to init image manager")
			return false, err
		}

		err = manager.Update(ctx)
		if err != nil {
			logger.WithField("err", err).Error("failed to update images")
		}

		return false, err
	}

	logger.WithField("cfg", fmt.Sprintf("%#v", i.Config)).Debug("read config")

	i.setupSentry()
	i.setupMetrics()

	err := i.setupOpenCensus(ctx)
	if err != nil {
		logger.WithField("err", err).Error("failed to set up opencensus")
		return false, err
	}

	ctx, span := trace.StartSpan(ctx, "CLI.Setup")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("provider", i.Config.ProviderName))

	generator := NewBuildScriptGenerator(i.Config)
	logger.WithField("build_script_generator", fmt.Sprintf("%#v", generator)).Debug("built")

	i.BuildScriptGenerator = generator

	persister := NewBuildTracePersister(i.Config)
	logger.WithField("build_trace_persister", fmt.Sprintf("%#v", persister)).Debug("built")

	i.BuildTracePersister = persister

	if i.Config.TravisSite != "" {
		i.Config.ProviderConfig.Set("TRAVIS_SITE", i.Config.TravisSite)
	}

	provider, err := backend.NewBackendProvider(i.Config.ProviderName, i.Config.ProviderConfig)
	if err != nil {
		logger.WithField("err", err).Error("couldn't create backend provider")
		return false, err
	}

	err = provider.Setup(ctx)
	if err != nil {
		logger.WithField("err", err).Error("couldn't setup backend provider")
		return false, err
	}

	logger.WithField("provider", fmt.Sprintf("%#v", provider)).Debug("built")

	i.BackendProvider = provider

	ppc := &ProcessorPoolConfig{
		Hostname: i.Config.Hostname,
		Context:  rootContext,
		Config:   i.Config,
	}

	pool := NewProcessorPool(ppc, i.BackendProvider, i.BuildScriptGenerator, i.BuildTracePersister, i.CancellationBroadcaster)

	logger.WithField("pool", pool).Debug("built")

	i.ProcessorPool = pool

	if i.c.String("remote-controller-addr") != "" {
		if i.c.String("remote-controller-auth") != "" {
			i.setupRemoteController()
		} else {
			i.logger.Info("skipping remote controller setup without remote-controller-auth set")
		}
		go func() {
			httpAddr := i.c.String("remote-controller-addr")
			i.logger.Info("listening at ", httpAddr)
			_ = http.ListenAndServe(httpAddr, nil)
		}()
	}

	err = i.setupJobQueueAndCanceller()
	if err != nil {
		logger.WithField("err", err).Error("couldn't create job queue and canceller")
		return false, err
	}

	err = i.setupLogWriterFactory()
	if err != nil {
		logger.WithField("err", err).Error("couldn't create logs queue")
		return false, err
	}

	return true, nil
}

// Run starts all long-running processes and blocks until the processor pool
// returns from its Run func
func (i *CLI) Run() {
	i.logger.Info("starting")

	i.handleStartHook()
	defer i.handleStopHook()

	i.logger.Info("worker started")
	defer i.logProcessorInfo("worker finished")

	i.logger.Info("setting up heartbeat")
	i.setupHeartbeat()

	i.logger.Info("starting signal handler loop")
	go i.signalHandler()

	i.logger.WithFields(logrus.Fields{
		"pool_size":         i.Config.PoolSize,
		"queue":             i.JobQueue,
		"logwriter_factory": i.LogWriterFactory,
	}).Debug("running pool")

	_ = i.ProcessorPool.Run(i.Config.PoolSize, i.JobQueue, i.LogWriterFactory)

	err := i.JobQueue.Cleanup()
	if err != nil {
		i.logger.WithField("err", err).Error("couldn't clean up job queue")
	}

	if i.LogWriterFactory != nil {
		err := i.LogWriterFactory.Cleanup()
		if err != nil {
			i.logger.WithField("err", err).Error("couldn't clean up logs queue")
		}
	}
}

func (i *CLI) setupHeartbeat() {
	hbURL := i.c.String("heartbeat-url")
	if hbURL == "" {
		return
	}

	hbTok := i.c.String("heartbeat-url-auth-token")
	if strings.HasPrefix(hbTok, "file://") {
		hbTokBytes, err := ioutil.ReadFile(strings.Split(hbTok, "://")[1])
		if err != nil {
			i.logger.WithField("err", err).Error("failed to read auth token from file")
		} else {
			hbTok = string(hbTokBytes)
		}
	}

	i.logger.WithField("heartbeat_url", hbURL).Info("starting heartbeat loop")
	go i.heartbeatHandler(hbURL, strings.TrimSpace(hbTok))
}

func (i *CLI) handleStartHook() {
	hookValue := i.c.String("start-hook")
	if hookValue == "" {
		return
	}

	i.logger.WithField("start_hook", hookValue).Info("running start hook")

	parts := stringSplitSpace(hookValue)
	outErr, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err == nil {
		return
	}

	i.logger.WithFields(logrus.Fields{
		"err":        err,
		"output":     string(outErr),
		"start_hook": hookValue,
	}).Error("start hook failed")
}

func (i *CLI) handleStopHook() {
	hookValue := i.c.String("stop-hook")
	if hookValue == "" {
		return
	}

	i.logger.WithField("stop_hook", hookValue).Info("running stop hook")

	parts := stringSplitSpace(hookValue)
	outErr, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err == nil {
		return
	}

	i.logger.WithFields(logrus.Fields{
		"err":       err,
		"output":    string(outErr),
		"stop_hook": hookValue,
	}).Error("stop hook failed")
}

func (i *CLI) setupSentry() {
	if i.Config.SentryDSN == "" {
		return
	}

	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
	}

	if i.Config.SentryHookErrors {
		levels = append(levels, logrus.ErrorLevel)
	}

	sentryHook, err := NewSentryHook(i.Config.SentryDSN, levels)

	if err != nil {
		i.logger.WithField("err", err).Error("couldn't create sentry hook")
	}

	logrus.AddHook(sentryHook)

	err = raven.SetDSN(i.Config.SentryDSN)
	if err != nil {
		i.logger.WithField("err", err).Error("couldn't set DSN in raven")
	}

	raven.SetRelease(VersionString)
}

func (i *CLI) setupMetrics() {
	go travismetrics.ReportMemstatsMetrics()

	if i.Config.LibratoEmail != "" && i.Config.LibratoToken != "" && i.Config.LibratoSource != "" {
		i.logger.Info("starting librato metrics reporter")

		go librato.Librato(metrics.DefaultRegistry, time.Minute,
			i.Config.LibratoEmail, i.Config.LibratoToken, i.Config.LibratoSource,
			[]float64{0.50, 0.75, 0.90, 0.95, 0.99, 0.999, 1.0}, time.Millisecond)
	}

	if i.c.Bool("log-metrics") {
		i.logger.Info("starting logger metrics reporter")

		go metrics.Log(metrics.DefaultRegistry, time.Minute,
			log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}
}

func loadStackdriverTraceJSON(ctx gocontext.Context, stackdriverTraceAccountJSON string) (*google.Credentials, error) {
	if stackdriverTraceAccountJSON == "" {
		creds, err := google.FindDefaultCredentials(ctx, scopeTraceAppend)
		return creds, errors.Wrap(err, "could not build default client")
	}

	credBytes, err := loadBytes(stackdriverTraceAccountJSON)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(ctx, credBytes, scopeTraceAppend)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func loadBytes(filenameOrJSON string) ([]byte, error) {
	var (
		bytes []byte
		err   error
	)

	if strings.HasPrefix(strings.TrimSpace(filenameOrJSON), "{") {
		bytes = []byte(filenameOrJSON)
	} else {
		bytes, err = ioutil.ReadFile(filenameOrJSON)
		if err != nil {
			return nil, err
		}
	}

	return bytes, nil
}

func (i *CLI) setupOpenCensus(ctx gocontext.Context) error {
	opencensusEnabled := i.Config.OpencensusTracingEnabled

	if !opencensusEnabled {
		return nil
	}

	creds, err := loadStackdriverTraceJSON(ctx, i.Config.StackdriverTraceAccountJSON)
	if err != nil {
		return err
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: i.Config.StackdriverProjectID,
		TraceClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
		MonitoringClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
	})

	if err != nil {
		return err
	}

	defer sd.Flush()

	// Register/enable the trace exporter
	trace.RegisterExporter(sd)

	traceSampleRate := i.Config.OpencensusSamplingRate
	if traceSampleRate <= 0 {
		i.logger.WithFields(logrus.Fields{
			"trace_sample_rate": traceSampleRate,
		}).Error("trace sample rate must be positive")
		return errors.New("invalid trace sample rate")
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1.0 / float64(traceSampleRate))})
	return nil
}

func (i *CLI) heartbeatHandler(heartbeatURL, heartbeatAuthToken string) {
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 10 * time.Second
	b.MaxElapsedTime = time.Minute

	for {
		err := backoff.Retry(func() error {
			return i.heartbeatCheck(heartbeatURL, heartbeatAuthToken)
		}, b)

		if err != nil {
			i.logger.WithFields(logrus.Fields{
				"heartbeat_url": heartbeatURL,
				"err":           err,
			}).Warn("failed to get heartbeat")
			time.Sleep(i.heartbeatErrSleep)
			continue
		}

		select {
		case <-i.ctx.Done():
			return
		default:
			time.Sleep(i.heartbeatSleep)
		}
	}
}

func (i *CLI) heartbeatCheck(heartbeatURL, heartbeatAuthToken string) error {
	req, err := http.NewRequest("GET", heartbeatURL, nil)
	if err != nil {
		return err
	}

	if heartbeatAuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", heartbeatAuthToken))
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("unhappy status code %d", resp.StatusCode)
	}

	body := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return err
	}

	if state, ok := body["state"]; ok && state == "down" {
		i.logger.WithField("heartbeat_state", state).Info("starting graceful shutdown")
		i.ProcessorPool.GracefulShutdown(false)
	}
	return nil
}

func (i *CLI) setupRemoteController() {
	i.logger.Info("setting up remote controller")
	(&RemoteController{
		pool:       i.ProcessorPool,
		auth:       i.c.String("remote-controller-auth"),
		workerInfo: i.workerInfo,
		cancel:     i.cancel,
	}).Setup()
}

func (i *CLI) workerInfo() workerInfo {
	info := workerInfo{
		Version:          VersionString,
		Revision:         RevisionString,
		Generated:        GeneratedString,
		Uptime:           time.Since(i.bootTime).String(),
		PoolSize:         i.ProcessorPool.Size(),
		ExpectedPoolSize: i.ProcessorPool.ExpectedSize(),
		TotalProcessed:   i.ProcessorPool.TotalProcessed(),
	}

	i.ProcessorPool.Each(func(_ int, p *Processor) {
		info.Processors = append(info.Processors, p.processorInfo())
	})

	return info
}

func (i *CLI) signalHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1,
		syscall.SIGTTIN, syscall.SIGTTOU,
		syscall.SIGUSR2)

	for {
		select {
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGINT:
				i.logger.Warn("SIGINT received, starting graceful shutdown")
				i.ProcessorPool.GracefulShutdown(false)
			case syscall.SIGTERM:
				i.logger.Warn("SIGTERM received, shutting down immediately")
				i.cancel()
			case syscall.SIGTTIN:
				i.logger.Info("SIGTTIN received, adding processor to pool")
				i.ProcessorPool.Incr()
			case syscall.SIGTTOU:
				i.logger.Info("SIGTTOU received, removing processor from pool")
				i.ProcessorPool.Decr()
			case syscall.SIGUSR2:
				i.logger.Warn("SIGUSR2 received, toggling graceful shutdown and pause")
				i.ProcessorPool.GracefulShutdown(true)
			case syscall.SIGUSR1:
				i.logProcessorInfo("received SIGUSR1")
			default:
				i.logger.WithField("signal", sig).Info("ignoring unknown signal")
			}
		default:
			time.Sleep(time.Second)
		}
	}
}

func (i *CLI) logProcessorInfo(msg string) {
	if msg == "" {
		msg = "processor pool info"
	}
	i.logger.WithFields(logrus.Fields{
		"version":         VersionString,
		"revision":        RevisionString,
		"generated":       GeneratedString,
		"boot_time":       i.bootTime.String(),
		"uptime_min":      time.Since(i.bootTime).Minutes(),
		"pool_size":       i.ProcessorPool.Size(),
		"total_processed": i.ProcessorPool.TotalProcessed(),
	}).Info(msg)
	i.ProcessorPool.Each(func(n int, proc *Processor) {
		i.logger.WithFields(logrus.Fields{
			"n":           n,
			"id":          proc.ID,
			"processed":   proc.ProcessedCount,
			"status":      proc.CurrentStatus,
			"last_job_id": proc.LastJobID,
		}).Info("processor info")
	})
}

func (i *CLI) setupJobQueueAndCanceller() error {
	subQueues := []JobQueue{}
	for _, queueType := range strings.Split(i.Config.QueueType, ",") {
		queueType = strings.TrimSpace(queueType)

		switch queueType {
		case "amqp":
			jobQueue, canceller, err := i.buildAMQPJobQueueAndCanceller()
			if err != nil {
				return err
			}
			go canceller.Run()
			subQueues = append(subQueues, jobQueue)
		case "file":
			jobQueue, err := i.buildFileJobQueue()
			if err != nil {
				return err
			}
			subQueues = append(subQueues, jobQueue)
		case "http":
			jobQueue, err := i.buildHTTPJobQueue()
			if err != nil {
				return err
			}
			subQueues = append(subQueues, jobQueue)
		default:
			return fmt.Errorf("unknown queue type %q", queueType)
		}
	}

	if len(subQueues) == 0 {
		return fmt.Errorf("no queues built")
	}

	if len(subQueues) == 1 {
		i.JobQueue = subQueues[0]
	} else {
		i.JobQueue = NewMultiSourceJobQueue(subQueues...)
	}
	return nil
}

func (i *CLI) buildAMQPJobQueueAndCanceller() (*AMQPJobQueue, *AMQPCanceller, error) {
	var amqpConn *amqp.Connection
	var err error

	if i.Config.AmqpTlsCert != "" || i.Config.AmqpTlsCertPath != "" {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()
		if i.Config.AmqpTlsCert != "" {
			cfg.RootCAs.AppendCertsFromPEM([]byte(i.Config.AmqpTlsCert))
		}
		if i.Config.AmqpTlsCertPath != "" {
			cert, err := ioutil.ReadFile(i.Config.AmqpTlsCertPath)
			if err != nil {
				return nil, nil, err
			}
			cfg.RootCAs.AppendCertsFromPEM(cert)
		}
		amqpConn, err = amqp.DialConfig(i.Config.AmqpURI,
			amqp.Config{
				Heartbeat:       i.Config.AmqpHeartbeat,
				Locale:          "en_US",
				TLSClientConfig: cfg,
			})
	} else if i.Config.AmqpInsecure {
		amqpConn, err = amqp.DialConfig(
			i.Config.AmqpURI,
			amqp.Config{
				Heartbeat:       i.Config.AmqpHeartbeat,
				Locale:          "en_US",
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			})
	} else {
		amqpConn, err = amqp.DialConfig(i.Config.AmqpURI,
			amqp.Config{
				Heartbeat: i.Config.AmqpHeartbeat,
				Locale:    "en_US",
			})
	}
	if err != nil {
		i.logger.WithField("err", err).Error("couldn't connect to AMQP")
		return nil, nil, err
	}

	go i.amqpErrorWatcher(amqpConn)

	i.logger.Debug("connected to AMQP")

	canceller := NewAMQPCanceller(i.ctx, amqpConn, i.CancellationBroadcaster)
	i.logger.WithField("canceller", fmt.Sprintf("%#v", canceller)).Debug("built")

	jobQueue, err := NewAMQPJobQueue(amqpConn, i.Config.QueueName, i.Config.StateUpdatePoolSize, i.Config.RabbitMQSharding)

	if err != nil {
		return nil, nil, err
	}

	// Set the consumer priority directly instead of altering the signature of
	// NewAMQPJobQueue :sigh_cat:
	jobQueue.priority = i.Config.AmqpConsumerPriority

	jobQueue.DefaultLanguage = i.Config.DefaultLanguage
	jobQueue.DefaultDist = i.Config.DefaultDist
	jobQueue.DefaultArch = i.Config.DefaultArch
	jobQueue.DefaultGroup = i.Config.DefaultGroup
	jobQueue.DefaultOS = i.Config.DefaultOS

	return jobQueue, canceller, nil
}

func (i *CLI) buildHTTPJobQueue() (*HTTPJobQueue, error) {
	jobBoardURL, err := url.Parse(i.Config.JobBoardURL)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing job board URL")
	}

	jobQueue, err := NewHTTPJobQueueWithIntervals(
		jobBoardURL, i.Config.TravisSite,
		i.Config.ProviderName, i.Config.QueueName,
		i.Config.HTTPPollingInterval, i.Config.HTTPRefreshClaimInterval,
		i.CancellationBroadcaster)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP job queue")
	}

	jobQueue.DefaultLanguage = i.Config.DefaultLanguage
	jobQueue.DefaultDist = i.Config.DefaultDist
	jobQueue.DefaultArch = i.Config.DefaultArch
	jobQueue.DefaultGroup = i.Config.DefaultGroup
	jobQueue.DefaultOS = i.Config.DefaultOS

	return jobQueue, nil
}

func (i *CLI) buildFileJobQueue() (*FileJobQueue, error) {
	jobQueue, err := NewFileJobQueue(
		i.Config.BaseDir, i.Config.QueueName, i.Config.FilePollingInterval)
	if err != nil {
		return nil, err
	}

	jobQueue.DefaultLanguage = i.Config.DefaultLanguage
	jobQueue.DefaultDist = i.Config.DefaultDist
	jobQueue.DefaultArch = i.Config.DefaultArch
	jobQueue.DefaultGroup = i.Config.DefaultGroup
	jobQueue.DefaultOS = i.Config.DefaultOS

	return jobQueue, nil
}

func (i *CLI) setupLogWriterFactory() error {
	if i.Config.LogsAmqpURI == "" {
		// If no separate URI is set for LogsAMQP, use the JobsQueue to send log parts
		return nil
	}
	logWriterFactory, err := i.buildAMQPLogWriterFactory()
	if err != nil {
		return err
	}
	i.LogWriterFactory = logWriterFactory
	return nil
}

func (i *CLI) buildAMQPLogWriterFactory() (*AMQPLogWriterFactory, error) {
	var amqpConn *amqp.Connection
	var err error

	if i.Config.LogsAmqpTlsCert != "" || i.Config.LogsAmqpTlsCertPath != "" {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()
		if i.Config.LogsAmqpTlsCert != "" {
			cfg.RootCAs.AppendCertsFromPEM([]byte(i.Config.LogsAmqpTlsCert))
		}
		if i.Config.LogsAmqpTlsCertPath != "" {
			cert, err := ioutil.ReadFile(i.Config.LogsAmqpTlsCertPath)
			if err != nil {
				return nil, err
			}
			cfg.RootCAs.AppendCertsFromPEM(cert)
		}
		amqpConn, err = amqp.DialConfig(i.Config.LogsAmqpURI,
			amqp.Config{
				Heartbeat:       i.Config.AmqpHeartbeat,
				Locale:          "en_US",
				TLSClientConfig: cfg,
			})
	} else if i.Config.AmqpInsecure {
		amqpConn, err = amqp.DialConfig(
			i.Config.LogsAmqpURI,
			amqp.Config{
				Heartbeat:       i.Config.AmqpHeartbeat,
				Locale:          "en_US",
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			})
	} else {
		amqpConn, err = amqp.DialConfig(i.Config.LogsAmqpURI,
			amqp.Config{
				Heartbeat: i.Config.AmqpHeartbeat,
				Locale:    "en_US",
			})
	}
	if err != nil {
		i.logger.WithField("err", err).Error("couldn't connect to the logs AMQP server")
		return nil, err
	}

	go i.amqpErrorWatcher(amqpConn)
	i.logger.Debug("connected to the logs AMQP server")

	logWriterFactory, err := NewAMQPLogWriterFactory(amqpConn, i.Config.RabbitMQSharding)
	if err != nil {
		return nil, err
	}
	return logWriterFactory, nil
}

func (i *CLI) amqpErrorWatcher(amqpConn *amqp.Connection) {
	errChan := make(chan *amqp.Error)
	errChan = amqpConn.NotifyClose(errChan)

	err, ok := <-errChan
	if ok {
		i.logger.WithField("err", err).Error("amqp connection errored, terminating")
		i.cancel()
		time.Sleep(time.Minute)
		i.logger.Panic("timed out waiting for shutdown after amqp connection error")
	}
}
