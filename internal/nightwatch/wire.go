// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//go:build wireinject
// +build wireinject

package nightwatch

//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	gwstore "github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/nightwatch/biz"
	"github.com/superproj/onex/internal/nightwatch/service/v1"
	nwstore "github.com/superproj/onex/internal/nightwatch/store"
	"github.com/superproj/onex/internal/nightwatch/validation"
	"github.com/superproj/onex/internal/pkg/client/store"
	ucstore "github.com/superproj/onex/internal/usercenter/store"
)

func wireAggregateStore(*gorm.DB) (store.Interface, error) {
	wire.Build(
		store.ProviderSet,
		gwstore.ProviderSet,
		ucstore.ProviderSet,
	)

	return nil, nil
}

func wireService(*gorm.DB) *v1.NightWatchService {
	wire.Build(
		validation.ProviderSet,
		biz.ProviderSet,
		nwstore.ProviderSet,
		v1.NewNightWatchService,
	)

	return nil
}

func wireStore(*gorm.DB) (nwstore.IStore, error) {
	wire.Build(nwstore.ProviderSet)

	return nil, nil
}
