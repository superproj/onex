package initializer

import (
	"github.com/superproj/onex/pkg/watch/manager"
	"github.com/superproj/onex/pkg/watch/registry"
)

// watcherInitializer is responsible for initializing specific watcher plugins.
type watcherInitializer struct {
	jm *manager.JobManager
	// Specify the maximum concurrency event of user watcher.
	maxWorkers int64
}

// Ensure that watcherInitializer implements the WatcherInitializer interface.
var _ WatcherInitializer = (*watcherInitializer)(nil)

// NewInitializer creates and returns a new watcherInitializer instance.
func NewInitializer(jm *manager.JobManager, maxWorkers int64) *watcherInitializer {
	return &watcherInitializer{jm: jm, maxWorkers: maxWorkers}
}

// Initialize configures the provided watcher by setting up the necessary dependencies
// such as the JobManager and maximum workers.
func (i *watcherInitializer) Initialize(wc registry.Watcher) {
	// We can set a specific configuration as needed, as shown in the example below.
	// However, for convenience, I directly assign all configurations to each watcher,
	// allowing the watcher to choose which ones to use.
	if wants, ok := wc.(WantsJobManager); ok {
		wants.SetJobManager(i.jm)
	}

	if wants, ok := wc.(WantsMaxWorkers); ok {
		wants.SetMaxWorkers(i.maxWorkers)
	}
}
