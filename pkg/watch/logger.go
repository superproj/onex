package watch

import (
	"github.com/robfig/cron/v3"
)

// Logger is an interface that extends the cron.Logger interface with an additional Debug method.
type Logger interface {
	cron.Logger
	Debug(msg string, keysAndValues ...interface{})
}
