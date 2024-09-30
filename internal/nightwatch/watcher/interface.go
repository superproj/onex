package watcher

import (
	"github.com/superproj/onex/internal/pkg/client/store"
	"github.com/superproj/onex/pkg/watch/registry"
)

// WantsAggregateConfig defines a function which sets AggregateConfig for watcher plugins that need it.
type WantsAggregateConfig interface {
	registry.Watcher
	SetAggregateConfig(config *AggregateConfig)
}

// WantsStore defines a function which sets store for watcher plugins that need it.
type WantsStore interface {
	registry.Watcher
	SetStore(store store.Interface)
}
