// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config/options"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	cmconfig "github.com/superproj/onex/internal/controller/apis/config"
	cmoptions "github.com/superproj/onex/pkg/config/options"
)

// GenericControllerManagerConfigurationOptions holds the options which are generic.
type GenericControllerManagerConfigurationOptions struct {
	*cmconfig.GenericControllerManagerConfiguration
}

// AddFlags adds flags related to ChainController for controller manager to the specified FlagSet.
func (o *GenericControllerManagerConfigurationOptions) AddFlags(fss *cliflag.NamedFlagSets) {
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
}
