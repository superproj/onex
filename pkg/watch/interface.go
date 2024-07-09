package watch

// WatcherInitializer is used for initialization of shareable resources between watcher plugins.
// After initialization the resources have to be set separately.
type WatcherInitializer interface {
	Initialize(watcher Watcher)
}
