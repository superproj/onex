// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options contains flags and options for initializing an nightwatch.
package options

import (
	"math"

	"github.com/spf13/viper"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/superproj/onex/internal/nightwatch"
	"github.com/superproj/onex/internal/pkg/feature"
	kubeutil "github.com/superproj/onex/internal/pkg/util/kube"
	"github.com/superproj/onex/pkg/app"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

const (
	// UserAgent is the userAgent name when starting onex-nightwatch server.
	UserAgent = "onex-nightwatch"
)

var _ app.CliOptions = (*Options)(nil)

// Options contains everything necessary to create and run a nightwatch server.
type Options struct {
	HealthOptions         *genericoptions.HealthOptions  `json:"health" mapstructure:"health"`
	MySQLOptions          *genericoptions.MySQLOptions   `json:"db" mapstructure:"db"`
	RedisOptions          *genericoptions.RedisOptions   `json:"redis" mapstructure:"redis"`
	UserWatcherMaxWorkers int64                          `json:"user-watcher-max-workers" mapstructure:"user-watcher-max-workers"`
	DisableWatchers       []string                       `json:"disable-watchers" mapstructure:"disable-watchers"`
	Metrics               *genericoptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	// Path to kubeconfig file with authorization and master location information.
	Kubeconfig   string          `json:"kubeconfig" mapstructure:"kubeconfig"`
	FeatureGates map[string]bool `json:"feature-gates"`
	Log          *log.Options    `json:"log" mapstructure:"log"`
}

// NewOptions returns initialized Options.
func NewOptions() *Options {
	o := &Options{
		HealthOptions:         genericoptions.NewHealthOptions(),
		MySQLOptions:          genericoptions.NewMySQLOptions(),
		RedisOptions:          genericoptions.NewRedisOptions(),
		UserWatcherMaxWorkers: math.MaxInt64,
		DisableWatchers:       []string{},
		Metrics:               genericoptions.NewMetricsOptions(),
		Log:                   log.NewOptions(),
	}

	return o
}

// Flags returns flags for a specific server by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.HealthOptions.AddFlags(fss.FlagSet("health"))
	o.MySQLOptions.AddFlags(fss.FlagSet("db"))
	o.RedisOptions.AddFlags(fss.FlagSet("redis"))
	o.Metrics.AddFlags(fss.FlagSet("metrics"))
	o.Log.AddFlags(fss.FlagSet("log"))

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.Int64Var(&o.UserWatcherMaxWorkers, "user-watcher-max-workers", o.UserWatcherMaxWorkers, "Specify the maximum concurrency event of user watcher.")
	fs.StringSliceVar(&o.DisableWatchers, "disable-watchers", o.DisableWatchers, "The list of watchers that should be disabled.")
	feature.DefaultMutableFeatureGate.AddFlag(fs)

	return fss
}

// Complete completes all the required options.
func (o *Options) Complete() error {
	if err := viper.Unmarshal(&o); err != nil {
		return err
	}

	if o.UserWatcherMaxWorkers < 1 {
		o.UserWatcherMaxWorkers = math.MaxInt64
	}

	_ = feature.DefaultMutableFeatureGate.SetFromMap(o.FeatureGates)
	return nil
}

// Validate validates all the required options.
func (o *Options) Validate() error {
	errs := []error{}

	errs = append(errs, o.HealthOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// ApplyTo fills up onex-nightwatch config with options.
func (o *Options) ApplyTo(c *nightwatch.Config) error {
	c.MySQLOptions = o.MySQLOptions
	c.RedisOptions = o.RedisOptions
	c.UserWatcherMaxWorkers = o.UserWatcherMaxWorkers
	c.DisableWatchers = o.DisableWatchers
	return nil
}

// Config return an onex-nightwatch config object.
func (o *Options) Config() (*nightwatch.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", o.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeutil.SetDefaultClientOptions(kubeutil.AddUserAgent(kubeconfig, UserAgent))

	client, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	c := &nightwatch.Config{
		Client: client,
	}

	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}
