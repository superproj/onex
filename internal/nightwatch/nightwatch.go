// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/nightwatch/watcher"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/all"
	fakeminio "github.com/superproj/onex/internal/pkg/client/minio/fake"
	"github.com/superproj/onex/internal/pkg/known"
	"github.com/superproj/onex/pkg/db"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
	"github.com/superproj/onex/pkg/store/where"
	"github.com/superproj/onex/pkg/watch"
	"github.com/superproj/onex/pkg/watch/logger/onex"
)

// nightWatch represents the night watch server.
type nightWatch struct {
	*watch.Watch
}

// Config holds the configuration for the nightwatch server.
type Config struct {
	HealthOptions     *genericoptions.HealthOptions
	MySQLOptions      *genericoptions.MySQLOptions
	RedisOptions      *genericoptions.RedisOptions
	WatchOptions      *watch.Options
	HTTPOptions       *genericoptions.HTTPOptions
	TLSOptions        *genericoptions.TLSOptions
	DisableRESTServer bool
	// The maximum concurrency event of user watcher.
	UserWatcherMaxWorkers int64
	// The list of watchers that should be disabled.
	Client clientset.Interface
	// Created from MySQLOptions.
	DB *gorm.DB
}

// New creates an asynchronous task instance.
func (c *Config) New(stopCh <-chan struct{}) (*nightWatch, error) {
	where.RegisterTenant("user_id", func(ctx context.Context) string {
		return ctx.(*gin.Context).GetString(known.XUserID)
	})

	var mysqlOptions db.MySQLOptions
	_ = copier.Copy(&mysqlOptions, c.MySQLOptions)
	dbIns, err := db.NewMySQL(&mysqlOptions)
	if err != nil {
		return nil, err
	}
	c.DB = dbIns

	cfg, err := c.CreateWatcherConfig()
	if err != nil {
		return nil, err
	}

	initialize := watcher.NewInitializer(cfg)
	opts := []watch.Option{
		watch.WithInitialize(initialize),
		watch.WithLogger(onex.NewLogger()),
	}

	watchIns, err := watch.NewWatch(c.WatchOptions, db.MustRawDB(c.DB), opts...)
	if err != nil {
		return nil, err
	}

	if !c.DisableRESTServer {
		go NewRESTServer(stopCh, c.HTTPOptions.Addr, c.TLSOptions, c.DB).Start()
	} else {
		go c.HealthOptions.ServeHealthCheck()
	}

	return &nightWatch{Watch: watchIns}, nil
}

// CreateWatcherConfig used to create configuration used by all watcher.
func (c *Config) CreateWatcherConfig() (*watcher.AggregateConfig, error) {
	storeClient, err := wireStore(c.DB)
	if err != nil {
		log.Errorw(err, "Failed to create MySQL client")
		return nil, err
	}

	aggregateStoreClient, err := wireAggregateStore(c.DB)
	if err != nil {
		log.Errorw(err, "Failed to create MySQL client")
		return nil, err
	}

	minioClient, err := fakeminio.NewFakeMinioClient("test-bucket-name")
	if err != nil {
		log.Errorw(err, "Failed to NewMinioClient")
		return nil, err
	}
	return &watcher.AggregateConfig{
		Minio:                 minioClient,
		Store:                 storeClient,
		AggregateStore:        aggregateStoreClient,
		Client:                c.Client,
		UserWatcherMaxWorkers: c.UserWatcherMaxWorkers,
	}, nil
}

// Run keep retrying to acquire lock and then start the Cron job.
func (nw *nightWatch) Run(stopCh <-chan struct{}) {
	nw.Start(stopCh)

	// Wait for stop signal
	<-stopCh
	nw.Stop()
}
