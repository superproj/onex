// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options contains flags and options for initializing an apiserver
package options

import (
	cliflag "k8s.io/component-base/cli/flag"
	"net"

	controlplaneoptions "github.com/superproj/onex/internal/controlplane/apiserver/options"
)

const defaultEtcdPathPrefix = "/registry/onex.io"

// ServerRunOptions contains state for master/api server.
type ServerRunOptions struct {
	*controlplaneoptions.Options

	Extra
}

type Extra struct {
	MasterCount int
	// In the future, perhaps an "onexlet" will be added, similar to the "kubelet".
	// OnexletConfig onexletclient.OnexletClientConfig
	APIServerServiceIP     net.IP
	EndpointReconcilerType string
}

// NewServerRunOptions returns a new ServerRunOptions.
func NewServerRunOptions() *ServerRunOptions {
	o := &ServerRunOptions{
		Options: controlplaneoptions.NewOptions(),
		Extra: Extra{
			MasterCount: 1,
		},
	}

	return o
}

func (o ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	o.Options.AddFlags(&fss)

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")

	fs.IntVar(&o.MasterCount, "apiserver-count", o.MasterCount,
		"The number of apiservers running in the cluster, must be a positive number. (In use when --endpoint-reconciler-type=master-count is enabled.)")
	fs.MarkDeprecated("apiserver-count", "apiserver-count is deprecated and will be removed in a future version.")

	return fss
}
