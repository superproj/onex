// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	consulapi "github.com/hashicorp/consul/api"
	helloworld "go.opencensus.io/examples/grpc/proto"

	"github.com/superproj/onex/internal/pkg/client"
)

// ProviderSet is the usercenter providers.
var ProviderSet = wire.NewSet(New, wire.Bind(new(Interface), new(*clientImpl)))

// Interface is an interface that presents a subset of the usercenter API.
type Interface interface {
	GetSecret(ctx context.Context, rq *GetSecretRequest) (*GetSecretResponse, error)
}

// clientImpl is an implementation of Interface.
type clientImpl struct {
	client helloworld.GreeterClient
}

var (
	once      sync.Once
	rpcclient helloworld.GreeterClient
)

// New creates a new client to work with usercenter services.
func New(opts *Options) (*clientImpl, error) {
	var err error
	once.Do(func() {
		fn := func(*Options) (helloworld.GreeterClient, error) {
			cliopts := make([]grpc.ClientOption, 0)
			if client.IsDiscoveryEndpoint(opts.Server) {
				apiclient, err := consulapi.NewClient(consulapi.DefaultConfig())
				if err != nil {
					return nil, err
				}

				cliopts = append(cliopts, grpc.WithDiscovery(consul.New(apiclient)))
			}

			// new grpc client
			cliopts = append(cliopts, grpc.WithEndpoint(opts.Server))
			conn, err := grpc.DialInsecure(context.Background(), cliopts...)
			if err != nil {
				return nil, err
			}

			rpcclient := helloworld.NewGreeterClient(conn)

			return rpcclient, nil
		}

		rpcclient, err = fn(opts)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil, err
	}

	client := &clientImpl{client: rpcclient}
	return client, nil
}
