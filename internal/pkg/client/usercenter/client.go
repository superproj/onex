// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	grpcx "google.golang.org/grpc"

	"github.com/superproj/onex/internal/pkg/client"
	"github.com/superproj/onex/internal/pkg/middleware/auth"
	"github.com/superproj/onex/internal/pkg/middleware/tracing"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// ProviderSet is the usercenter providers.
var ProviderSet = wire.NewSet(NewUserCenter, wire.Bind(new(Interface), new(*impl)), wire.Bind(new(auth.AuthProvider), new(*impl)))

var (
	once sync.Once
	cli  *impl
)

// Interface is an interface that presents a subset of the usercenter API.
type Interface interface {
	Auth(ctx context.Context, token string, obj, act string) (string, bool, error)
}

// impl is an implementation of Interface.
type impl struct {
	client v1.UserCenterClient
}

type Impl = impl

// NewUserCenter creates a new client to work with usercenter services.
func NewUserCenter(opts *UserCenterOptions, etcdOpts *genericoptions.EtcdOptions) *impl {
	once.Do(func() {
		fn := func(*UserCenterOptions) (v1.UserCenterClient, error) {
			cliopts := []grpc.ClientOption{
				grpc.WithMiddleware(tracing.Client()),
				grpc.WithEndpoint(opts.Server),
				grpc.WithTimeout(opts.Timeout),
				// for tracing remote ip recording
				grpc.WithOptions(grpcx.WithStatsHandler(&tracing.ClientHandler{})),
			}
			if client.IsDiscoveryEndpoint(opts.Server) {
				client, err := newEtcdClient(etcdOpts)
				if err != nil {
					return nil, err
				}
				cliopts = append(cliopts, grpc.WithDiscovery(client))
			}

			conn, err := grpc.DialInsecure(context.Background(), cliopts...)
			if err != nil {
				return nil, err
			}

			rpcclient := v1.NewUserCenterClient(conn)

			return rpcclient, nil
		}

		rpcclient, err := fn(opts)
		if err != nil {
			panic(err)
		}

		cli = &impl{rpcclient}
	})

	return cli
}

// GetClient returns the globally initialized client.
func GetClient() *impl {
	return cli
}

// Auth implements the Interface interface.
func (i *impl) Auth(ctx context.Context, token string, obj, act string) (userID string, allowed bool, err error) {
	rq := &v1.AuthRequest{Token: token, Obj: obj, Act: act}
	resp, err := i.client.Auth(ctx, rq)
	if err != nil {
		return "", false, err
	}

	return resp.UserID, resp.Allowed, nil
}
