// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/cacheserver/biz"
	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
)

// ProviderSet contains providers for creating instances of the biz struct.
var ProviderSet = wire.NewSet(NewCacheServerService, wire.Bind(new(v1.CacheServerServer), new(*CacheServerService)))

type CacheServerService struct {
	v1.UnimplementedCacheServerServer

	biz biz.IBiz
}

// Ensure that CacheServerService implements the v1.CacheServerServer interface.
var _ v1.CacheServerServer = (*CacheServerService)(nil)

func NewCacheServerService(biz biz.IBiz) *CacheServerService {
	return &CacheServerService{biz: biz}
}
