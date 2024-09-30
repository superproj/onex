// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/superproj/onex/internal/nightwatch/biz"
	"github.com/superproj/onex/internal/nightwatch/controller/v1/cronjob"
	"github.com/superproj/onex/internal/nightwatch/controller/v1/job"
	"github.com/superproj/onex/internal/nightwatch/watcher"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/all"
	"github.com/superproj/onex/internal/pkg/core"
	"github.com/superproj/onex/pkg/api/zerrors"
	"github.com/superproj/onex/pkg/db"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/log"
	genericmw "github.com/superproj/onex/pkg/middleware/gin"
	genericoptions "github.com/superproj/onex/pkg/options"
	"github.com/superproj/onex/pkg/watch"
	"github.com/superproj/onex/pkg/watch/logger/onex"
)

func (c *CompletedConfig) StartRESTServer(stopCh <-chan struct{}) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), genericmw.NoCache, genericmw.Cors, genericmw.Secure, mw.TraceID())

	InstallJobAPI(router, c.DB)

	// Create HTTP Server instance.
	srv := &http.Server{Addr: c.HTTPOptions.Addr, Handler: g}
	if c.TLSOptions != nil && c.TLSOptions.UseTLS {
		tlsConfig, err := c.TLSOptions.TLSConfig()
		if err != nil {
			return err
		}
		srv.TLSConfig = tlsConfig
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw("Server error", "err", err)
		}
	}()

	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Infof("HTTP server forced to shutdown: %v", err)
		return err
	}

	log.Infof("HTTP server exited gracefully")
	return nil
}

func InstallJobAPI(router *gin.Engine, db *gorm.DB) {
	router.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, zerror.ErrorPageNotFound("route not found"), nil)
	})

	biz := wireBiz(db)

	cronJobController := cronjob.New(biz)
	jobController := job.New(biz)

	v1 := g.Group("/v1")
	{
		cronjobv1 := v1.Group("/cronjobs")
		{
			cronjobv1.POST("", cronJobController.Create)
			cronjobv1.PUT(":cronJobID", cronJobController.Update)
			cronjobv1.DELETE(":cronJobID", cronJobController.Delete)
			cronjobv1.GET(":cronJobID", cronJobController.Get)
			cronjobv1.GET("", cronJobController.List)
		}

		jobv1 := v1.Group("/jobs")
		{
			jobv1.POST("", jobController.Create)
			jobv1.PUT(":jobID", jobController.Update)
			jobv1.DELETE(":jobID", jobController.Delete)
			jobv1.GET(":jobID", jobController.Get)
			jobv1.GET("", jobController.List)
		}
	}
}
