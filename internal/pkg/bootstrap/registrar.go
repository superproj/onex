// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package bootstrap

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	consulapi "github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"

	genericoptions "github.com/superproj/onex/pkg/options"
)

func NewEtcdRegistrar(opts *genericoptions.EtcdOptions) registry.Registrar {
	if opts == nil {
		panic("etcd registrar options must be set.")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   opts.Endpoints,
		DialTimeout: opts.DialTimeout,
		TLS:         opts.TLSOptions.MustTLSConfig(),
		Username:    opts.Username,
		Password:    opts.Password,
	})
	if err != nil {
		panic(err)
	}
	r := etcd.New(client)
	return r
}

func NewConsulRegistrar(opts *genericoptions.ConsulOptions) registry.Registrar {
	if opts == nil {
		panic("consul registrar options must be set.")
	}

	c := consulapi.DefaultConfig()
	c.Address = opts.Addr
	c.Scheme = opts.Scheme
	cli, err := consulapi.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}
