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

	"github.com/superproj/onex/internal/pkg/client"
	"github.com/superproj/onex/internal/pkg/feature"
	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/usercenter"
	"github.com/superproj/onex/pkg/app"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

const (
	// UserAgent is the userAgent name when starting onex-gateway server.
	UserAgent = "onex-usercenter"
)

var _ app.CliOptions = (*Options)(nil)

// Options contains state for master/api server.
type Options struct {
	// GenericOptions *genericoptions.Options       `json:"server"   mapstructure:"server"`
	// gRPC options for configuring gRPC related options.
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	// HTTP options for configuring HTTP related options.
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	// TLS options for configuring TLS related options.
	TLSOptions *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	// MySQL options for configuring MySQL database related options.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// Redis options for configuring Redis related options.
	RedisOptions *genericoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	// Etcd options for configuring Etcd related options.
	EtcdOptions *genericoptions.EtcdOptions `json:"etcd" mapstructure:"etcd"`
	// Kafka options for configuring Kafka related options.
	KafkaOptions *genericoptions.KafkaOptions `json:"kafka" mapstructure:"kafka"`
	// Jaeger options for configuring Jaeger related options.
	JaegerOptions *genericoptions.JaegerOptions `json:"jaeger" mapstructure:"jaeger"`
	// Consul options for configuring Consul related options.
	ConsulOptions *genericoptions.ConsulOptions `json:"consul" mapstructure:"consul"`
	// JWT options for configuring JWT related options.
	JWTOptions *genericoptions.JWTOptions `json:"jwt" mapstructure:"jwt"`
	// Metrics options for configuring metric related options.
	Metrics *genericoptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	// TODO: add `mapstructure` tag for FeatureGates
	// A map of string to boolean values representing feature gates for enabling or disabling specific features.
	FeatureGates map[string]bool `json:"feature-gates"`
	// Log options for configuring log related options.
	Log *log.Options `json:"log" mapstructure:"log"`
}

// NewOptions returns initialized Options.
func NewOptions() *Options {
	o := &Options{
		// GenericOptions: genericoptions.NewOptions(),
		GRPCOptions:   genericoptions.NewGRPCOptions(),
		HTTPOptions:   genericoptions.NewHTTPOptions(),
		TLSOptions:    genericoptions.NewTLSOptions(),
		MySQLOptions:  genericoptions.NewMySQLOptions(),
		RedisOptions:  genericoptions.NewRedisOptions(),
		EtcdOptions:   genericoptions.NewEtcdOptions(),
		KafkaOptions:  genericoptions.NewKafkaOptions(),
		JaegerOptions: genericoptions.NewJaegerOptions(),
		ConsulOptions: genericoptions.NewConsulOptions(),
		JWTOptions:    genericoptions.NewJWTOptions(),
		Metrics:       genericoptions.NewMetricsOptions(),
		Log:           log.NewOptions(),
	}

	return o
}

// Flags returns flags for a specific server by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.GRPCOptions.AddFlags(fss.FlagSet("grpc"))
	o.HTTPOptions.AddFlags(fss.FlagSet("http"))
	o.TLSOptions.AddFlags(fss.FlagSet("tls"))
	o.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
	o.RedisOptions.AddFlags(fss.FlagSet("redis"))
	o.EtcdOptions.AddFlags(fss.FlagSet("etcd"))
	o.KafkaOptions.AddFlags(fss.FlagSet("kafka"))
	o.JaegerOptions.AddFlags(fss.FlagSet("jaeger"))
	o.ConsulOptions.AddFlags(fss.FlagSet("consul"))
	o.JWTOptions.AddFlags(fss.FlagSet("jwt"))
	o.Metrics.AddFlags(fss.FlagSet("metrics"))
	o.Log.AddFlags(fss.FlagSet("log"))

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")
	client.AddFlags(fs)
	feature.DefaultMutableFeatureGate.AddFlag(fs)

	return fss
}

// Complete completes all the required options.
func (o *Options) Complete() error {
	if o.JaegerOptions.ServiceName == "" {
		o.JaegerOptions.ServiceName = UserAgent
	}
	o.JWTOptions.Expired = known.RefreshTokenExpire
	_ = feature.DefaultMutableFeatureGate.SetFromMap(o.FeatureGates)
	return nil
}

// Validate validates all the required options.
func (o *Options) Validate() error {
	errs := []error{}

	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	errs = append(errs, o.TLSOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.EtcdOptions.Validate()...)
	errs = append(errs, o.KafkaOptions.Validate()...)
	errs = append(errs, o.JaegerOptions.Validate()...)
	errs = append(errs, o.ConsulOptions.Validate()...)
	errs = append(errs, o.JWTOptions.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// ApplyTo fills up onex-usercenter config with options.
func (o *Options) ApplyTo(c *usercenter.Config) error {
	c.GRPCOptions = o.GRPCOptions
	c.HTTPOptions = o.HTTPOptions
	c.TLSOptions = o.TLSOptions
	c.JWTOptions = o.JWTOptions
	c.MySQLOptions = o.MySQLOptions
	c.RedisOptions = o.RedisOptions
	c.EtcdOptions = o.EtcdOptions
	c.KafkaOptions = o.KafkaOptions
	c.JaegerOptions = o.JaegerOptions
	c.ConsulOptions = o.ConsulOptions
	return nil
}

// Config return an onex-usercenter config object.
func (o *Options) Config() (*usercenter.Config, error) {
	c := &usercenter.Config{}

	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}
