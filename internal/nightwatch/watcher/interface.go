package watcher

import (
	"github.com/superproj/onex/internal/nightwatch/store"
	aggregatestore "github.com/superproj/onex/internal/pkg/client/store"
	"github.com/superproj/onex/pkg/watch/registry"
)

// WantsAggregateConfig defines a function which sets AggregateConfig for watcher plugins that need it.
type WantsAggregateConfig interface {
	registry.Watcher
	SetAggregateConfig(config *AggregateConfig)
}

// WantsAggregateStore defines a function which sets aggregate store for watcher plugins that need it.
type WantsAggregateStore interface {
	registry.Watcher
	SetAggregateStore(store aggregatestore.Interface)
}

// WantsStore defines a function which sets store for watcher plugins that need it.
type WantsStore interface {
	registry.Watcher
	SetStore(store store.IStore)
}
