// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/superproj/onex/cmd/onex-nightwatch/app/options"
	"github.com/superproj/onex/internal/nightwatch"
	"github.com/superproj/onex/pkg/app"
)

const commandDesc = `The nightwatch server is responsible for executing some async tasks 
like linux cronjob. You can add Cron(github.com/robfig/cron) jobs on the given schedule
use the Cron spec format.`

// NewApp creates an App object with default parameters.
func NewApp() *app.App {
	opts := options.NewOptions()
	application := app.NewApp("onex-nightwatch", "Launch a onex asynchronous task processing server",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
		app.WithHealthCheckFunc(func() error {
			go opts.HealthOptions.ServeHealthCheck()
			return nil
		}),
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
func Run(c *nightwatch.Config, stopCh <-chan struct{}) error {
	nw, err := c.Complete().New()
	if err != nil {
		return err
	}

	nw.Run(stopCh)

	return nil
}
