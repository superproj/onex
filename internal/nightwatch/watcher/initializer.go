package watcher

import (
	"github.com/superproj/onex/internal/pkg/client/store"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/watch"
)

// WatcherInitializer is used for initialization of the onex specific watcher plugins.
type WatcherInitializer struct {
	// The purpose of nightwatch is to handle asynchronous tasks on the onex platform
	// in a unified manner, so a store aggregation type is needed here.
	store store.Interface

	// Client is the client for onex-apiserver.
	client clientset.Interface

	// Then maximum concurrency event of user watcher.
	userWatcherMaxWorkers int64
}

var _ watch.WatcherInitializer = &WatcherInitializer{}

func NewWatcherInitializer(store store.Interface, client clientset.Interface, maxWorkers int64) *WatcherInitializer {
	return &WatcherInitializer{
		store:                 store,
		client:                client,
		userWatcherMaxWorkers: maxWorkers,
	}
}

func (w *WatcherInitializer) Initialize(wc watch.Watcher) {
	// We can set a specific configuration as needed, as shown in the example below.
	// However, for convenience, I directly assign all configurations to each watcher,
	// allowing the watcher to choose which ones to use.
	if wants, ok := wc.(WantsStore); ok {
		wants.SetStore(w.store)
	}

	if wants, ok := wc.(WantsAggregateConfig); ok {
		wants.SetAggregateConfig(&AggregateConfig{
			Store:                 w.store,
			Client:                w.client,
			UserWatcherMaxWorkers: w.userWatcherMaxWorkers,
		})
	}
}
