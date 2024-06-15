// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/gateway/biz"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGatewayService)

type GatewayService struct {
	v1.UnimplementedGatewayServer

	biz biz.IBiz
}

func NewGatewayService(biz biz.IBiz) *GatewayService {
	return &GatewayService{biz: biz}
}
