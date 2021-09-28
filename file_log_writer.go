package worker

import (
	"os"
	"time"

	gocontext "context"
)

type fileLogWriter struct {
	ctx     gocontext.Context
	logFile string
	fd      *os.File

	timer   *time.Timer
	timeout time.Duration
}

func newFileLogWriter(ctx gocontext.Context, logFile string, timeout time.Duration) (LogWriter, error) {
	fd, err := os.Create(logFile)
	if err != nil {
		return nil, err
	}

	return &fileLogWriter{
		ctx:     ctx,
		logFile: logFile,
		fd:      fd,

		timer:   time.NewTimer(time.Hour),
		timeout: timeout,
	}, nil
}

func (w *fileLogWriter) Write(b []byte) (int, error) {
	return w.fd.Write(b)
}

func (w *fileLogWriter) Close() error {
	return w.fd.Close()
}

func (w *fileLogWriter) SetMaxLogLength(n int) {}

func (w *fileLogWriter) SetJobStarted(meta *JobStartedMeta) {}

func (w *fileLogWriter) SetCancelFunc(cancel gocontext.CancelFunc) {}

func (w *fileLogWriter) MaxLengthReached() bool {
	return false
}

func (w *fileLogWriter) Timeout() <-chan time.Time {
	return w.timer.C
}

func (w *fileLogWriter) WriteAndClose(b []byte) (int, error) {
	n, err := w.Write(b)
	if err != nil {
		return n, err
	}

	err = w.Close()
	return n, err
}
