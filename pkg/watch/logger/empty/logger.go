package empty

// cronLogger implement the cron.Logger interface.
type cronLogger struct{}

// NewLogger returns a cron logger.
func NewLogger() *cronLogger {
	return &cronLogger{}
}

// Debug logs an debug condition.
func (l *cronLogger) Debug(msg string, keysAndValues ...any) {}

// Info logs routine messages about cron's operation.
func (l *cronLogger) Info(msg string, keysAndValues ...any) {}

// Error logs an error condition.
func (l *cronLogger) Error(err error, msg string, keysAndValues ...any) {}
