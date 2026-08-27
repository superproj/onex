// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package gateway

//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/db"
	"github.com/onexstack/onexstack/pkg/server"
	genericvalidation "github.com/onexstack/onexstack/pkg/validation"

	"github.com/onexstack/onex/internal/gateway/biz"
	"github.com/onexstack/onex/internal/gateway/handler"
	"github.com/onexstack/onex/internal/gateway/pkg/validation"
	"github.com/onexstack/onex/internal/gateway/store"
	"github.com/onexstack/onex/internal/pkg/client/usercenter"
	"github.com/onexstack/onex/internal/pkg/idempotent"
	"github.com/onexstack/onex/internal/pkg/middleware/validate"
	clientset "github.com/onexstack/onex/pkg/generated/clientset/versioned"
)

func InitializeWebServer(
	<-chan struct{},
	*Config,
	clientset.Interface,
	*db.MySQLOptions,
	*db.RedisOptions,
) (server.Server, error) {
	wire.Build(
		NewWebServer,
		NewMiddlewares,
		ProvideKratosAppConfig,
		wire.NewSet(server.NewEtcdRegistrar, wire.FieldsOf(new(*Config), "EtcdOptions")),
		ProvideKratosLogger,
		handler.ProviderSet,
		store.ProviderSet,
		biz.ProviderSet,
		wire.NewSet(usercenter.ProviderSet, wire.FieldsOf(new(*Config), "UserCenterOptions")),
		db.ProviderSet,
		idempotent.ProviderSet,
		wire.NewSet(
			validation.ProviderSet,
			genericvalidation.NewValidator,
			wire.Bind(new(validate.RequestValidator), new(*genericvalidation.Validator)),
		),
		createInformers,
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
	)

	return nil, nil
}
