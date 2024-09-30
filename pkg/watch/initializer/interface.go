package initializer

import (
	"github.com/superproj/onex/pkg/watch/manager"
	"github.com/superproj/onex/pkg/watch/registry"
)

// WatcherInitializer is used for initialization of shareable resources between watcher plugins.
// After initialization the resources have to be set separately.
type WatcherInitializer interface {
	Initialize(watcher registry.Watcher)
}

// WantsJobManager defines a function which sets job manager for watcher plugins that need it.
type WantsJobManager interface {
	registry.Watcher
	SetJobManager(jm *manager.JobManager)
}

// WantsMaxWorkers defines a function which sets max workers for watcher plugins that need it.
type WantsMaxWorkers interface {
	registry.Watcher
	SetMaxWorkers(maxWorkers int64)
}
