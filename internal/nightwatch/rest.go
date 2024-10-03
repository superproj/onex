// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/nightwatch/middleware"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/all"
	"github.com/superproj/onex/internal/pkg/core"
	"github.com/superproj/onex/pkg/api/zerrors"
	"github.com/superproj/onex/pkg/log"
	genericmw "github.com/superproj/onex/pkg/middleware/gin"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// RESTServer represents the HTTP server with optional TLS and graceful shutdown capabilities.
type RESTServer struct {
	stopCh <-chan struct{}
	addr   string
	srv    *http.Server
}

// NewRESTServer creates a new instance of RESTServer with the specified options.
func NewRESTServer(stopCh <-chan struct{}, addr string, tlsOptions *genericoptions.TLSOptions, db *gorm.DB) *RESTServer {
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Recovery(), genericmw.NoCache, genericmw.Cors, genericmw.Secure, middleware.Context())

	InstallJobAPI(router, db)

	// Create HTTP Server instance.
	srv := &http.Server{Addr: addr, Handler: router}

	if tlsOptions != nil && tlsOptions.UseTLS {
		tlsConfig, err := tlsOptions.TLSConfig()
		if err != nil {
			log.Fatalw("Failed to create TLS config", "err", err)
		}
		srv.TLSConfig = tlsConfig
	}

	return &RESTServer{stopCh: stopCh, addr: addr, srv: srv}
}

// Start begins listening for incoming requests and handles graceful shutdown if stopCh is provided.
func (rs *RESTServer) Start() {
	listenAndServe := func() {
		log.Infow("Starting job server", "addr", rs.addr)
		if err := rs.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw("Server error", "err", err)
		}
	}

	if rs.stopCh == nil {
		listenAndServe()
		return
	}

	go listenAndServe()
	<-rs.stopCh
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rs.srv.Shutdown(ctx); err != nil {
		log.Infof("HTTP server forced to shutdown: %v", err)
	}

	log.Infof("HTTP server exited gracefully")
}

func InstallJobAPI(router *gin.Engine, db *gorm.DB) {
	router.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, zerrors.ErrorPageNotFound("route not found"), nil)
	})

	svc := wireService(db)

	v1 := router.Group("/v1")
	{
		cronjobv1 := v1.Group("/cronjobs")
		{
			cronjobv1.POST("", svc.CreateCronJob)
			cronjobv1.PUT(":cronJobID", svc.UpdateCronJob)
			cronjobv1.DELETE(":cronJobID", svc.DeleteCronJob)
			cronjobv1.GET(":cronJobID", svc.GetCronJob)
			cronjobv1.GET("", svc.ListCronJob)
		}

		jobv1 := v1.Group("/jobs")
		{
			jobv1.POST("", svc.CreateJob)
			jobv1.PUT(":jobID", svc.UpdateJob)
			jobv1.DELETE(":jobID", svc.DeleteJob)
			jobv1.GET(":jobID", svc.GetJob)
			jobv1.GET("", svc.ListJob)
		}
	}
}
