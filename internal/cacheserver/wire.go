// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//go:build wireinject
// +build wireinject

package cacheserver

//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/wire"

	"github.com/superproj/onex/internal/cacheserver/biz"
	"github.com/superproj/onex/internal/cacheserver/service"
	"github.com/superproj/onex/internal/cacheserver/store"
	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
	"github.com/superproj/onex/pkg/cache"
	"github.com/superproj/onex/pkg/db"
	// genericoptions "github.com/superproj/onex/pkg/options"
)

func wireServer(
	*db.MySQLOptions,
	cache.Cache[*any.Any],
	bool,
) (v1.CacheServerServer, error) {
	wire.Build(
		db.ProviderSet,
		store.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
	)

	return nil, nil
}
