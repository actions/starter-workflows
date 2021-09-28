package worker

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	gocontext "context"

	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/config"
	"github.com/travis-ci/worker/context"
)

// A ProcessorPool spins up multiple Processors handling build jobs from the
// same queue.
type ProcessorPool struct {
	Context                 gocontext.Context
	Provider                backend.Provider
	Generator               BuildScriptGenerator
	Persister               BuildTracePersister
	CancellationBroadcaster *CancellationBroadcaster
	Hostname                string
	Config                  *config.Config

	queue            JobQueue
	logWriterFactory LogWriterFactory
	poolErrors       []error
	processorsLock   sync.Mutex
	processors       []*Processor
	processorsWG     sync.WaitGroup
	pauseCount       int

	activeProcessorCount  int32
	firstProcessorStarted chan struct{}
	startProcessorOnce    sync.Once
}

type ProcessorPoolConfig struct {
	Hostname string
	Context  gocontext.Context
	Config   *config.Config
}

// NewProcessorPool creates a new processor pool using the given arguments.
func NewProcessorPool(ppc *ProcessorPoolConfig,
	provider backend.Provider, generator BuildScriptGenerator, persister BuildTracePersister,
	cancellationBroadcaster *CancellationBroadcaster) *ProcessorPool {

	return &ProcessorPool{
		Hostname: ppc.Hostname,
		Context:  ppc.Context,
		Config:   ppc.Config,

		Provider:                provider,
		Generator:               generator,
		Persister:               persister,
		CancellationBroadcaster: cancellationBroadcaster,

		firstProcessorStarted: make(chan struct{}),
	}
}

// Each loops through all the processors in the pool and calls the given
// function for each of them, passing in the index and the processor. The order
// of the processors is the same for the same set of processors.
func (p *ProcessorPool) Each(f func(int, *Processor)) {
	procIDs := []string{}
	procsByID := map[string]*Processor{}

	for _, proc := range p.processors {
		procIDs = append(procIDs, proc.ID)
		procsByID[proc.ID] = proc
	}

	sort.Strings(procIDs)

	for i, procID := range procIDs {
		f(i, procsByID[procID])
	}
}

// Ready returns true if the processor pool is running as expected.
// Returns false if the processor pool has not been started yet.
func (p *ProcessorPool) Ready() bool {
	return p.queue != nil
}

// Size returns the number of processors that are currently running.
//
// This includes processors that are in the process of gracefully shutting down. It's
// important to track these because they are still running jobs and thus still using
// resources that need to be managed and tracked.
func (p *ProcessorPool) Size() int {
	val := atomic.LoadInt32(&p.activeProcessorCount)
	return int(val)
}

// SetSize adjust the pool to run the given number of processors.
//
// This operates in an eventually consistent manner. Because some workers may be
// running jobs, we may not be able to immediately adjust the pool size. Once jobs
// finish, the pool size should rest at the given value.
func (p *ProcessorPool) SetSize(newSize int) {
	// It's important to lock over the whole method rather than use the lock for
	// individual Incr and Decr calls. We don't want other calls to SetSize to see
	// the intermediate state where only some processors have been started, or they
	// will do the wrong math and start the wrong number of processors.
	p.processorsLock.Lock()
	defer p.processorsLock.Unlock()

	cur := len(p.processors)

	if newSize > cur {
		diff := newSize - cur
		for i := 0; i < diff; i++ {
			p.incr()
		}
	} else if newSize < cur {
		diff := cur - newSize
		for i := 0; i < diff; i++ {
			p.decr()
		}
	}
}

// ExpectedSize returns the size of the pool once gracefully shutdown processors
// complete.
//
// After calling SetSize, ExpectedSize will soon reflect the requested new size,
// while Size will include processors that are still processing their last job
// before shutting down.
func (p *ProcessorPool) ExpectedSize() int {
	return len(p.processors)
}

// TotalProcessed returns the sum of all processor ProcessedCount values.
func (p *ProcessorPool) TotalProcessed() int {
	total := 0
	p.Each(func(_ int, pr *Processor) {
		total += pr.ProcessedCount
	})
	return total
}

