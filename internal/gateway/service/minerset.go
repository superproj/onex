// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/superproj/onex/internal/pkg/onexx"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

func (s *GatewayService) CreateMinerSet(ctx context.Context, ms *v1beta1.MinerSet) (*emptypb.Empty, error) {
	if err := s.biz.MinerSets().Create(ctx, onexx.FromUserID(ctx), ms); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *GatewayService) ListMinerSet(ctx context.Context, rq *v1.ListMinerSetRequest) (*v1.ListMinerSetResponse, error) {
	mss, err := s.biz.MinerSets().List(ctx, onexx.FromUserID(ctx), rq)
	if err != nil {
		return &v1.ListMinerSetResponse{}, err
	}

	return mss, nil
}

func (s *GatewayService) GetMinerSet(ctx context.Context, rq *v1.GetMinerSetRequest) (*v1beta1.MinerSet, error) {
	ms, err := s.biz.MinerSets().Get(ctx, onexx.FromUserID(ctx), rq.Name)
	if err != nil {
		return &v1beta1.MinerSet{}, err
	}

	return ms, nil
}

func (s *GatewayService) UpdateMinerSet(ctx context.Context, ms *v1beta1.MinerSet) (*emptypb.Empty, error) {
	if err := s.biz.MinerSets().Update(ctx, onexx.FromUserID(ctx), ms); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *GatewayService) DeleteMinerSet(ctx context.Context, rq *v1.DeleteMinerSetRequest) (*emptypb.Empty, error) {
	if err := s.biz.MinerSets().Delete(ctx, onexx.FromUserID(ctx), rq.Name); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *GatewayService) ScaleMinerSet(ctx context.Context, rq *v1.ScaleMinerSetRequest) (*emptypb.Empty, error) {
	if err := s.biz.MinerSets().Scale(ctx, onexx.FromUserID(ctx), rq.Name, rq.Replicas); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
