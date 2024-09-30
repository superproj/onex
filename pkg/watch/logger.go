package watch

import (
	"github.com/robfig/cron/v3"
)

// Logger is an interface that extends the cron.Logger interface.
// It provides an additional method for logging debug messages.
//
// This interface allows implementing custom logging behaviors
// while maintaining compatibility with the cron package's logging
// requirements. Implementations of this interface should provide
// the capability to log standard messages, as defined in the
// cron.Logger interface, as well as a method to log debug
// messages with optional key-value pairs for structured logging.
type Logger interface {
	cron.Logger
	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, keysAndValues ...interface{})
}
