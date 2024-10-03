package watch

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"

	"github.com/superproj/onex/pkg/distlock"
	stringsutil "github.com/superproj/onex/pkg/util/strings"
	"github.com/superproj/onex/pkg/watch/initializer"
	"github.com/superproj/onex/pkg/watch/logger/empty"
	"github.com/superproj/onex/pkg/watch/manager"
	"github.com/superproj/onex/pkg/watch/registry"
)

var (
	// Timeout duration for stopping jobs.
	jobStopTimeout = 3 * time.Minute
	// Time extension for lock expiration.
	extendExpiration = 5 * time.Second
	// Default expiration time for locks.
	defaultExpiration = 10 * extendExpiration
)

// Option configures a Watch instance with customizable settings.
type Option func(w *Watch)

// Watch represents a monitoring system that schedules and runs tasks at specified intervals.
type Watch struct {
	// Used to implement dist lock.
	db *sql.DB
	// Job manager to handle scheduling and execution of jobs.
	jm *manager.JobManager
	// Maximum number of concurrent workers for each watcher.
	maxWorkers int64
	// Logger for logging events and errors.
	logger Logger
	// Distributed lock name to be used across instances.
	lockName string
	// Distributed lock instance.
	locker *distlock.Lock
	// healthzPort is the port number for the health check endpoint.
	healthzPort int
	// List of watcher names that should be disabled.
	disableWatchers []string
	// Function for internal initialization of watchers.
	initializer initializer.WatcherInitializer
	// Function for external initialization of watchers.
	externalInitializer initializer.WatcherInitializer
}

// WithInitialize returns an Option function that sets the provided WatcherInitializer
// function to initialize the Watch during its creation.
func WithInitialize(initialize initializer.WatcherInitializer) Option {
	return func(w *Watch) {
		w.externalInitializer = initialize
	}
}

// WithLogger returns an Option function that sets the provided Logger to the Watch for logging purposes.
func WithLogger(logger Logger) Option {
	return func(w *Watch) {
		w.logger = logger
	}
}

// NewWatch creates a new Watch monitoring system with the provided options.
func NewWatch(opts *Options, db *sql.DB, withOptions ...Option) (*Watch, error) {
	logger := empty.NewLogger()

	// Create a new Watch with default settings.
	w := &Watch{
		lockName:        opts.LockName,
		healthzPort:     opts.HealthzPort,
		logger:          logger,
		disableWatchers: opts.DisableWatchers,
		db:              db,
		maxWorkers:      opts.MaxWorkers,
	}

	// Apply user-defined options to the Watch.
	for _, opt := range withOptions {
		opt(w)
	}

	runner := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(w.logger),
		cron.WithChain(cron.DelayIfStillRunning(w.logger), cron.Recover(w.logger)),
	)

	// Initialize the job manager and the watcher initializer.
	w.jm = manager.NewJobManager(manager.WithCron(runner))
	w.initializer = initializer.NewInitializer(w.jm, w.maxWorkers)

	if err := w.addWatchers(); err != nil {
		return nil, err
	}

	return w, nil
}

// addWatchers initializes all registered watchers and adds them as Cron jobs.
// It skips the watchers that are specified in the disableWatchers slice.
func (w *Watch) addWatchers() error {
	for jobName, watcher := range registry.ListWatchers() {
		if stringsutil.StringIn(jobName, w.disableWatchers) {
			continue
		}

		w.initializer.Initialize(watcher)
		if w.externalInitializer != nil {
			w.externalInitializer.Initialize(watcher)
		}

		spec := registry.Every3Seconds
		if obj, ok := watcher.(registry.ISpec); ok {
			spec = obj.Spec()
		}

		if _, err := w.jm.AddJob(jobName, spec, watcher); err != nil {
			w.logger.Error(err, "Failed to add job to the cron", "watcher", jobName)
			return err
		}
	}

	return nil
}

// Start attempts to acquire a distributed lock and starts the Cron job scheduler.
// It retries acquiring the lock until successful.
func (w *Watch) Start(stopCh <-chan struct{}) {
	if w.healthzPort != 0 {
		go w.serveHealthz()
	}

	locker := distlock.NewMySQLLocker(w.db, distlock.WithRefreshInterval(extendExpiration))
	ticker := time.NewTicker(defaultExpiration + (5 * time.Second))
	var err error
	for {
		// Obtain a lock for our given mutex. After this is successful, no one else
		// can obtain the same lock (the same mutex name) until we unlock it.
		w.locker, err = locker.ObtainTimeout(w.lockName, 5)
		if err == nil {
			w.logger.Debug("Successfully acquired lock", "lockName", w.lockName)
			break
		}
		w.logger.Debug("Failed to acquire lock.", "lockName", w.lockName, "err", err)
		<-ticker.C
	}

	w.jm.Start()

	w.logger.Info("Successfully started watch server")
}

// Stop blocks until all jobs are completed and releases the distributed lock.
func (w *Watch) Stop() {
	ctx := w.jm.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(jobStopTimeout):
		w.logger.Error(errors.New("context was not done immediately"), "timeout", jobStopTimeout.String())
	}

	if err := w.locker.Release(); err != nil {
		w.logger.Debug("Failed to release lock", "err", err)
	}

	w.logger.Info("Successfully stopped watch server")
}

// serveHealthz starts the health check server for the Watch instance.
func (w *Watch) serveHealthz() {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthzHandler).Methods(http.MethodGet)

	address := fmt.Sprintf("0.0.0.0:%d", w.healthzPort)

	if err := http.ListenAndServe(address, r); err != nil {
		w.logger.Error(err, "Error serving health check endpoint")
	}

	w.logger.Info("Successfully started health check server", "address", address)
}

// healthzHandler handles the health check requests for the service.
func healthzHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"status": "ok"}`))
}
