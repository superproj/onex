// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package toyblc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"

	"github.com/superproj/onex/internal/toyblc/blc"
	wscontroller "github.com/superproj/onex/internal/toyblc/controller/v1/ws"
	mw "github.com/superproj/onex/internal/toyblc/middleware"
	"github.com/superproj/onex/internal/toyblc/miner"
	"github.com/superproj/onex/internal/toyblc/ws"
	"github.com/superproj/onex/pkg/log"
	genericmw "github.com/superproj/onex/pkg/middleware/gin"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// Config represents the configuration of the service.
type Config struct {
	Miner           bool
	MinMineInterval time.Duration
	Address         string
	Accounts        map[string]string
	HTTPOptions     *genericoptions.HTTPOptions
	TLSOptions      *genericoptions.TLSOptions
	P2PAddr         string
	Peers           []string
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() completedConfig {
	return completedConfig{cfg}
}

type completedConfig struct {
	*Config
}

// New returns a new instance of ToyBLC from the given config.
func (c completedConfig) New() (*ToyBLC, error) {
	bs, ss := blc.NewBlockSet(c.Address), ws.NewSockets()

	// gin.Recovery() 中间件，用来捕获任何 panic，并恢复
	mws := []gin.HandlerFunc{gin.Recovery(), genericmw.NoCache, genericmw.Cors, genericmw.Secure, mw.TraceID()}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	g := gin.New()

	// 添加中间件
	g.Use(mws...)

	// 并初始化路由
	installRouters(g, bs, ss, c.Accounts)

	// 创建 HTTP Server 实例
	httpsrv := &http.Server{Addr: c.HTTPOptions.Addr, Handler: g}
	if c.TLSOptions != nil && c.TLSOptions.UseTLS {
		tlsConfig, err := c.TLSOptions.TLSConfig()
		if err != nil {
			return nil, err
		}

		httpsrv.TLSConfig = tlsConfig
	}

	p2p := gin.New()
	wsc := wscontroller.New(bs, ss)
	p2p.Use(gin.WrapH(websocket.Handler(wsc.WSHandler)))

	p2psrv := &http.Server{Addr: c.P2PAddr, Handler: p2p}
	return &ToyBLC{
		config:          c,
		srv:             httpsrv,
		p2psrv:          p2psrv,
		bs:              bs,
		ss:              ss,
		miner:           c.Miner,
		minMineInterval: c.MinMineInterval,
		peers:           c.Peers,
	}, nil
}

// ToyBLC represents the toyblc application.
type ToyBLC struct {
	config          completedConfig
	srv             *http.Server
	p2psrv          *http.Server
	bs              *blc.BlockSet
	ss              *ws.Sockets
	miner           bool
	minMineInterval time.Duration
	peers           []string
}

func (t *ToyBLC) Run(stopCh <-chan struct{}) error {
	if t.miner {
		miner.NewMiner(t.bs, t.ss, t.minMineInterval).Start()
	}

	// 运行 HTTP 服务器。在 goroutine 中启动服务器，它不会阻止下面的正常关闭处理流程
	// 打印一条日志，用来提示 HTTP 服务已经起来，方便排障
	log.Infof("Start to listening incoming %s requests on %s",
		scheme(t.config.TLSOptions),
		t.config.HTTPOptions.Addr,
	)

	go func() {
		if err := t.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw(err.Error())
		}
	}()

	log.Infof("Start listening for incoming http requests on the P2P address %s", t.config.P2PAddr)
	go func() {
		if err := t.p2psrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw(err.Error())
		}
	}()

	ws.ConnectToPeers(context.Background(), t.bs, t.ss, t.peers)

	<-stopCh
	log.Infow("Shutting down server ...")

	// 创建 ctx 用于通知服务器 goroutine, 它有 10 秒时间完成当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 10 秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过 10 秒就超时退出
	if err := t.srv.Shutdown(ctx); err != nil {
		log.Errorw(err, "HTTP(s) server forced to shutdown")
		return err
	}

	if err := t.p2psrv.Shutdown(ctx); err != nil {
		log.Errorw(err, "P2P server forced to shutdown")
		return err
	}

	log.Infow("Server exiting")
	return nil
}
