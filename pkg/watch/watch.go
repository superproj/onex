package watch

import (
	"context"
	"errors"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

	stringsutil "github.com/superproj/onex/pkg/util/strings"
	"github.com/superproj/onex/pkg/watch/logger/empty"
)

var (
	defaultLockName   = "watch-lock"
	jobStopTimeout    = 3 * time.Minute
	extendExpiration  = 5 * time.Second
	defaultExpiration = 10 * extendExpiration
)

// Option configures a framework.Registry.
type Option func(nw *Watch)

// Watch represents a monitoring system that schedules and runs tasks at specified intervals.
type Watch struct {
	// Scheduler for running tasks
	runner *cron.Cron
	// Logger for logging
	logger Logger
	// distributed lock
	locker *redsync.Mutex
	// Distributed lock for ensuring only one instance runs at a time
	disableWatchers []string
	// Function for initializing watchers
	initializer WatcherInitializer
	// Distributed lock name for watch server
	lockName string
}

// WithInitialize returns an Option function that sets the provided WatcherInitializer function to initialize the Watch.
func WithInitialize(initialize WatcherInitializer) Option {
	return func(nw *Watch) {
		nw.initializer = initialize
	}
}

// WithLogger returns an Option function that sets the provided Logger to the Watch for logging purposes.
func WithLogger(logger Logger) Option {
	return func(nw *Watch) {
		nw.logger = logger
	}
}

// WithLockName returns an Option function that sets the provided lockName to the Watch.
func WithLockName(lockName string) Option {
	return func(nw *Watch) {
		nw.lockName = lockName
	}
}

// NewWatch creates a new Watch monitoring system with the provided options.
func NewWatch(opts *Options, client *redis.Client, withOptions ...Option) (*Watch, error) {
	logger := empty.NewLogger()
	runner := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(logger),
		cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)),
	)

	// Create a pool with go-redis which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	pool := goredis.NewPool(client)
	lockOpts := []redsync.Option{
		redsync.WithRetryDelay(50 * time.Microsecond),
		redsync.WithTries(3),
		redsync.WithExpiry(defaultExpiration),
	}
	// Create an instance of redisync and obtain a new mutex by using the same name
	// for all instances wanting the same lock.
	locker := redsync.New(pool).NewMutex(defaultLockName, lockOpts...)

	nw := &Watch{runner: runner, locker: locker, disableWatchers: opts.DisableWatchers}
	if err := nw.addWatchers(); err != nil {
		return nil, err
	}

	return nw, nil
}

// addWatchers used to initialize all registered watchers and add the watchers as a Cron job.
func (nw *Watch) addWatchers() error {
	for n, w := range ListWatchers() {
		if stringsutil.StringIn(n, nw.disableWatchers) {
			continue
		}

		if nw.initializer != nil {
			nw.initializer.Initialize(w)
		}

		spec := Every3Seconds
		if obj, ok := w.(ISpec); ok {
			spec = obj.Spec()
		}

		if _, err := nw.runner.AddJob(spec, w); err != nil {
			nw.logger.Error(err, "Failed to add job to the cron", "watcher", n)
			return err
		}
	}

	return nil
}

// Run keep retrying to acquire lock and then start the Cron job.
func (nw *Watch) Start(ctx context.Context) {
	ticker := time.NewTicker(defaultExpiration + (5 * time.Second))
	for {
		// Obtain a lock for our given mutex. After this is successful, no one else
		// can obtain the same lock (the same mutex name) until we unlock it.
		err := nw.locker.LockContext(ctx)
		if err == nil {
			nw.logger.Debug("Successfully acquired lock", "lockName", nw.lockName)
			break
		}
		nw.logger.Debug("Failed to acquire lock.", "lockName", nw.lockName, "err", err)
		<-ticker.C
	}

	go func() {
		ticker := time.NewTicker(extendExpiration)
		for {
			<-ticker.C
			if ok, err := nw.locker.ExtendContext(ctx); !ok || err != nil {
				nw.logger.Debug("Failed to extend mutex", "err", err, "status", ok)
			}
		}
	}()

	nw.runner.Start()
	nw.logger.Info("Successfully started watch server")
}

// Stop used to blocking waits for the job to complete and releases the lock.
func (nw *Watch) Stop() {
	ctx := nw.runner.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(jobStopTimeout):
		nw.logger.Error(errors.New("context was not done immediately"), "timeout", jobStopTimeout.String())
	}

	if ok, err := nw.locker.Unlock(); !ok || err != nil {
		nw.logger.Debug("Failed to unlock", "err", err, "status", ok)
	}

	nw.logger.Info("Successfully stopped watch server")
}
