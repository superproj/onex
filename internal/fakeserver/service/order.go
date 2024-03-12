// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/superproj/onex/pkg/api/fakeserver/v1"
	"github.com/superproj/onex/pkg/log"
)

func (s *FakeServerService) CreateOrder(ctx context.Context, rq *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	log.C(ctx).Infow("CreateOrder function called")
	return s.biz.Orders().Create(ctx, rq)
}

func (s *FakeServerService) ListOrder(ctx context.Context, rq *v1.ListOrderRequest) (*v1.ListOrderResponse, error) {
	log.C(ctx).Infow("ListOrder function called")
	return s.biz.Orders().List(ctx, rq)
}

func (s *FakeServerService) GetOrder(ctx context.Context, rq *v1.GetOrderRequest) (*v1.OrderReply, error) {
	log.C(ctx).Infow("GetOrder function called")
	return s.biz.Orders().Get(ctx, rq)
}

func (s *FakeServerService) UpdateOrder(ctx context.Context, rq *v1.UpdateOrderRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("UpdateOrder function called")
	return &emptypb.Empty{}, s.biz.Orders().Update(ctx, rq)
}

func (s *FakeServerService) DeleteOrder(ctx context.Context, rq *v1.DeleteOrderRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("DeleteOrder function called")
	return &emptypb.Empty{}, s.biz.Orders().Delete(ctx, rq)
}
