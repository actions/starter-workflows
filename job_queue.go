package worker

import (
	gocontext "context"
)

// JobQueue is the minimal interface needed by a ProcessorPool
type JobQueue interface {
	Jobs(gocontext.Context) (<-chan Job, error)
	Name() string
	Cleanup() error
}
