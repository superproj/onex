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
type Option func(opts *Options)

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
	return func(nw *Watch) {
		nw.externalInitializer = initialize
	}
}

// WithLogger returns an Option function that sets the provided Logger to the Watch for logging purposes.
func WithLogger(logger Logger) Option {
	return func(nw *Watch) {
		nw.logger = logger
	}
}

// NewWatch creates a new Watch monitoring system with the provided options.
func NewWatch(opts *Options, db *sql.DB, withOptions ...Option) (*Watch, error) {
	logger := empty.NewLogger()

	// Create a new Watch with default settings.
	nw := &Watch{
		lockName:        opts.LockName,
		healthzPort: opts.HealthzPort
		logger:          logger,
		disableWatchers: opts.DisableWatchers,
		db:              db,
		maxWorkers:      opts.MaxWorkers,
	}

	// Apply user-defined options to the Watch.
	for _, opt := range withOptions {
		opt(nw)
	}

	runner := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(nw.logger),
		cron.WithChain(cron.DelayIfStillRunning(nw.logger), cron.Recover(nw.logger)),
	)

	// Initialize the job manager and the watcher initializer.
	nw.jm = manager.NewJobManager(manager.WithCron(runner))
	nw.initializer = initializer.NewInitializer(nw.jm, nw.maxWorkers)

	if err := nw.addWatchers(); err != nil {
		return nil, err
	}

	return nw, nil
}

// addWatchers initializes all registered watchers and adds them as Cron jobs.
// It skips the watchers that are specified in the disableWatchers slice.
func (nw *Watch) addWatchers() error {
	for jobName, watcher := range registry.ListWatchers() {
		if stringsutil.StringIn(jobName, nw.disableWatchers) {
			continue
		}

		nw.initializer.Initialize(watcher)
		if nw.externalInitializer != nil {
			nw.externalInitializer.Initialize(watcher)
		}

		spec := registry.Every3Seconds
		if obj, ok := watcher.(registry.ISpec); ok {
			spec = obj.Spec()
		}

		if _, err := nw.jm.AddJob(jobName, spec, watcher); err != nil {
			nw.logger.Error(err, "Failed to add job to the cron", "watcher", jobName)
			return err
		}
	}

	return nil
}

// Start attempts to acquire a distributed lock and starts the Cron job scheduler.
// It retries acquiring the lock until successful.
func (nw *Watch) Start(stopCh <-chan struct{}) {
	if nw.healthzPort != 0 {
		go nw.serveHealthz()
	}

	locker := distlock.NewMySQLLocker(nw.db, distlock.WithRefreshInterval(extendExpiration))
	ticker := time.NewTicker(defaultExpiration + (5 * time.Second))
	var err error
	for {
		// Obtain a lock for our given mutex. After this is successful, no one else
		// can obtain the same lock (the same mutex name) until we unlock it.
		nw.locker, err = locker.ObtainTimeout(nw.lockName, 5)
		if err == nil {
			nw.logger.Debug("Successfully acquired lock", "lockName", nw.lockName)
			break
		}
		nw.logger.Debug("Failed to acquire lock.", "lockName", nw.lockName, "err", err)
		<-ticker.C
	}

	nw.jm.Start()

	nw.logger.Info("Successfully started watch server")
}

// Stop blocks until all jobs are completed and releases the distributed lock.
func (nw *Watch) Stop() {
	ctx := nw.jm.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(jobStopTimeout):
		nw.logger.Error(errors.New("context was not done immediately"), "timeout", jobStopTimeout.String())
	}

	if err := nw.locker.Release(); err != nil {
		nw.logger.Debug("Failed to release lock", "err", err)
	}

	nw.logger.Info("Successfully stopped watch server")
}

// serveHealthz starts the health check server for the Watch instance.
func (nw *Watch) serveHealthz() {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthzHandler).Methods(http.MethodGet)

	address := fmt.Sprintf("0.0.0.0:%d", nw.healthzPort)

	if err := http.ListenAndServe(address, r); err != nil {
		nw.logger.Error(err, "Error serving health check endpoint")
	}

	nw.logger.Info("Successfully started health check server", "address", address)
}

// healthzHandler handles the health check requests for the service.
func healthzHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"status": "ok"}`))
}
