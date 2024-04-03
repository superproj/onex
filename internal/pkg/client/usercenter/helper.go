// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	consulapi "github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"

	genericoptions "github.com/superproj/onex/pkg/options"
)

func newEtcdClient(opts *genericoptions.EtcdOptions) (*etcd.Registry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: opts.Endpoints,
	})
	if err != nil {
		return nil, err
	}
	r := etcd.New(cli)

	return r, nil
}

//nolint:unparam,unused
func newConsulClient(opts *genericoptions.ConsulOptions) (*consul.Registry, error) {
	apiclient, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return consul.New(apiclient), nil
}
