// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options provides the flags used for the minerset controller.
package options

import (
	"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"

	controllerconfig "github.com/superproj/onex/cmd/onex-minerset-controller/app/config"
	minersetcontrollerconfig "github.com/superproj/onex/internal/controller/minerset/apis/config"
	"github.com/superproj/onex/internal/controller/minerset/apis/config/latest"
	"github.com/superproj/onex/internal/controller/minerset/apis/config/validation"
	clientcmdutil "github.com/superproj/onex/internal/pkg/util/clientcmd"
	kubeutil "github.com/superproj/onex/internal/pkg/util/kube"
)

const (
	// ControllerUserAgent is the userAgent name when starting onex-minerset controller.
	ControllerUserAgent = "onex-minerset-controller"
)

// Options is the main context object for the onex-minerset controller.
type Options struct {
	// ConfigFile is the location of the minerset controller server's configuration file.
	ConfigFile string

	// WriteConfigTo is the path where the default configuration will be written.
	WriteConfigTo string

	// The address of the Kubernetes API server (overrides any value in kubeconfig).
	Master string

	// Path to kubeconfig file with authorization and master location information.
	Kubeconfig string

	Logs *logs.Options

	// config is the minerset controller server's configuration object.
	// The default values.
	config *minersetcontrollerconfig.MinerSetControllerConfiguration
}

// NewOptions creates a new Options with a default config.
func NewOptions() (*Options, error) {
	o := Options{
		Kubeconfig: clientcmdutil.DefaultKubeconfig(),
		Logs:       logs.NewOptions(),
	}

	defaultComponentConfig, err := latest.Default()
	if err != nil {
		return nil, err
	}
	o.config = defaultComponentConfig

	return &o, nil
}

// Complete completes all the required options.
func (o *Options) Complete() error {
	if len(o.ConfigFile) == 0 {
		// If the --config arg is not specified, honor the deprecated as well as leader election CLI args.
		o.ApplyDeprecated()
	} else {
		cfg, err := LoadConfigFromFile(o.ConfigFile)
		if err != nil {
			return err
		}

		o.config = cfg
	}

	return utilfeature.DefaultMutableFeatureGate.SetFromMap(o.config.FeatureGates)
}

// ApplyDeprecated obtains the deprecated CLI args and set them to `o.ComponentConfig` if specified.
func (o *Options) ApplyDeprecated() {
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	// o.Logs.AddFlags(fss.FlagSet("logs"))
	// componentbaseoptions.BindLeaderElectionFlags(&o.config.LeaderElection, fss.FlagSet("leader elect"))
	///o.config.Cloud.AddFlags(fss.FlagSet("cloud"))

	fs := fss.FlagSet("misc")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file.")
	fs.StringVar(&o.WriteConfigTo, "write-config-to", o.WriteConfigTo, "If set, write the default configuration values to this file and exit.")
	fs.StringVar(&o.Master, "master", o.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")

	logsapi.AddFlags(o.Logs, fss.FlagSet("logs"))
	// utilfeature.DefaultMutableFeatureGate.AddFlag(fss.FlagSet("generic"))

	return fss
}

// Validate is used to validate the options and config before launching the controller.
func (o *Options) Validate() error {
	var errs []error

	if err := validation.Validate(o.config).ToAggregate(); err != nil {
		errs = append(errs, err.Errors()...)
	}

	// TODO: validate master and kubeconfig
	if o.config.Parallelism <= 0 {
		errs = append(errs, fmt.Errorf("--parallelism must be greater than or equal to 0"))
	}

	// errs = append(errs, o.Cloud.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// ApplyTo fills up minerset controller config with options.
func (o *Options) ApplyTo(c *controllerconfig.Config) error {
	c.ComponentConfig = o.config
	return nil
}

// Config return a minerset controller config object.
func (o *Options) Config() (*controllerconfig.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
	if err != nil {
		return nil, err
	}

	restConfig := kubeutil.AddUserAgent(kubeconfig, ControllerUserAgent)
	c := &controllerconfig.Config{
		Kubeconfig: kubeutil.SetClientOptionsForController(restConfig),
	}

	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}
