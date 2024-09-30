package onex

import (
	"github.com/superproj/onex/pkg/log"
)

// onexLogger is a logger that implements the Logger interface.
// It uses the log package to log error messages with additional context.
type onexLogger struct{}

// NewLogger creates and returns a new instance of onexLogger.
func NewLogger() *onexLogger {
	return &onexLogger{}
}

// Error logs an error message with the provided context using the log package.
func (l *onexLogger) Error(err error, msg string, kvs ...any) {
	log.Errorw(err, msg, kvs...)
}