// Run starts up a number of processors and connects them to the given queue.
// This method stalls until all processors have finished.
func (p *ProcessorPool) Run(poolSize int, queue JobQueue, logWriterFactory LogWriterFactory) error {
	p.queue = queue
	p.logWriterFactory = logWriterFactory
	p.poolErrors = []error{}

	for i := 0; i < poolSize; i++ {
		p.Incr()
	}

	p.waitForFirstProcessor()

	if len(p.poolErrors) > 0 {
		context.LoggerFromContext(p.Context).WithFields(logrus.Fields{
			"self":        "processor_pool",
			"pool_errors": p.poolErrors,
		}).Panic("failed to populate pool")
	}

	p.processorsWG.Wait()

	return nil
}

// GracefulShutdown causes each processor in the pool to start its graceful
// shutdown.
func (p *ProcessorPool) GracefulShutdown(togglePause bool) {
	p.processorsLock.Lock()
	defer p.processorsLock.Unlock()

	logger := context.LoggerFromContext(p.Context).WithField("self", "processor_pool")

	if togglePause {
		p.pauseCount++

		if p.pauseCount == 1 {
			logger.Info("incrementing wait group for pause")
			p.processorsWG.Add(1)
		} else if p.pauseCount == 2 {
			logger.Info("finishing wait group to unpause")
			p.processorsWG.Done()
		} else if p.pauseCount > 2 {
			return
		}
	}

	// In case no processors were ever started, we still want a graceful shutdown
	// request to proceed. Without this, we will wait forever until the process is
	// forcefully killed.
	p.startProcessorOnce.Do(func() {
		close(p.firstProcessorStarted)
	})

	ps := len(p.processors)
	for i := 0; i < ps; i++ {
		// Use decr to make sure the processor is removed from the list in the pool
		p.decr()
	}
}

// Incr adds a single running processor to the pool
func (p *ProcessorPool) Incr() {
	p.processorsLock.Lock()
	defer p.processorsLock.Unlock()

	p.incr()
}

// incr assumes the processorsLock has already been locked
func (p *ProcessorPool) incr() {
	proc, err := p.makeProcessor(p.queue, p.logWriterFactory)
	if err != nil {
		context.LoggerFromContext(p.Context).WithFields(logrus.Fields{
			"err":  err,
			"self": "processor_pool",
		}).Error("couldn't create processor")
		p.poolErrors = append(p.poolErrors, err)
		return
	}

	p.processors = append(p.processors, proc)
	p.processorsWG.Add(1)

	go func() {
		defer p.processorsWG.Done()

		atomic.AddInt32(&p.activeProcessorCount, 1)
		proc.Run()
		atomic.AddInt32(&p.activeProcessorCount, -1)
	}()

	p.startProcessorOnce.Do(func() {
		close(p.firstProcessorStarted)
	})
}

// Decr pops a processor out of the pool and issues a graceful shutdown
func (p *ProcessorPool) Decr() {
}

// decr assumes the processorsLock has already been locked
func (p *ProcessorPool) decr() {
	if len(p.processors) == 0 {
		return
	}

	var proc *Processor
	proc, p.processors = p.processors[len(p.processors)-1], p.processors[:len(p.processors)-1]
	proc.GracefulShutdown()
}

func (p *ProcessorPool) makeProcessor(queue JobQueue, logWriterFactory LogWriterFactory) (*Processor, error) {
	processorUUID := uuid.NewRandom()
	processorID := fmt.Sprintf("%s@%d.%s", processorUUID.String(), os.Getpid(), p.Hostname)
	ctx := context.FromProcessor(p.Context, processorID)

	return NewProcessor(ctx, p.Hostname,
		queue, logWriterFactory, p.Provider, p.Generator, p.Persister, p.CancellationBroadcaster,
		ProcessorConfig{
			Config: p.Config,
		})
}

func (p *ProcessorPool) waitForFirstProcessor() {
	// wait until this channel is closed. the first processor to start running
	// will close it.
	<-p.firstProcessorStarted
}
