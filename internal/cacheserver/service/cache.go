// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
	"github.com/superproj/onex/pkg/log"
)

func (s *CacheServerService) Set(ctx context.Context, rq *v1.SetRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("Set function called")
	return &emptypb.Empty{}, s.biz.Namespace(rq.Namespace).Set(ctx, rq.Key, rq.Value, rq.Expire)
}

func (s *CacheServerService) Get(ctx context.Context, rq *v1.GetRequest) (*v1.GetResponse, error) {
	log.C(ctx).Infow("Get function called")
	return s.biz.Namespace(rq.Namespace).Get(ctx, rq.Key)
}

func (s *CacheServerService) Del(ctx context.Context, rq *v1.DelRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("Del function called")
	return &emptypb.Empty{}, s.biz.Namespace(rq.Namespace).Del(ctx, rq.Key)
}

func (s *CacheServerService) SetSecret(ctx context.Context, rq *v1.SetSecretRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("SetSecret function called")
	return &emptypb.Empty{}, s.biz.Secrets().Set(ctx, rq)
}

func (s *CacheServerService) GetSecret(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	log.C(ctx).Infow("Get secret function called")
	return s.biz.Secrets().Get(ctx, rq)
}

func (s *CacheServerService) DelSecret(ctx context.Context, rq *v1.DelSecretRequest) (*emptypb.Empty, error) {
	log.C(ctx).Infow("Del function called")
	return &emptypb.Empty{}, s.biz.Secrets().Del(ctx, rq)
}
