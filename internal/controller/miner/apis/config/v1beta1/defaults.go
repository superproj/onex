// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"

	"github.com/superproj/onex/internal/pkg/ports"
	genericconfigv1beta1 "github.com/superproj/onex/pkg/config/v1beta1"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_MinerControllerConfiguration sets additional defaults.
func SetDefaults_MinerControllerConfiguration(obj *MinerControllerConfiguration) {
	if obj.FeatureGates == nil {
		obj.FeatureGates = make(map[string]bool)
	}

	if obj.MetricsBindAddress == "" {
		obj.MetricsBindAddress = fmt.Sprintf("%s:%v", "0.0.0.0", ports.MinerControllerStatusPort)
	}
	if obj.HealthzBindAddress == "" {
		obj.HealthzBindAddress = fmt.Sprintf("%s:%v", "0.0.0.0", ports.MinerControllerHealthzPort)
	}
	if obj.ProviderKubeconfig == "" {
		// Here KUBECONFIG environment variable will not be used, KUBECONFIG is reserved for onex-apiserver.
		obj.ProviderKubeconfig = clientcmd.RecommendedHomeFile
	}

	componentbaseconfigv1alpha1.RecommendedDefaultLeaderElectionConfiguration(&obj.LeaderElection)
	// Use lease-based leader election to reduce cost.
	obj.LeaderElection.ResourceLock = "leases"
	if len(obj.LeaderElection.ResourceNamespace) == 0 {
		obj.LeaderElection.ResourceNamespace = MinerControllerDefaultLockObjectNamespace
	}
	if len(obj.LeaderElection.ResourceName) == 0 {
		obj.LeaderElection.ResourceName = MinerControllerDefaultLockObjectName
	}

	if obj.SyncPeriod.String() == "" {
		obj.SyncPeriod = metav1.Duration{Duration: 10 * time.Minute}
	}

	if obj.Parallelism == 0 {
		obj.Parallelism = 10
	}

	if len(obj.Types) == 0 {
		obj.Types = map[string]MinerProfile{
			"S1.SMALL1": {
				CPU:              resource.MustParse("50m"),
				Memory:           resource.MustParse("128Mi"),
				MiningDifficulty: 7,
			},
			"S1.SMALL2": {
				CPU:              resource.MustParse("100m"),
				Memory:           resource.MustParse("256Mi"),
				MiningDifficulty: 5,
			},
			"M1.MEDIUM1": {
				CPU:              resource.MustParse("150m"),
				Memory:           resource.MustParse("512Mi"),
				MiningDifficulty: 3,
			},
			"M1.MEDIUM2": {
				CPU:              resource.MustParse("250m"),
				Memory:           resource.MustParse("1024Mi"),
				MiningDifficulty: 1,
			},
		}
	}

	genericconfigv1beta1.RecommendedDefaultRedisConfiguration(&obj.Redis)
}
