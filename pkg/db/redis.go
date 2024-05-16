// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions defines options for redis database.
type RedisOptions struct {
	Addr         string
	Username     string
	Password     string
	Database     int
	MaxRetries   int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	PoolSize     int
}

// NewRedis create a new redis db instance with the given options.
func NewRedis(opts *RedisOptions) (*redis.Client, error) {
	options := &redis.Options{
		Addr:         opts.Addr,
		Username:     opts.Username,
		Password:     opts.Password,
		DB:           opts.Database,
		MaxRetries:   opts.MaxRetries,
		MinIdleConns: opts.MinIdleConns,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolTimeout:  opts.PoolTimeout,
		PoolSize:     opts.PoolSize,
	}

	rdb := redis.NewClient(options)

	// check redis if is ok
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}
