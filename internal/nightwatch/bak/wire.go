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

	gwstore "github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/pkg/client/store"
	ucstore "github.com/superproj/onex/internal/usercenter/store"
	"github.com/superproj/onex/pkg/db"
)

func wireStoreClient(*db.MySQLOptions) (store.Interface, error) {
	wire.Build(
		db.ProviderSet,
		store.ProviderSet,
		gwstore.ProviderSet,
		ucstore.ProviderSet,
	)

	return nil, nil
}
