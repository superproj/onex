package store

// Logger defines an interface for logging errors with contextual information.
type Logger interface {
	// Error logs an error message with the associated context.
	Error(err error, message string, kvs ...any)
}
