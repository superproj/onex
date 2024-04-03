// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package bootstrap

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(wire.Struct(new(AppConfig), "*"), NewLogger, NewApp)

type AppInfo struct {
	ID       string
	Name     string
	Version  string
	Metadata map[string]string
}

func NewAppInfo(id, name, version string) AppInfo {
	if id == "" {
		id, _ = os.Hostname()
	}
	return AppInfo{
		Name:     name,
		Version:  version,
		ID:       id,
		Metadata: map[string]string{},
	}
}

// The purpose of defining the AppConfig is to demonstrate the usage of wire.Struct.
type AppConfig struct {
	Info      AppInfo
	Logger    krtlog.Logger
	Registrar registry.Registrar
}

// NewApp creates a new kratos app.
func NewApp(c AppConfig, servers ...transport.Server) *kratos.App {
	return kratos.New(
		kratos.ID(c.Info.ID+"."+c.Info.Name),
		kratos.Name(c.Info.Name),
		kratos.Version(c.Info.Version),
		kratos.Metadata(c.Info.Metadata),
		kratos.Logger(c.Logger),
		kratos.Registrar(c.Registrar),
		kratos.Server(servers...),
	)
}
