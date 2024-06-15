// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package ports

// In this file, we can see all default port of cluster.
// It's also an important documentation for us. So don't remove them easily.
const (
	// ControllerManagerStatusPort is the default port for the proxy metrics server.
	// May be overridden by a flag at startup.
	ControllerManagerStatusPort = 20251
	// ControllerManagerHealthzPort is the default port for the proxy healthz server.
	// May be overridden by a flag at startup.
	ControllerManagerHealthzPort = 20250

	// KubeletPort is the default port for the kubelet server on each host machine.
	// May be overridden by a flag at startup.
	KubeletPort = 10250
	// KubeletReadOnlyPort exposes basic read-only services from the kubelet.
	// May be overridden by a flag at startup.
	// This is necessary for heapster to collect monitoring stats from the kubelet
	// until heapster can transition to using the SSL endpoint.
	// TODO(roberthbailey): Remove this once we have a better solution for heapster.
	KubeletReadOnlyPort = 10255
	// KubeletHealthzPort exposes a healthz endpoint from the kubelet.
	// May be overridden by a flag at startup.
	KubeletHealthzPort = 10248
	// KubeControllerManagerPort is the default port for the controller manager status server.
	// May be overridden by a flag at startup.
	KubeControllerManagerPort = 10257
	// CloudControllerManagerPort is the default port for the cloud controller manager server.
	// This value may be overridden by a flag at startup.
	CloudControllerManagerPort = 10258
)
