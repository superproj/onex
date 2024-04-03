// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package config

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfig "k8s.io/component-base/config"

	genericconfig "github.com/superproj/onex/pkg/config"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinerControllerConfiguration configures a scheduler.
type MinerControllerConfiguration struct {
	// TypeMeta contains the API version and kind.
	metav1.TypeMeta

	// FeatureGates is a map of feature names to bools that enable or disable alpha/experimental features.
	FeatureGates map[string]bool

	// Parallelism defines the amount of parallelism to process miners. Must be greater than 0. Defaults to 16
	Parallelism int32

	// DryRun tells if the dry run mode is enabled, do not create an actual miner pod,
	// but directly set the miner status to Running.
	DryRun bool

	// Path to miner provider kubeconfig file with authorization and master location information.
	ProviderKubeconfig string

	// Create miner pod in the cluster where miner controller is located.
	InCluster bool

	// SyncPeriod determines the minimum frequency at which watched resources are
	// reconciled. A lower period will correct entropy more quickly, but reduce
	// responsiveness to change if there are many watched resources. Change this
	// value only if you know what you are doing. Defaults to 10 hours if unset.
	SyncPeriod metav1.Duration

	// Label value that the controller watches to reconcile cloud miner objects.
	// Label key is always %s. If unspecified, the controller watches for allcluster-api objects.
	WatchFilterValue string

	// leaderElection defines the configuration of leader election client.
	LeaderElection componentbaseconfig.LeaderElectionConfiguration

	// Namespace that the controller watches to reconcile onex-apiserver objects.
	// If unspecified, the controller watches for onex-apiserver objects across all namespaces
	Namespace string

	// MetricsBindAddress is the IP address and port for the metrics server to serve on,
	// defaulting to 127.0.0.1:20249 (set to 0.0.0.0 for all interfaces)
	MetricsBindAddress string

	// HealthzBindAddress is the IP address and port for the health check server to serve on,
	// defaulting to 0.0.0.0:20250
	HealthzBindAddress string

	// Types specifies the configuration of the cloud mining machine.
	Types map[string]MinerProfile

	// Redis defines the configuration of redis client.
	Redis genericconfig.RedisConfiguration

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
