// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/superproj/onex/cmd/onex-gateway/app/options"
	"github.com/superproj/onex/internal/gateway"
	"github.com/superproj/onex/pkg/app"
)

const commandDesc = `The gateway server is the back-end portal server of onex. All 
requests from the front-end will arrive at the gateway, requests will be uniformly processed 
and distributed by the gateway.`

// NewApp creates an App object with default parameters.
func NewApp() *app.App {
	opts := options.NewOptions()
	application := app.NewApp(gateway.Name, "Launch a onex gateway server",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

func run(opts *options.Options) app.RunFunc {
	return func() error {
		cfg, err := opts.Config()
		if err != nil {
			return err
		}

		return Run(cfg, genericapiserver.SetupSignalHandler())
	}
}

// Run runs the specified APIServer. This should never exit.
func Run(c *gateway.Config, stopCh <-chan struct{}) error {
	server, err := c.Complete().New(stopCh)
	if err != nil {
		return err
	}

	return server.Run(stopCh)
}
