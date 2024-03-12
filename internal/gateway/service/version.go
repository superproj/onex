// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/version"
)

func (s *GatewayService) GetVersion(ctx context.Context, rq *emptypb.Empty) (*v1.GetVersionResponse, error) {
	vinfo := version.Get()
	return &v1.GetVersionResponse{
		GitVersion:   vinfo.GitVersion,
		GitCommit:    vinfo.GitCommit,
		GitTreeState: vinfo.GitTreeState,
		BuildDate:    vinfo.BuildDate,
		GoVersion:    vinfo.GoVersion,
		Compiler:     vinfo.Compiler,
		Platform:     vinfo.Platform,
	}, nil
}
