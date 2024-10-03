package watcher

import (
	"github.com/superproj/onex/pkg/watch/initializer"
	"github.com/superproj/onex/pkg/watch/registry"
)

// WatcherInitializer is used for initialization of the onex specific watcher plugins.
type WatcherInitializer struct {
	*AggregateConfig
}

// Ensure that WatcherInitializer implements the initializer.WatcherInitializer interface.
var _ initializer.WatcherInitializer = &WatcherInitializer{}

// NewInitializer creates and returns a new WatcherInitializer instance.
func NewInitializer(aggregate *AggregateConfig) *WatcherInitializer {
	return &WatcherInitializer{AggregateConfig: aggregate}
}

// Initialize configures the provided watcher by injecting dependencies
// such as the Store and AggregateConfig when supported by the watcher.
func (w *WatcherInitializer) Initialize(wc registry.Watcher) {
	// We can set a specific configuration as needed, as shown in the example below.
	// However, for convenience, I directly assign all configurations to each watcher,
	// allowing the watcher to choose which ones to use.
	if wants, ok := wc.(WantsAggregateStore); ok {
		wants.SetAggregateStore(w.AggregateStore)
	}

	if wants, ok := wc.(WantsStore); ok {
		wants.SetStore(w.Store)
	}

	if wants, ok := wc.(WantsAggregateConfig); ok {
		wants.SetAggregateConfig(w.AggregateConfig)
	}
}
