// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cacheserver

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/superproj/onex/pkg/api/cacheserver/v1"
	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

type Server interface {
	RunOrDie()
	GracefulStop()
}

type GRPCServer struct {
	srv  *grpc.Server
	opts *genericoptions.GRPCOptions
}

func NewGRPCServer(
	grpcOptions *genericoptions.GRPCOptions,
	tlsOptions *genericoptions.TLSOptions,
	srv pb.CacheServerServer,
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
	pb.RegisterCacheServerServer(grpcsrv, srv)

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
