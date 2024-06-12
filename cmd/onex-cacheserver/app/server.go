// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/superproj/onex/cmd/onex-cacheserver/app/options"
	"github.com/superproj/onex/internal/cacheserver"
	"github.com/superproj/onex/pkg/app"
)

const commandDesc = `onex-cacheserver is an example cache server, demonstrating 
how to develop a caching service.

Find more onex-cacheserver information at:
    https://github.com/superproj/onex/blob/master/docs/guide/en-US/cmd/onex-cacheserver.md`

// NewApp creates an App object with default parameters.
func NewApp() *app.App {
	opts := options.NewOptions()
	application := app.NewApp("onex-cacheserver", "Launch a onex cache server",
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
func Run(c *cacheserver.Config, stopCh <-chan struct{}) error {
	server, err := c.Complete().New(stopCh)
	if err != nil {
		return err
	}

	return server.Run(stopCh)
}
