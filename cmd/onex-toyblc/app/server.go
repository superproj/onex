// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/superproj/onex/cmd/onex-toyblc/app/options"
	"github.com/superproj/onex/internal/toyblc"
	"github.com/superproj/onex/pkg/app"
)

// Define the description of the command.
const commandDesc = `The toy blc is used to start a naive and simple blockchain node.`

// NewApp creates and returns a new App object with default parameters.
func NewApp() *app.App {
	opts := options.NewOptions()
	application := app.NewApp("onex-toyblc", "Launch a onex toy blockchain node",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

// Returns the function to run the application.
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
func Run(c *toyblc.Config, stopCh <-chan struct{}) error {
	server, err := c.Complete().New()
	if err != nil {
		return err
	}

	return server.Run(stopCh)
}
