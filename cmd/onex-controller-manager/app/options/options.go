// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options provides the flags used for the controller manager.
package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/metrics"
	"k8s.io/kubernetes/pkg/controller/garbagecollector"

	controllermanagerconfig "github.com/superproj/onex/cmd/onex-controller-manager/app/config"
	"github.com/superproj/onex/cmd/onex-controller-manager/names"
	"github.com/superproj/onex/internal/controller/apis/config/latest"
	clientcmdutil "github.com/superproj/onex/internal/pkg/util/clientcmd"
	kubeutil "github.com/superproj/onex/internal/pkg/util/kube"
	genericconfig "github.com/superproj/onex/pkg/config"
	genericconfigoptions "github.com/superproj/onex/pkg/config/options"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

const (
	// ControllerManagerUserAgent is the userAgent name when starting onex-controller managers.
	ControllerManagerUserAgent = "onex-controller-manager"
)

// Options is the main context object for the onex-controller manager.
type Options struct {
	Generic                    *genericconfigoptions.GenericControllerManagerConfigurationOptions
	GarbageCollectorController *genericconfigoptions.GarbageCollectorControllerOptions
	MySQL                      *genericconfigoptions.MySQLOptions
	ChainController            *ChainControllerOptions
	//NamespaceController        *NamespaceControllerOptions

	// ConfigFile is the location of the miner controller server's configuration file.
	ConfigFile string

	// WriteConfigTo is the path where the default configuration will be written.
	WriteConfigTo string

	// The address of the Kubernetes API server (overrides any value in kubeconfig).
	Master string
	// Path to kubeconfig file with authorization and master location information.
	Kubeconfig string
	Metrics    *metrics.Options
	Logs       *logs.Options

	// config is the onex controller manager server's configuration object.
	// The default values.
	//config *ctrlmgrconfig.OneXControllerManagerConfiguration
}

// NewOptions creates a new Options with a default config.
func NewOptions() (*Options, error) {
	componentConfig, err := latest.Default()
	if err != nil {
		return nil, err
	}

	o := Options{
		Generic:                    genericconfigoptions.NewGenericControllerManagerConfigurationOptions(&componentConfig.Generic),
		GarbageCollectorController: genericconfigoptions.NewGarbageCollectorControllerOptions(&componentConfig.GarbageCollectorController),
		MySQL:                      genericconfigoptions.NewMySQLOptions(&componentConfig.MySQL),
		ChainController:            NewChainControllerOptions(&componentConfig.ChainController),
		Kubeconfig:                 clientcmdutil.DefaultKubeconfig(),
		Metrics:                    metrics.NewOptions(),
		Logs:                       logs.NewOptions(),
		//config:     componentConfig,
	}

	gcIgnoredResources := make([]genericconfig.GroupResource, 0, len(garbagecollector.DefaultIgnoredResources()))
	for r := range garbagecollector.DefaultIgnoredResources() {
		gcIgnoredResources = append(gcIgnoredResources, genericconfig.GroupResource{Group: r.Group, Resource: r.Resource})
	}
	o.GarbageCollectorController.GCIgnoredResources = gcIgnoredResources
	o.Generic.LeaderElection.ResourceName = "onex-controller-manager"
	o.Generic.LeaderElection.ResourceNamespace = metav1.NamespaceSystem

	return &o, nil
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags(allControllers []string, disabledControllers []string, controllerAliases map[string]string) cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	o.Generic.AddFlags(&fss, allControllers, disabledControllers, controllerAliases)
	o.GarbageCollectorController.AddFlags(fss.FlagSet(names.GarbageCollectorController))
	o.MySQL.AddFlags(fss.FlagSet("mysql"))
	o.ChainController.AddFlags(fss.FlagSet(names.ChainController))

	o.Metrics.AddFlags(fss.FlagSet("metrics"))
	logsapi.AddFlags(o.Logs, fss.FlagSet("logs"))

	fs := fss.FlagSet("misc")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file.")
	fs.StringVar(&o.WriteConfigTo, "write-config-to", o.WriteConfigTo, "If set, write the default configuration values to this file and exit.")
	fs.StringVar(&o.Master, "master", o.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")

	utilfeature.DefaultMutableFeatureGate.AddFlag(fss.FlagSet("generic"))

	return fss
}

func (o *Options) Complete() error {
	return nil
}

// ApplyTo fills up onex controller manager config with options.
func (o *Options) ApplyTo(c *controllermanagerconfig.Config, allControllers []string, disabledControllers []string, controllerAliases map[string]string) error {
	if err := o.Generic.ApplyTo(&c.ComponentConfig.Generic, allControllers, disabledControllers, controllerAliases); err != nil {
		return err
	}
	if err := o.GarbageCollectorController.ApplyTo(&c.ComponentConfig.GarbageCollectorController); err != nil {
		return err
	}

	if err := o.ChainController.ApplyTo(&c.ComponentConfig.ChainController); err != nil {
		return err
	}

	o.Metrics.Apply()

	return nil
}

// Validate is used to validate the options and config before launching the controller.
func (o *Options) Validate(allControllers []string, disabledControllers []string, controllerAliases map[string]string) error {
	var errs []error

	errs = append(errs, o.Generic.Validate(allControllers, disabledControllers, controllerAliases)...)
	errs = append(errs, o.GarbageCollectorController.Validate()...)
	errs = append(errs, o.ChainController.Validate()...)

	// TODO: validate component config, master and kubeconfig

	return utilerrors.NewAggregate(errs)
}

// Config return a controller manager config objective.
func (o Options) Config(allControllers []string, disabledControllers []string, controllerAliases map[string]string) (*controllermanagerconfig.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeconfig.DisableCompression = true

	restConfig := kubeutil.AddUserAgent(kubeconfig, ControllerManagerUserAgent)
	client, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	c := &controllermanagerconfig.Config{
		Kubeconfig: kubeutil.SetClientOptionsForController(restConfig),
		Client:     client,
	}

	if err := o.ApplyTo(c, allControllers, disabledControllers, controllerAliases); err != nil {
		return nil, err
	}

	return c, nil
}
