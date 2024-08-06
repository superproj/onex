// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

const (
	// ControllerManagerDefaultLockObjectNamespace defines default onex controller manager lock object namespace ("kube-system").
	ControllerManagerDefaultLockObjectNamespace string = metav1.NamespaceSystem

	// ControllerManagerDefaultLockObjectName defines default onex controller manager lock object name ("onex-controller-manager").
	ControllerManagerDefaultLockObjectName = "controller-manager"
)

// GenericControllerManagerConfiguration holds configuration for a generic controller-manager.
type GenericControllerManagerConfiguration struct {
	// leaderElection defines the configuration of leader election client.
	LeaderElection componentbaseconfigv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`

	// Namespace that the controller watches to reconcile onex-apiserver objects.
	Namespace string `json:"namespace,omitempty"`

	// bindAddress is the IP address for the proxy server to serve on (set to 0.0.0.0
	// for all interfaces)
	BindAddress string `json:"bindAddress,omitempty"`

	// MetricsBindAddress is the IP address and port for the metrics server to serve on,
	// defaulting to 127.0.0.1:20249 (set to 0.0.0.0 for all interfaces)
	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`

	// HealthzBindAddress is the IP address and port for the health check server to serve on,
	// defaulting to 0.0.0.0:20250
	HealthzBindAddress string `json:"healthzBindAddress,omitempty"`

	// PprofBindAddress is the TCP address that the controller should bind to
	// for serving pprof.
	// It can be set to "" or "0" to disable the pprof serving.
	// Since pprof may contain sensitive information, make sure to protect it
	// before exposing it to public.
	PprofBindAddress string `json:"pprofBindAddress,omitempty"`

	// Parallelism defines the amount of parallelism to process miners. Must be greater than 0. Defaults to 16
	Parallelism int32 `json:"parallelism,omitempty"`

	// SyncPeriod determines the minimum frequency at which watched resources are
	// reconciled. A lower period will correct entropy more quickly, but reduce
	// responsiveness to change if there are many watched resources. Change this
	// value only if you know what you are doing. Defaults to 10 hours if unset.
	SyncPeriod metav1.Duration `json:"syncPeriod,omitempty"`

	// Label value that the controller watches to reconcile cloud miner objects
	WatchFilterValue string `json:"watchFilterValue,omitempty"`

	// Controllers is the list of controllers to enable or disable
	// '*' means "all enabled by default controllers"
	// 'foo' means "enable 'foo'"
	// '-foo' means "disable 'foo'"
	// first item for a particular name wins
	Controllers []string `json:"controllers,omitempty"`
}

// GroupResource describes an group resource.
type GroupResource struct {
	// group is the group portion of the GroupResource.
	Group string `json:"group,omitempty"`
	// resource is the resource portion of the GroupResource.
	Resource string `json:"resource,omitempty"`
}

// GarbageCollectorControllerConfiguration contains elements describing GarbageCollectorController.
type GarbageCollectorControllerConfiguration struct {
	// enables the generic garbage collector. MUST be synced with the
	// corresponding flag of the kube-apiserver. WARNING: the generic garbage
	// collector is an alpha feature.
	EnableGarbageCollector *bool `json:"enableGarbageCollector,omitempty"`
	// concurrentGCSyncs is the number of garbage collector workers that are
	// allowed to sync concurrently.
	ConcurrentGCSyncs int32 `json:"concurrentGCSyncs,omitempty"`
	// gcIgnoredResources is the list of GroupResources that garbage collection should ignore.
	GCIgnoredResources []GroupResource `json:"gcIgnoredResources,omitempty"`
}

// MySQLConfiguration defines the configuration of mysql
// clients for components that can run with mysql database.
type MySQLConfiguration struct {
	// MySQL service host address. If left blank, the following related mysql options will be ignored.
	Host string `json:"host"`
	// Username for access to mysql service.
	Username string `json:"username"`
	// Password for access to mysql, should be used pair with password.
	Password string `json:"password"`
	// Database name for the server to use.
	Database string `json:"database"`
	// Maximum idle connections allowed to connect to mysql.
	MaxIdleConnections int32 `json:"maxIdleConnections"`
	// Maximum open connections allowed to connect to mysql.
	MaxOpenConnections int32 `json:"maxOpenConnections"`
	// Maximum connection life time allowed to connect to mysql.
	MaxConnectionLifeTime metav1.Duration `json:"maxConnectionLifeTime"`
}

// RedisConfiguration defines the configuration of redis
// clients for components that can run with redis key-value database.
type RedisConfiguration struct {
	// Address of your Redis server(ip:port).
	Addr string `json:"addr"`
	// Username for access to redis service.
	Username string `json:"username"`
	// Optional auth password for Redis db.
	Password string `json:"password"`
	// Database to be selected after connecting to the server.
	Database int `json:"database"`
	// Maximum number of retries before giving up.
	MaxRetries int `json:"maxRetries"`
	// Timeout when connecting to redis service.
	Timeout metav1.Duration `json:"timeout"`
}
