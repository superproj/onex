// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"context"
	"errors"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/jinzhu/copier"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/superproj/onex/internal/nightwatch/watcher"
	"github.com/superproj/onex/pkg/db"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/log"

	// trigger init functions in `internal/nightwatch/watcher/all`.
	_ "github.com/superproj/onex/internal/nightwatch/watcher/all"
	genericoptions "github.com/superproj/onex/pkg/options"
	stringsutil "github.com/superproj/onex/pkg/util/strings"
)

var (
	lockName          = "onex-nightwatch-lock"
	jobStopTimeout    = 3 * time.Minute
	extendExpiration  = 5 * time.Second
	defaultExpiration = 10 * extendExpiration
)

type nightWatch struct {
	runner *cron.Cron
	// distributed lock
	locker          *redsync.Mutex
	config          *watcher.Config
	disableWatchers []string
}

// Config is the configuration for the nightwatch server.
type Config struct {
	MySQLOptions *genericoptions.MySQLOptions
	RedisOptions *genericoptions.RedisOptions
	// The maximum concurrency event of user watcher.
	UserWatcherMaxWorkers int64
	// The list of watchers that should be disabled.
	DisableWatchers []string
	Client          clientset.Interface
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() *CompletedConfig {
	return &CompletedConfig{c}
}

// CreateWatcherConfig used to create configuration used by all watcher.
func (c *Config) CreateWatcherConfig() (*watcher.Config, error) {
	var mysqlOptions db.MySQLOptions
	_ = copier.Copy(&mysqlOptions, c.MySQLOptions)
	storeClient, err := wireStoreClient(&mysqlOptions)
	if err != nil {
		log.Errorw(err, "Failed to create MySQL client")
		return nil, err
	}

	return &watcher.Config{
		Store:                 storeClient,
		Client:                c.Client,
		UserWatcherMaxWorkers: c.UserWatcherMaxWorkers,
	}, nil
}

// New creates an asynchronous task instance.
func (c *Config) New() (*nightWatch, error) {
	// Create a pool with go-redis which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client, err := c.RedisOptions.NewClient()
	if err != nil {
		log.Errorw(err, "Failed to create Redis client")
		return nil, err
	}

	logger := newCronLogger()
	runner := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(logger),
		cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)),
	)

	pool := goredis.NewPool(client)
	lockOpts := []redsync.Option{
		redsync.WithRetryDelay(50 * time.Microsecond),
		redsync.WithTries(3),
		redsync.WithExpiry(defaultExpiration),
	}
	// Create an instance of redisync and obtain a new mutex by using the same name
	// for all instances wanting the same lock.
	locker := redsync.New(pool).NewMutex(lockName, lockOpts...)

	cfg, err := c.CreateWatcherConfig()
	if err != nil {
		return nil, err
	}

	nw := &nightWatch{runner: runner, locker: locker, config: cfg, disableWatchers: c.DisableWatchers}
	if err := nw.addWatchers(); err != nil {
		return nil, err
	}

	return nw, nil
}

// addWatchers used to initialize all registered watchers and add the watchers as a Cron job.
func (nw *nightWatch) addWatchers() error {
	for n, w := range watcher.ListWatchers() {
		if stringsutil.StringIn(n, nw.disableWatchers) {
			continue
		}

		if err := w.Init(context.Background(), nw.config); err != nil {
			log.Errorw(err, "Failed to construct watcher", "watcher", n)
			return err
		}

		spec := watcher.Every3Seconds
		if obj, ok := w.(watcher.ISpec); ok {
			spec = obj.Spec()
		}

		if _, err := nw.runner.AddJob(spec, w); err != nil {
			log.Errorw(err, "Failed to add job to the cron", "watcher", n)
			return err
		}
	}

	return nil
}

// Run keep retrying to acquire lock and then start the Cron job.
func (nw *nightWatch) Run(stopCh <-chan struct{}) {
	ctx := wait.ContextForChannel(stopCh)

	ticker := time.NewTicker(defaultExpiration + (5 * time.Second))
	for {
		// Obtain a lock for our given mutex. After this is successful, no one else
		// can obtain the same lock (the same mutex name) until we unlock it.
		err := nw.locker.LockContext(ctx)
		if err == nil {
			log.Debugw("Successfully acquired lock", "lockName", lockName)
			break
		}
		log.Debugw("Failed to acquire lock.", "lockName", lockName, "err", err)
		<-ticker.C
	}

	go func() {
		ticker := time.NewTicker(extendExpiration)
		for {
			<-ticker.C
			if ok, err := nw.locker.ExtendContext(ctx); !ok || err != nil {
				log.Debugw("Failed to extend mutex", "err", err, "status", ok)
			}
		}
	}()

	nw.runner.Start()
	log.Infof("Successfully started nightwatch server")

	<-stopCh

	nw.stop()
}

// stop used to blocking waits for the job to complete and releases the lock.
func (nw *nightWatch) stop() {
	ctx := nw.runner.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(jobStopTimeout):
		log.Errorw(errors.New("context was not done immediately"), "timeout", jobStopTimeout.String())
	}

	if ok, err := nw.locker.Unlock(); !ok || err != nil {
		log.Debugw("Failed to unlock", "err", err, "status", ok)
	}

	log.Infof("Successfully stopped nightwatch server")
}
