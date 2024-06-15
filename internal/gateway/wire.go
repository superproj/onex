// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package gateway

//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"

	"github.com/superproj/onex/internal/gateway/biz"
	"github.com/superproj/onex/internal/gateway/server"
	"github.com/superproj/onex/internal/gateway/service"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/pkg/bootstrap"
	"github.com/superproj/onex/pkg/db"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// wireApp init kratos application.
func wireApp(
	<-chan struct{},
	bootstrap.AppInfo,
	*server.Config,
	clientset.Interface,
	*db.MySQLOptions,
	*db.RedisOptions,
	*genericoptions.RedisOptions,
	*genericoptions.EtcdOptions,
) (*kratos.App, func(), error) {
	wire.Build(
		bootstrap.ProviderSet,
		bootstrap.NewEtcdRegistrar,
		server.ProviderSet,
		store.ProviderSet,
		db.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		createInformers,
	)

	return nil, nil, nil
}
