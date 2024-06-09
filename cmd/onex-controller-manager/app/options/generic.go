// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config/options"
	utilflag "k8s.io/kubernetes/pkg/util/flag"
	"strings"

	cmconfig "github.com/superproj/onex/internal/controller/apis/config"
	cmoptions "github.com/superproj/onex/pkg/config/options"
)

// GenericControllerManagerConfigurationOptions holds the options which are generic.
type GenericControllerManagerConfigurationOptions struct {
	*cmconfig.GenericControllerManagerConfiguration
}

func NewGenericControllerManagerConfigurationOptions(cfg *cmconfig.GenericControllerManagerConfiguration) *GenericControllerManagerConfigurationOptions {
	return &GenericControllerManagerConfigurationOptions{
		GenericControllerManagerConfiguration: cfg,
	}
}

// AddFlags adds flags related to ChainController for controller manager to the specified FlagSet.
func (o *GenericControllerManagerConfigurationOptions) AddFlags(
	fss *cliflag.NamedFlagSets,
	allControllers []string,
	disabledControllers []string,
	controllerAliasesmap map[string]string,
) {
	if o == nil {
		return
	}

	cmoptions.BindMySQLFlags(&o.MySQL, fss.FlagSet("mysql"))
	options.BindLeaderElectionFlags(&o.LeaderElection, fss.FlagSet("leader election"))

	genericfs := fss.FlagSet("generic")
	genericfs.StringVar(&o.Namespace, "namespace", o.Namespace, "Namespace that the controller watches to reconcile onex-apiserver objects. "+
		"This parameter is ignored if a config file is specified by --config.")
	// genericfs.StringVar(&o.MetricsBindAddress, "metrics-bind-address", o.MetricsBindAddress, "The IP address with port for the metrics "+
	// "server to serve on (set to '0.0.0.0:10249' for all IPv4 interfaces and '[::]:10249' for all IPv6 interfaces). Set empty to disable. "+
	// "This parameter is ignored if a config file is specified by --config.")
	// genericfs.StringVar(&o.HealthzBindAddress, "healthz-bind-address", o.HealthzBindAddress, "The IP address with port for the health check "+
	// "server to serve on (set to '0.0.0.0:10256' for all IPv4 interfaces and '[::]:10256' for all IPv6 interfaces). Set empty to disable. "+
	// "This parameter is ignored if a config file is specified by --config.")
	genericfs.Var(&utilflag.IPVar{Val: &o.BindAddress}, "bind-address", "The IP address for the proxy server to serve on (set to '0.0.0.0' for all IPv4 interfaces and '::' for all   IPv6 interfaces). This parameter is ignored if a config file is specified by --config.")
	genericfs.Var(&utilflag.IPPortVar{Val: &o.MetricsBindAddress}, "metrics-bind-address", "The IP address with port for the metrics "+
		"server to serve on (set to '0.0.0.0:10249' for all IPv4 interfaces and '[::]:10249' for all IPv6 interfaces). Set empty to disable. "+
		"This parameter is ignored if a config file is specified by --config.")
	genericfs.Var(&utilflag.IPPortVar{Val: &o.HealthzBindAddress}, "healthz-bind-address", "The IP address with port for the health check "+
		"server to serve on (set to '0.0.0.0:10256'  for all IPv4 interfaces and '[::]:10256' for all IPv6 interfaces). Set empty to disable. "+
		"This parameter is ignored if a config file is specified by --config.")
	genericfs.Int32Var(&o.Parallelism, "parallelism", o.Parallelism, "The amount of parallelism to process. Must be greater than 0. Defaults to 16."+
		"This parameter is ignored if a config file is specified by --config.")
	genericfs.DurationVar(&o.SyncPeriod.Duration, "sync-period", o.SyncPeriod.Duration, "The minimum interval at which watched resources are reconciled."+
		"This parameter is ignored if a config file is specified by --config.")
	genericfs.StringVar(&o.WatchFilterValue, "watch-filter-value", o.WatchFilterValue, "The label value used to filter events prior to reconciliation."+
		"This parameter is ignored if a config file is specified by --config.")
	genericfs.StringSliceVar(&o.Controllers, "controllers", o.Controllers, fmt.Sprintf(""+
		"A list of controllers to enable. '*' enables all on-by-default controllers, 'foo' enables the controller "+
		"named 'foo', '-foo' disables the controller named 'foo'.\nAll controllers: %s\nDisabled-by-default controllers: %s",
		strings.Join(allControllers, ", "), strings.Join(disabledControllers, ", ")))
}

func (o *GenericControllerManagerConfigurationOptions) ApplyTo(
	cfg *cmconfig.GenericControllerManagerConfiguration,
	allControllers []string,
	disabledControllers []string,
	controllerAliases map[string]string,
) error {
	*cfg = *o.GenericControllerManagerConfiguration

	// copy controller names and replace aliases with canonical names
	cfg.Controllers = make([]string, len(o.Controllers))
	for i, initialName := range o.Controllers {
		initialNameWithoutPrefix := strings.TrimPrefix(initialName, "-")
		controllerName := initialNameWithoutPrefix
		if canonicalName, ok := controllerAliases[controllerName]; ok {
			controllerName = canonicalName
		}
		if strings.HasPrefix(initialName, "-") {
			controllerName = fmt.Sprintf("-%s", controllerName)
		}
		cfg.Controllers[i] = controllerName
	}

	return nil
}

// Validate checks validation of GenericControllerManagerConfigurationOptions.
func (o *GenericControllerManagerConfigurationOptions) Validate(allControllers []string, disabledControllers []string, controllerAliases map[string]string) []error {
	if o == nil {
		return nil
	}

	errs := []error{}

	allControllersSet := sets.NewString(allControllers...)
	for _, initialName := range o.Controllers {
		if initialName == "*" {
			continue
		}
		initialNameWithoutPrefix := strings.TrimPrefix(initialName, "-")
		controllerName := initialNameWithoutPrefix
		if canonicalName, ok := controllerAliases[controllerName]; ok {
			controllerName = canonicalName
		}
		if !allControllersSet.Has(controllerName) {
			errs = append(errs, fmt.Errorf("%q is not in the list of known controllers", initialNameWithoutPrefix))
		}
	}

	return errs
}
