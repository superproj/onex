// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/superproj/onex/cmd/onex-nightwatch/app/options"
	"github.com/superproj/onex/internal/nightwatch"
	"github.com/superproj/onex/pkg/app"
	genericoptions "github.com/superproj/onex/pkg/options"
)

const commandDesc = `The nightwatch server is responsible for executing some async tasks 
like linux cronjob. You can add Cron(github.com/robfig/cron) jobs on the given schedule
use the Cron spec format.`

// jobServer represents the HTTP server with optional TLS and graceful shutdown capabilities.
type jobServer struct {
	stopCh     <-chan struct{}
	tlsOptions *genericoptions.TLSOptions
}

// Option is a function that configures the jobServer.
type Option func(jrs *jobServer)

// WithTLSOptions sets the TLS options for the job REST server.
func WithTLSOptions(tlsOptions *genericoptions.TLSOptions) Option {
	return func(jrs *jobServer) {
		jrs.tlsOptions = tlsOptions
	}
}

// WithStopChannel sets the stop channel for graceful shutdown.
func WithStopChannel(stopCh <-chan struct{}) Option {
	return func(jrs *jobServer) {
		jrs.stopCh = stopCh
	}
}

// NewApp creates an App object with default parameters and configurations.
func NewApp(appName string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp(appName, "Launch an asynchronous task processing server",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

// run executes the application logic using the provided options.
func run(opts *options.Options) app.RunFunc {
	return func() error {
		cfg, err := opts.Config()
		if err != nil {
			return err
		}

		return Run(cfg, genericapiserver.SetupSignalHandler())
	}
}

// Run starts the specified APIServer. This function should never exit.
func Run(cfg *nightwatch.Config, stopCh <-chan struct{}) error {
	nw, err := cfg.New(stopCh)
	if err != nil {
		return err
	}

	nw.Run(stopCh)

	return nil
}

// NewJobServer creates a new instance of the job server with the specified options.
func NewJobServer(addr string, db *gorm.DB, opts ...Option) *nightwatch.RESTServer {
	jrs := jobServer{}
	for _, opt := range opts {
		opt(&jrs)
	}

	return nightwatch.NewRESTServer(jrs.stopCh, addr, jrs.tlsOptions, db)
}

// InstallJobAPI sets up the job-related routes in the provided router.
func InstallJobAPI(router *gin.Engine, db *gorm.DB) {
	nightwatch.InstallJobAPI(router, db)
}
