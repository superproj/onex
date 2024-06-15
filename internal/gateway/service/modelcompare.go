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

func (s *GatewayService) CreateModelCompare(ctx context.Context, ms *v1beta1.ModelCompare) (*emptypb.Empty, error) {
	if err := s.biz.ModelCompares().Create(ctx, onexx.FromUserID(ctx), ms); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *GatewayService) ListModelCompare(ctx context.Context, rq *v1.ListModelCompareRequest) (*v1.ListModelCompareResponse, error) {
	mss, err := s.biz.ModelCompares().List(ctx, onexx.FromUserID(ctx), rq)
	if err != nil {
		return &v1.ListModelCompareResponse{}, err
	}

	return mss, nil
}

func (s *GatewayService) GetModelCompare(ctx context.Context, rq *v1.GetModelCompareRequest) (*v1beta1.ModelCompare, error) {
	ms, err := s.biz.ModelCompares().Get(ctx, onexx.FromUserID(ctx), rq.Name)
	if err != nil {
		return &v1beta1.ModelCompare{}, err
	}

	return ms, nil
}

func (s *GatewayService) UpdateModelCompare(ctx context.Context, ms *v1beta1.ModelCompare) (*emptypb.Empty, error) {
	if err := s.biz.ModelCompares().Update(ctx, onexx.FromUserID(ctx), ms); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *GatewayService) DeleteModelCompare(ctx context.Context, rq *v1.DeleteModelCompareRequest) (*emptypb.Empty, error) {
	if err := s.biz.ModelCompares().Delete(ctx, onexx.FromUserID(ctx), rq.Name); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
