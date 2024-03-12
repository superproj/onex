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

	"github.com/superproj/onex/internal/pkg/ports"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_MinerSetControllerConfiguration sets additional defaults.
func SetDefaults_MinerSetControllerConfiguration(obj *MinerSetControllerConfiguration) {
	if obj.FeatureGates == nil {
		obj.FeatureGates = make(map[string]bool)
	}

	if obj.MetricsBindAddress == "" {
		obj.MetricsBindAddress = fmt.Sprintf("%s:%v", "0.0.0.0", ports.MinerSetControllerStatusPort)
	}
	if obj.HealthzBindAddress == "" {
		obj.HealthzBindAddress = fmt.Sprintf("%s:%v", "0.0.0.0", ports.MinerSetControllerHealthzPort)
	}

	componentbaseconfigv1alpha1.RecommendedDefaultLeaderElectionConfiguration(&obj.LeaderElection)
	// Use lease-based leader election to reduce cost.
	obj.LeaderElection.ResourceLock = "leases"
	if len(obj.LeaderElection.ResourceNamespace) == 0 {
		obj.LeaderElection.ResourceNamespace = MinerSetControllerDefaultLockObjectNamespace
	}
	if len(obj.LeaderElection.ResourceName) == 0 {
		obj.LeaderElection.ResourceName = MinerSetControllerDefaultLockObjectName
	}

	if obj.SyncPeriod.String() == "" {
		obj.SyncPeriod = metav1.Duration{Duration: 10 * time.Minute}
	}

	if obj.Parallelism == 0 {
		obj.Parallelism = 10
	}
}
