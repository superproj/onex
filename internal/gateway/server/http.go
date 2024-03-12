// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/superproj/onex/internal/gateway/service"
	"github.com/superproj/onex/internal/pkg/pprof"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

// NewHTTPServer creates a new HTTP server with middleware and handler chain.
func NewHTTPServer(c *Config, gw *service.GatewayService, middlewares []middleware.Middleware) *http.Server {
	opts := []http.ServerOption{
		// http.WithDiscovery(nil),
		// http.WithEndpoint("discovery:///matrix.creation.service.grpc"),
		// Define the middleware chain with variable options.
		http.Middleware(middlewares...),
		// Add filter options to the middleware chain.
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{
				"X-Requested-With",
				"Content-Type",
				"Authorization",
				"X-Idempotent-ID",
			}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
	if c.HTTP.Network != "" {
		opts = append(opts, http.Network(c.HTTP.Network))
	}
	if c.HTTP.Timeout != 0 {
		opts = append(opts, http.Timeout(c.HTTP.Timeout))
	}
	if c.HTTP.Addr != "" {
		opts = append(opts, http.Address(c.HTTP.Addr))
	}
	if c.TLS.UseTLS {
		opts = append(opts, http.TLSConfig(c.TLS.MustTLSConfig()))
	}

	// Create and return the server instance.
	srv := http.NewServer(opts...)
	h := openapiv2.NewHandler()
	srv.HandlePrefix("/openapi/", h)
	srv.Handle("/metrics", promhttp.Handler())
	srv.Handle("", pprof.NewHandler())

	v1.RegisterGatewayHTTPServer(srv, gw)
	return srv
}
