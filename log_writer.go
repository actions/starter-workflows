package worker

import (
	"io"
	"time"

	gocontext "context"
)

var (
	// LogWriterTick is how often the buffer should be flushed out and sent to
	// travis-logs.
	LogWriterTick = 500 * time.Millisecond

	// LogChunkSize is a bit of a magic number, calculated like this: The
	// maximum Pusher payload is 10 kB (or 10 KiB, who knows, but let's go with
	// 10 kB since that is smaller). Looking at the travis-logs source, the
	// current message overhead (i.e. the part of the payload that isn't
	// the content of the log part) is 42 bytes + the length of the JSON-
	// encoded ID and the length of the JSON-encoded sequence number. A 64-
	// bit number is up to 20 digits long, so that means (assuming we don't
	// go over 64-bit numbers) the overhead is up to 82 bytes. That means
	// we can send up to 9918 bytes of content. However, the JSON-encoded
	// version of a string can be significantly longer than the raw bytes.
	// Worst case that I could find is "<", which with the Go JSON encoder
	// becomes "\u003c" (i.e. six bytes long). So, given a string of just
	// left angle brackets, the string would become six times as long,
	// meaning that the longest string we can take is 1653. We could still
	// get errors if we go over 64-bit numbers, but I find the likeliness
	// of that happening to both the sequence number, the ID, and us maxing
	// out the worst-case logs to be quite unlikely, so I'm willing to live
	// with that. --Sarah
	LogChunkSize = 1653
)

// JobStartedMeta is metadata that is useful for computing time to first
// log line downstream, and breaking it down into further dimensions.
type JobStartedMeta struct {
	QueuedAt *time.Time `json:"queued_at"`
	Repo     string     `json:"repo"`
	Queue    string     `json:"queue"`
	Infra    string     `json:"infra"`
}

// LogWriter is primarily an io.Writer that will send all bytes to travis-logs
// for processing, and also has some utility methods for timeouts and log length
// limiting. Each LogWriter is tied to a given job, and can be gotten by calling
// the LogWriter() method on a Job.
type LogWriter interface {
	io.WriteCloser
	WriteAndClose([]byte) (int, error)
	Timeout() <-chan time.Time
	SetMaxLogLength(int)
	SetJobStarted(meta *JobStartedMeta)
	SetCancelFunc(gocontext.CancelFunc)
	MaxLengthReached() bool
}
