// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cacheserver

import (
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/util/wait"

	// "github.com/superproj/onex/internal/cacheserver/biz"
	// "github.com/superproj/onex/internal/cacheserver/service"
	// "github.com/superproj/onex/internal/cacheserver/store".
	"github.com/superproj/onex/pkg/cache"
	redisstore "github.com/superproj/onex/pkg/cache/store/redis"
	"github.com/superproj/onex/pkg/db"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// Config represents the configuration of the service.
type Config struct {
	DisableCache  bool
	GRPCOptions   *genericoptions.GRPCOptions
	TLSOptions    *genericoptions.TLSOptions
	RedisOptions  *genericoptions.RedisOptions
	MySQLOptions  *genericoptions.MySQLOptions
	JaegerOptions *genericoptions.JaegerOptions
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() completedConfig {
	return completedConfig{cfg}
}

type completedConfig struct {
	*Config
}

// CacheServer represents the cache server.
type CacheServer struct {
	grpcsrv Server
	config  completedConfig
}

// New returns a new instance of Server from the given config.
func (c completedConfig) New(stopCh <-chan struct{}) (*CacheServer, error) {
	if err := c.JaegerOptions.SetTracerProvider(); err != nil {
		return nil, err
	}

	rds, err := c.RedisOptions.NewClient()
	if err != nil {
		return nil, err
	}

	redisStore := redisstore.NewRedis(rds)
	l2cache := cache.New[*any.Any](redisStore)
	l2mgr := cache.NewL2[*any.Any](l2cache, cache.L2WithDisableCache(c.DisableCache))
	l2mgr.Wait(wait.ContextForChannel(stopCh))

	var dbOptions db.MySQLOptions
	_ = copier.Copy(&dbOptions, c.MySQLOptions)

	srv, err := wireServer(&dbOptions, l2mgr, c.DisableCache)
	if err != nil {
		return nil, err
	}

	grpcsrv, err := NewGRPCServer(c.GRPCOptions, c.TLSOptions, srv)
	if err != nil {
		return nil, err
	}

	return &CacheServer{grpcsrv: grpcsrv, config: c}, nil
}

// Run run the cache server.
func (s *CacheServer) Run(stopCh <-chan struct{}) error {
	go s.grpcsrv.RunOrDie()

	<-stopCh

	// The most gracefully way is to shutdown the dependent service first,
	// and then shutdown the depended service.
	s.grpcsrv.GracefulStop()

	return nil
}
