// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/superproj/onex/internal/nightwatch/watcher"
	"github.com/superproj/onex/pkg/db"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/watch"
	onexlogger "github.com/superproj/onex/pkg/watch/logger/onex"

	// trigger init functions in `internal/nightwatch/watcher/all`.
	_ "github.com/superproj/onex/internal/nightwatch/watcher/all"
	genericoptions "github.com/superproj/onex/pkg/options"
)

type nightWatch struct {
	*watch.Watch
}

// Config is the configuration for the nightwatch server.
type Config struct {
	MySQLOptions *genericoptions.MySQLOptions
	RedisOptions *genericoptions.RedisOptions
	WatchOptions *watch.Options
	// The maximum concurrency event of user watcher.
	UserWatcherMaxWorkers int64
	// The list of watchers that should be disabled.
	Client clientset.Interface
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
func (c *Config) CreateWatcherConfig() (*watcher.AggregateConfig, error) {
	var mysqlOptions db.MySQLOptions
	_ = copier.Copy(&mysqlOptions, c.MySQLOptions)
	storeClient, err := wireStoreClient(&mysqlOptions)
	if err != nil {
		log.Errorw(err, "Failed to create MySQL client")
		return nil, err
	}

	return &watcher.AggregateConfig{
		Store:                 storeClient,
		Client:                c.Client,
		UserWatcherMaxWorkers: c.UserWatcherMaxWorkers,
	}, nil
}

// New creates an asynchronous task instance.
func (c *Config) New() (*nightWatch, error) {
	client, err := c.RedisOptions.NewClient()
	if err != nil {
		log.Errorw(err, "Failed to create Redis client")
		return nil, err
	}

	cfg, err := c.CreateWatcherConfig()
	if err != nil {
		return nil, err
	}

	initialize := watcher.NewWatcherInitializer(cfg.Store, cfg.Client, cfg.UserWatcherMaxWorkers)
	opts := []watch.Option{
		watch.WithInitialize(initialize),
		watch.WithLogger(onexlogger.NewLogger()),
		watch.WithLockName("onex-nightwatch-lock"),
	}

	nw, err := watch.NewWatch(c.WatchOptions, client, opts...)
	if err != nil {
		return nil, err
	}

	return &nightWatch{nw}, nil
}

// Run keep retrying to acquire lock and then start the Cron job.
func (nw *nightWatch) Run(stopCh <-chan struct{}) {
	nw.Start(wait.ContextForChannel(stopCh))
	// graceful shutdown
	<-stopCh
	nw.Stop()
}
