// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"time"

	"github.com/redis/go-redis/extra/rediscensus/v9"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/pflag"

	"github.com/superproj/onex/pkg/db"
)

var _ IOptions = (*RedisOptions)(nil)

// RedisOptions defines options for redis cluster.
type RedisOptions struct {
	Addr         string        `json:"addr" mapstructure:"addr"`
	Username     string        `json:"username" mapstructure:"username"`
	Password     string        `json:"password" mapstructure:"password"`
	Database     int           `json:"database" mapstructure:"database"`
	MaxRetries   int           `json:"max-retries" mapstructure:"max-retries"`
	MinIdleConns int           `json:"min-idle-conns" mapstructure:"min-idle-conns"`
	DialTimeout  time.Duration `json:"dial-timeout" mapstructure:"dial-timeout"`
	ReadTimeout  time.Duration `json:"read-timeout" mapstructure:"read-timeout"`
	WriteTimeout time.Duration `json:"write-timeout" mapstructure:"write-timeout"`
	PoolTimeout  time.Duration `json:"pool-time" mapstructure:"pool-time"`
	PoolSize     int           `json:"pool-size" mapstructure:"pool-size"`
	// tracing switch
	EnableTrace bool `json:"enable-trace" mapstructure:"enable-trace"`
}

// NewRedisOptions create a `zero` value instance.
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Addr:         "127.0.0.1:6379",
		Username:     "",
		Password:     "",
		Database:     0,
		MaxRetries:   3,
		MinIdleConns: 0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		EnableTrace:  false,
	}
}

// Validate verifies flags passed to RedisOptions.
func (o *RedisOptions) Validate() []error {
	errs := []error{}

	if o.WriteTimeout == 0 {
		o.WriteTimeout = o.ReadTimeout
	}

	if o.PoolTimeout == 0 {
		o.PoolTimeout = o.ReadTimeout + 1*time.Second
	}

	return errs
}

// AddFlags adds flags related to redis storage for a specific APIServer to the specified FlagSet.
func (o *RedisOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, "redis.addr", o.Addr, "Address of your Redis server(ip:port).")
	fs.StringVar(&o.Username, "redis.username", o.Username, "Username for access to redis service.")
	fs.StringVar(&o.Password, "redis.password", o.Password, "Optional auth password for redis db.")
	fs.IntVar(&o.Database, "redis.database", o.Database, "Database to be selected after connecting to the server.")
	fs.IntVar(&o.MaxRetries, "redis.max-retries", o.MaxRetries, "Maximum number of retries before giving up.")
	fs.IntVar(&o.MinIdleConns, "redis.min-idle-conns", o.MinIdleConns, ""+
		"Minimum number of idle connections which is useful when establishing new connection is slow.")
	fs.DurationVar(&o.DialTimeout, "redis.dial-timeout", o.DialTimeout, "Dial timeout for establishing new connections.")
	fs.DurationVar(&o.ReadTimeout, "redis.read-timeout", o.ReadTimeout, "Timeout for socket reads.")
	fs.DurationVar(&o.WriteTimeout, "redis.write-timeout", o.WriteTimeout, "Timeout for socket writes.")
	fs.DurationVar(&o.PoolTimeout, "redis.pool-timeout", o.PoolTimeout, ""+
		"Amount of time client waits for connection if all connections are busy before returning an error.")
	fs.IntVar(&o.PoolSize, "redis.pool-size", o.PoolSize, "Maximum number of socket connections.")
	fs.BoolVar(&o.EnableTrace, "redis.enable-trace", o.EnableTrace, "Redis hook tracing (using open telemetry).")
}

func (o *RedisOptions) NewClient() (*redis.Client, error) {
	opts := &db.RedisOptions{
		Addr:         o.Addr,
		Username:     o.Username,
		Password:     o.Password,
		Database:     o.Database,
		MaxRetries:   o.MaxRetries,
		MinIdleConns: o.MinIdleConns,
		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
		PoolSize:     o.PoolSize,
		PoolTimeout:  o.PoolTimeout,
	}

	rdb, err := db.NewRedis(opts)
	if err != nil {
		return nil, err
	}

	// hook tracing (using open telemetry)
	if o.EnableTrace {
		rdb.AddHook(rediscensus.NewTracingHook())
	}

	return rdb, nil
}
