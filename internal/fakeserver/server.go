// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package fakeserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/superproj/onex/pkg/api/fakeserver/v1"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

type Server interface {
	RunOrDie()
	GracefulStop()
}

type HTTPServer struct {
	srv         *http.Server
	httpOptions *genericoptions.HTTPOptions
	tlsOptions  *genericoptions.TLSOptions
}

type GRPCServer struct {
	srv  *grpc.Server
	opts *genericoptions.GRPCOptions
}

func NewHTTPServer(
	httpOptions *genericoptions.HTTPOptions,
	tlsOptions *genericoptions.TLSOptions,
	grpcOptions *genericoptions.GRPCOptions,
) (*HTTPServer, error) {
	var tlsConfig *tls.Config
	var err error
	if tlsOptions != nil && tlsOptions.UseTLS {
		tlsConfig, err = tlsOptions.TLSConfig()
		if err != nil {
			return nil, err
		}
	}

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	dialOptions := []grpc.DialOption{grpc.WithBlock()}
	if tlsOptions != nil && tlsOptions.UseTLS {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcOptions.Timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, grpcOptions.Addr, dialOptions...)
	if err != nil {
		return nil, err
	}

	gwmux := runtime.NewServeMux()
	if err := pb.RegisterFakeServerHandler(context.Background(), gwmux, conn); err != nil {
		return nil, err
	}

	gwsrv := &http.Server{Addr: httpOptions.Addr, Handler: gwmux}
	if tlsOptions != nil && tlsOptions.UseTLS {
		gwsrv.TLSConfig = tlsConfig
	}

	return &HTTPServer{srv: gwsrv, httpOptions: httpOptions, tlsOptions: tlsOptions}, nil
}

func (s *HTTPServer) RunOrDie() {
	log.Infof("Start to listening the incoming %s requests on %s", scheme(s.tlsOptions), s.httpOptions.Addr)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalw("Failed to start http(s) server", "err", err)
	}
}

func (s *HTTPServer) GracefulStop() {
	// 创建 ctx 用于通知服务器 goroutine, 它有 10 秒时间完成当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Errorw(err, "Failed to gracefully shutdown http(s) server")
	}
}

func NewGRPCServer(
	grpcOptions *genericoptions.GRPCOptions,
	tlsOptions *genericoptions.TLSOptions,
	srv pb.FakeServerServer,
) (*GRPCServer, error) {
	dialOptions := []grpc.ServerOption{}
	if tlsOptions != nil && tlsOptions.UseTLS {
		tlsConfig, err := tlsOptions.TLSConfig()
		if err != nil {
			return nil, err
		}

		dialOptions = append(dialOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcsrv := grpc.NewServer(dialOptions...)
	pb.RegisterFakeServerServer(grpcsrv, srv)
	reflection.Register(grpcsrv)

	return &GRPCServer{srv: grpcsrv, opts: grpcOptions}, nil
}

func (s *GRPCServer) RunOrDie() {
	lis, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		log.Fatalw("Failed to listen", "err", err)
	}

	log.Infow("Start to listening the incoming requests on grpc address", "addr", s.opts.Addr)
	if err := s.srv.Serve(lis); err != nil {
		log.Fatalw(err.Error())
	}
}

func (s *GRPCServer) GracefulStop() {
	log.Infof("Gracefully stop grpc server")
	s.srv.GracefulStop()
}
