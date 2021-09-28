package worker

import (
	gocontext "context"
	"time"
)

type LogWriterFactory interface {
	LogWriter(gocontext.Context, time.Duration, Job) (LogWriter, error)
	Cleanup() error
}
