package onex

import (
	"github.com/superproj/onex/pkg/log"
)

// cronLogger implement the cron.Logger interface.
type cronLogger struct{}

// NewLogger returns a cron logger.
func NewLogger() *cronLogger {
	return &cronLogger{}
}

// Debug logs routine messages about cron's operation.
func (l *cronLogger) Debug(msg string, kvs ...any) {
	log.Debugw(msg, kvs...)
}

// Info logs routine messages about cron's operation.
func (l *cronLogger) Info(msg string, kvs ...any) {
	log.Infow(msg, kvs...)
}

// Error logs an error condition.
func (l *cronLogger) Error(err error, msg string, kvs ...any) {
	log.Errorw(err, msg, kvs...)
}
