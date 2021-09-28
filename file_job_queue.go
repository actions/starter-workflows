package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	gocontext "context"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
)

// FileJobQueue is a JobQueue that uses directories for input, state, and output
type FileJobQueue struct {
	queue           string
	pollingInterval time.Duration

	buildJobChan chan Job

	baseDir     string
	createdDir  string
	receivedDir string
	startedDir  string
	finishedDir string
	logDir      string

	DefaultLanguage, DefaultDist, DefaultArch, DefaultGroup, DefaultOS string
}

// NewFileJobQueue creates a *FileJobQueue from a base directory and queue name
func NewFileJobQueue(baseDir, queue string, pollingInterval time.Duration) (*FileJobQueue, error) {
	_, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}

	fd, err := os.Create(filepath.Join(baseDir, ".write-test"))
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	createdDir := filepath.Join(baseDir, queue, "10-created.d")
	receivedDir := filepath.Join(baseDir, queue, "30-received.d")
	startedDir := filepath.Join(baseDir, queue, "50-started.d")
	finishedDir := filepath.Join(baseDir, queue, "70-finished.d")
	logDir := filepath.Join(baseDir, queue, "log")

	for _, dirname := range []string{createdDir, receivedDir, startedDir, finishedDir, logDir} {
		err := os.MkdirAll(dirname, os.FileMode(0755))
		if err != nil {
			return nil, err
		}
	}

	return &FileJobQueue{
		queue:           queue,
		pollingInterval: pollingInterval,

		baseDir:     baseDir,
		createdDir:  createdDir,
		receivedDir: receivedDir,
		startedDir:  startedDir,
		finishedDir: finishedDir,
		logDir:      logDir,
	}, nil
}

// Jobs returns a channel of jobs from the created directory
func (f *FileJobQueue) Jobs(ctx gocontext.Context) (<-chan Job, error) {
	if f.buildJobChan == nil {
		f.buildJobChan = make(chan Job)
		go f.pollInDirForJobs(ctx)
	}
	return f.buildJobChan, nil
}

func (f *FileJobQueue) pollInDirForJobs(ctx gocontext.Context) {
	for {
		f.pollInDirTick(ctx)
		time.Sleep(f.pollingInterval)
	}
}

func (f *FileJobQueue) pollInDirTick(ctx gocontext.Context) {
	logger := context.LoggerFromContext(ctx).WithField("self", "file_job_queue")
	entries, err := ioutil.ReadDir(f.createdDir)
	if err != nil {
		logger.WithField("err", err).Error("input directory read error")
		return
	}

	logger.WithFields(logrus.Fields{
		"entries":        entries,
		"file_job_queue": fmt.Sprintf("%p", f),
	}).Debug("entries")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		buildJob := &fileJob{
			createdFile:     filepath.Join(f.createdDir, entry.Name()),
			payload:         &JobPayload{},
			startAttributes: &backend.StartAttributes{},
		}
		startAttrs := &jobPayloadStartAttrs{Config: &backend.StartAttributes{}}

		fb, err := ioutil.ReadFile(buildJob.createdFile)
		if err != nil {
			logger.WithField("err", err).Error("input file read error")
			continue
		}

		err = json.Unmarshal(fb, buildJob.payload)
		if err != nil {
			logger.WithField("err", err).Error("payload JSON parse error, skipping")
			continue
		}

		err = json.Unmarshal(fb, &startAttrs)
		if err != nil {
			logger.WithField("err", err).Error("start attributes JSON parse error, skipping")
			continue
		}

		buildJob.rawPayload, err = simplejson.NewJson(fb)
		if err != nil {
			logger.WithField("err", err).Error("raw payload JSON parse error, skipping")
			continue
		}

		buildJob.startAttributes = startAttrs.Config
		buildJob.startAttributes.VMConfig = buildJob.payload.VMConfig
		buildJob.startAttributes.VMType = buildJob.payload.VMType
		buildJob.startAttributes.SetDefaults(f.DefaultLanguage, f.DefaultDist, f.DefaultArch, f.DefaultGroup, f.DefaultOS, VMTypeDefault, VMConfigDefault)
		buildJob.receivedFile = filepath.Join(f.receivedDir, entry.Name())
		buildJob.startedFile = filepath.Join(f.startedDir, entry.Name())
		buildJob.finishedFile = filepath.Join(f.finishedDir, entry.Name())
		buildJob.logFile = filepath.Join(f.logDir, strings.Replace(entry.Name(), ".json", ".log", -1))
		buildJob.bytes = fb

		f.buildJobChan <- buildJob
	}
}

// Name returns the name of this queue type, wow!
func (q *FileJobQueue) Name() string {
	return "file"
}

// Cleanup is a no-op
func (f *FileJobQueue) Cleanup() error {
	return nil
}
