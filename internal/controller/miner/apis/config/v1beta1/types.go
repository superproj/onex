// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"

	genericconfigv1beta1 "github.com/superproj/onex/pkg/config/v1beta1"
)

const (
	// MinerControllerDefaultLockObjectNamespace defines default miner controller lock object namespace ("kube-system").
	MinerControllerDefaultLockObjectNamespace string = metav1.NamespaceSystem

	// MinerControllerDefaultLockObjectName defines default miner controller lock object name ("onex-miner-controller").
	MinerControllerDefaultLockObjectName = "onex-miner-controller"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinerControllerConfiguration configures a scheduler.
type MinerControllerConfiguration struct {
	// TypeMeta contains the API version and kind.
	metav1.TypeMeta `json:",inline"`

	// FeatureGates is a map of feature names to bools that enable or disable alpha/experimental features.
	FeatureGates map[string]bool `json:"featureGates,omitempty"`

	// Parallelism defines the amount of parallelism to process miners. Must be greater than 0. Defaults to 16
	Parallelism int32 `json:"parallelism,omitempty"`

	// DryRun tells if the dry run mode is enabled, do not create an actual miner pod,
	// but directly set the miner status to Running.
	// If DryRun is set to true, the DryRun mode will be prioritized.
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// Path to miner provider kubeconfig file with authorization and master location information.
	// +optional
	ProviderKubeconfig string `json:"providerKubeconfig,omitempty"`

	// Create miner pod in the cluster where miner controller is located.
	// +optional
	InCluster bool `json:"inCluster,omitempty"`

	// SyncPeriod determines the minimum frequency at which watched resources are
	// reconciled. A lower period will correct entropy more quickly, but reduce
	// responsiveness to change if there are many watched resources. Change this
	// value only if you know what you are doing. Defaults to 10 hours if unset.
	SyncPeriod metav1.Duration `json:"syncPeriod,omitempty"`

	// Label value that the controller watches to reconcile cloud miner objects.
	// Label key is always %s. If unspecified, the controller watches for allcluster-api objects.
	WatchFilterValue string `json:"watchFilterValue,omitempty"`

	// leaderElection defines the configuration of leader election client.
	LeaderElection componentbaseconfigv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`

	// Namespace that the controller watches to reconcile onex-apiserver objects.
	// If unspecified, the controller watches for onex-apiserver objects across all namespaces
	Namespace string `json:"namespace,omitempty"`

	// MetricsBindAddress is the IP address and port for the metrics server to serve on,
	// defaulting to 127.0.0.1:20249 (set to 0.0.0.0 for all interfaces)
	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`

	// HealthzBindAddress is the IP address and port for the health check server to serve on,
	// defaulting to 0.0.0.0:20250
	HealthzBindAddress string `json:"healthzBindAddress,omitempty"`

	// Types specifies the configuration of the cloud mining machine.
	Types map[string]MinerProfile `json:"types,omitempty"`

	// Redis defines the configuration of redis client.
	Redis genericconfigv1beta1.RedisConfiguration `json:"redis,omitempty"`

	// Logs *logs.Options `json:"logs,omitempty"`
	// Metrics            *metrics.Options
	// Cloud options
	// Cloud *cloud.CloudOptions `json:"cloud,omitempty"`
}

type MinerProfile struct {
	CPU              resource.Quantity `json:"cpu,omitempty"`
	Memory           resource.Quantity `json:"memory,omitempty"`
	MiningDifficulty int               `json:"miningDifficulty,omitempty"`
}
