package worker

import (
	"fmt"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

var (
	severityMap = map[logrus.Level]raven.Severity{
		logrus.DebugLevel: raven.DEBUG,
		logrus.InfoLevel:  raven.INFO,
		logrus.WarnLevel:  raven.WARNING,
		logrus.ErrorLevel: raven.ERROR,
		logrus.FatalLevel: raven.FATAL,
		logrus.PanicLevel: raven.FATAL,
	}
)

// SentryHook delivers logs to a sentry server
type SentryHook struct {
	Timeout time.Duration
	client  *raven.Client
	levels  []logrus.Level
}

// NewSentryHook creates a hook to be added to an instance of logger and
// initializes the raven client. This method sets the timeout to 100
// milliseconds.
func NewSentryHook(DSN string, levels []logrus.Level) (*SentryHook, error) {
	client, err := raven.New(DSN)
	if err != nil {
		return nil, err
	}
	return &SentryHook{100 * time.Millisecond, client, levels}, nil
}

// Fire is called when an event should be sent to sentry
func (hook *SentryHook) Fire(entry *logrus.Entry) error {
	defer func() {
		if r := recover(); r != nil {
			entry.Logger.WithField("panic", r).Error("paniced when trying to send log to sentry")
		}
	}()

	packet := &raven.Packet{
		Message:   entry.Message,
		Timestamp: raven.Timestamp(entry.Time),
		Level:     severityMap[entry.Level],
		Platform:  "go",
	}

	if serverName, ok := entry.Data["server_name"]; ok {
		packet.ServerName = serverName.(string)
		delete(entry.Data, "server_name")
	}
	packet.Extra = map[string]interface{}(entry.Data)

	if errMaybe, ok := packet.Extra["err"]; ok {
		if err, ok := errMaybe.(error); ok {
			packet.Extra["err"] = err.Error()
		}
	}

	packet.Interfaces = append(packet.Interfaces, raven.NewStacktrace(4, 3, []string{"github.com/travis-ci/worker"}))

	_, errCh := hook.client.Capture(packet, nil)
	if hook.Timeout != 0 {
		timeoutCh := time.After(hook.Timeout)
		select {
		case err := <-errCh:
			return err
		case <-timeoutCh:
			return fmt.Errorf("no response from sentry server in %s", hook.Timeout)
		}
	}
	return nil
}

// Levels returns the available logging levels.
func (hook *SentryHook) Levels() []logrus.Level {
	return hook.levels
}
