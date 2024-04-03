// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options contains flags and options for initializing an apiserver
package options

import (
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/superproj/onex/internal/cacheserver"
	"github.com/superproj/onex/pkg/app"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

const (
	// UserAgent is the userAgent name when starting onex-cacheserver server.
	UserAgent = "onex-cacheserver"
)

var _ app.CliOptions = (*Options)(nil)

// Options contains state for master/api server.
type Options struct {
	DisableCache  bool                           `json:"disable-cache" mapstructure:"disable-cache"`
	GRPCOptions   *genericoptions.GRPCOptions    `json:"grpc" mapstructure:"grpc"`
	TLSOptions    *genericoptions.TLSOptions     `json:"tls" mapstructure:"tls"`
	RedisOptions  *genericoptions.RedisOptions   `json:"redis" mapstructure:"redis"`
	MySQLOptions  *genericoptions.MySQLOptions   `json:"mysql" mapstructure:"mysql"`
	JaegerOptions *genericoptions.JaegerOptions  `json:"jaeger" mapstructure:"jaeger"`
	Metrics       *genericoptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	Log           *log.Options                   `json:"log" mapstructure:"log"`
}

// NewOptions returns initialized Options.
func NewOptions() *Options {
	o := &Options{
		DisableCache:  false,
		GRPCOptions:   genericoptions.NewGRPCOptions(),
		TLSOptions:    genericoptions.NewTLSOptions(),
		RedisOptions:  genericoptions.NewRedisOptions(),
		MySQLOptions:  genericoptions.NewMySQLOptions(),
		JaegerOptions: genericoptions.NewJaegerOptions(),
		Metrics:       genericoptions.NewMetricsOptions(),
		Log:           log.NewOptions(),
	}

	return o
}

// Flags returns flags for a specific server by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.GRPCOptions.AddFlags(fss.FlagSet("grpc"))
	o.TLSOptions.AddFlags(fss.FlagSet("tls"))
	o.RedisOptions.AddFlags(fss.FlagSet("redis"))
	o.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
	o.JaegerOptions.AddFlags(fss.FlagSet("jaeger"))
	o.Metrics.AddFlags(fss.FlagSet("metrics"))
	o.Log.AddFlags(fss.FlagSet("log"))

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")
	fs.BoolVar(&o.DisableCache, "disable-cache", o.DisableCache, "Used to indicate whether to disable local memory cache.")

	return fss
}

// Complete completes all the required options.
func (o *Options) Complete() error {
	if o.JaegerOptions.ServiceName == "" {
		o.JaegerOptions.ServiceName = UserAgent
	}
	return nil
}

// Validate validates all the required options.
func (o *Options) Validate() error {
	errs := []error{}

	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.TLSOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.JaegerOptions.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// ApplyTo fills up onex-cacheserver config with options.
func (o *Options) ApplyTo(c *cacheserver.Config) error {
	c.DisableCache = o.DisableCache
	c.GRPCOptions = o.GRPCOptions
	c.TLSOptions = o.TLSOptions
	c.RedisOptions = o.RedisOptions
	c.MySQLOptions = o.MySQLOptions
	c.JaegerOptions = o.JaegerOptions

	return nil
}

// Config return a onex-cacheserver config object.
func (o *Options) Config() (*cacheserver.Config, error) {
	c := &cacheserver.Config{}

	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}
