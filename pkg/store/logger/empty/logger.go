package empty

// emptyLogger is a no-op logger that implements the Logger interface.
// It does not perform any logging operations.
type emptyLogger struct{}

// NewLogger creates and returns a new instance of emptyLogger.
func NewLogger() *emptyLogger {
	return &emptyLogger{} // Return a new instance of emptyLogger
}

// Error is a no-op method that satisfies the Logger interface.
// It does not log any error messages or context.
func (l *emptyLogger) Error(err error, msg string, kvs ...any) {
	// No operation performed for logging errors
}
