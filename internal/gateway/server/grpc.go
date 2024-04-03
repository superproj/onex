// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/superproj/onex/internal/gateway/service"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

// NewGRPCServer creates a new gRPC server with middleware options, and registers the UserCenterService.
func NewGRPCServer(c *Config, gw *service.GatewayService, middlewares []middleware.Middleware) *grpc.Server {
	opts := []grpc.ServerOption{
		// grpc.WithDiscovery(nil),
		// grpc.WithEndpoint("discovery:///matrix.creation.service.grpc"),
		// Define the middleware chain with variable options.
		grpc.Middleware(middlewares...),
	}
	if c.GRPC.Network != "" {
		opts = append(opts, grpc.Network(c.GRPC.Network))
	}
	if c.GRPC.Timeout != 0 {
		opts = append(opts, grpc.Timeout(c.GRPC.Timeout))
	}
	if c.GRPC.Addr != "" {
		opts = append(opts, grpc.Address(c.GRPC.Addr))
	}
	if c.TLS.UseTLS {
		opts = append(opts, grpc.TLSConfig(c.TLS.MustTLSConfig()))
	}

	// Create a new gRPC server with the middleware options.
	srv := grpc.NewServer(opts...)
	v1.RegisterGatewayServer(srv, gw)
	return srv
}
