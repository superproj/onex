// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	netutils "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	"github.com/superproj/onex/internal/pkg/ports"
	controllermanagerutil "github.com/superproj/onex/internal/pkg/util/controllermanager"
	genericconfigv1beta1 "github.com/superproj/onex/pkg/config/v1beta1"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_OneXControllerManagerConfiguration(obj *OneXControllerManagerConfiguration) {
	RecommendedDefaultGenericControllerManagerConfiguration(&obj.Generic)
	RecommendedDefaultGarbageCollectorControllerConfiguration(&obj.GarbageCollectorController)
	RecommendedDefaultChainControllerConfiguration(&obj.ChainController)
}

// RecommendedDefaultGenericControllerManagerConfiguration defaults a pointer to a
// GenericControllerManagerConfiguration struct. This will set the recommended default
// values, but they may be subject to change between API versions. This function
// is intentionally not registered in the scheme as a "normal" `SetDefaults_Foo`
// function to allow consumers of this type to set whatever defaults for their
// embedded configs. Forcing consumers to use these defaults would be problematic
// as defaulting in the scheme is done as part of the conversion, and there would
// be no easy way to opt-out. Instead, if you want to use this defaulting method
// run it in your wrapper struct of this type in its `SetDefaults_` method.
func RecommendedDefaultGenericControllerManagerConfiguration(obj *GenericControllerManagerConfiguration) {
	genericconfigv1beta1.RecommendedDefaultMySQLConfiguration(&obj.MySQL)
	componentbaseconfigv1alpha1.RecommendedDefaultLeaderElectionConfiguration(&obj.LeaderElection)

	if len(obj.BindAddress) == 0 {
		obj.BindAddress = "0.0.0.0"
	}

	defaultHealthzAddress, defaultMetricsAddress := getDefaultAddresses(obj.BindAddress)
	if obj.HealthzBindAddress == "" {
		obj.HealthzBindAddress = fmt.Sprintf("%s:%v", defaultHealthzAddress, ports.ControllerManagerHealthzPort)
	} else {
		obj.HealthzBindAddress = controllermanagerutil.AppendPortIfNeeded(obj.HealthzBindAddress, ports.ControllerManagerHealthzPort)
	}
	if obj.MetricsBindAddress == "" {
		obj.MetricsBindAddress = fmt.Sprintf("%s:%v", defaultMetricsAddress, ports.ControllerManagerStatusPort)
	} else {
		obj.MetricsBindAddress = controllermanagerutil.AppendPortIfNeeded(obj.MetricsBindAddress, ports.ControllerManagerStatusPort)
	}

	if obj.Parallelism == 0 {
		obj.Parallelism = 16
	}

	if obj.SyncPeriod.Duration == 0 {
		obj.SyncPeriod = metav1.Duration{Duration: 10 * time.Hour}
	}

	// Use lease-based leader election to reduce cost.
	obj.LeaderElection.ResourceLock = "leases"
	if len(obj.LeaderElection.ResourceNamespace) == 0 {
		obj.LeaderElection.ResourceNamespace = OneXControllerManagerDefaultLockObjectNamespace
	}
	if len(obj.LeaderElection.ResourceName) == 0 {
		obj.LeaderElection.ResourceName = OneXControllerManagerDefaultLockObjectName
	}
}

// RecommendedDefaultGarbageCollectorControllerConfiguration defaults a pointer to a
// GarbageCollectorControllerConfiguration struct. This will set the recommended default
// values, but they may be subject to change between API versions. This function
// is intentionally not registered in the scheme as a "normal" `SetDefaults_Foo`
// function to allow consumers of this type to set whatever defaults for their
// embedded configs. Forcing consumers to use these defaults would be problematic
// as defaulting in the scheme is done as part of the conversion, and there would
// be no easy way to opt-out. Instead, if you want to use this defaulting method
// run it in your wrapper struct of this type in its `SetDefaults_` method.
func RecommendedDefaultGarbageCollectorControllerConfiguration(obj *GarbageCollectorControllerConfiguration) {
	if obj.EnableGarbageCollector == nil {
		obj.EnableGarbageCollector = ptr.To(true)
	}
	if obj.ConcurrentGCSyncs == 0 {
		obj.ConcurrentGCSyncs = 20
	}
}

func RecommendedDefaultChainControllerConfiguration(obj *ChainControllerConfiguration) {
	if obj.Image == "" {
		obj.Image = "ccr.ccs.tencentyun.com/superproj/onex-toyblc-amd64:v0.1.0"
	}
}

// getDefaultAddresses returns default address of healthz and metrics server
// based on the given bind address. IPv6 addresses are enclosed in square
// brackets for appending port.
func getDefaultAddresses(bindAddress string) (defaultHealthzAddress, defaultMetricsAddress string) {
	if netutils.ParseIPSloppy(bindAddress).To4() != nil {
		return "0.0.0.0", "127.0.0.1"
	}
	return "[::]", "[::1]"
}
