// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// this file contains factories with no other dependencies

package util

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"

	clioptions "github.com/superproj/onex/internal/onexctl/util/options"
	"github.com/superproj/onex/internal/pkg/middleware/authn/jwt"
	kubeutil "github.com/superproj/onex/internal/pkg/util/kube"
	gatewayv1 "github.com/superproj/onex/pkg/api/gateway/v1"
	usercenterv1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

type factoryImpl struct {
	opts *clioptions.Options
}

var _ Factory = (*factoryImpl)(nil)

func NewFactory(opts *clioptions.Options) Factory {
	if opts == nil {
		klog.Fatal("attempt to instantiate client_access_factory with nil clientGetter")
	}

	return &factoryImpl{opts: opts}
}

func (f *factoryImpl) GetOptions() *clioptions.Options {
	return f.opts
}

func (f *factoryImpl) UserCenterClient() usercenterv1.UserCenterHTTPClient {
	conn := newConnect(f.opts.UserCenterOptions, jwt.WithToken(f.MustToken()))
	return usercenterv1.NewUserCenterHTTPClient(conn)
}

func (f *factoryImpl) GatewayClient() gatewayv1.GatewayHTTPClient {
	conn := newConnect(f.opts.GatewayOptions, jwt.WithToken(f.MustToken()))
	return gatewayv1.NewGatewayHTTPClient(conn)
}

func (f *factoryImpl) MustToken() string {
	opts := f.opts.UserOptions
	// Using BearerToken as the first choice.
	if opts.BearerToken != "" {
		return opts.BearerToken
	}

	// Using SecretID and SecretKey as the second choice.
	if opts.SecretID != "" && opts.SecretKey != "" {
		token, err := SignToken(opts.SecretID, opts.SecretKey)
		if err != nil {
			klog.Fatal(err.Error())
		}

		return token
	}

	// Using username and password as the third choice.
	if opts.Username != "" && opts.Password != "" {
		token, err := f.Login()
		if err != nil {
			klog.Fatal(err.Error())
		}

		return token
	}

	return ""
}

func (f *factoryImpl) WithToken(ctx context.Context) (context.Context, error) {
	md := metadata.Pairs("Authorization", "Bearer "+f.MustToken())
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func (f *factoryImpl) Login() (token string, err error) {
	client := usercenterv1.NewUserCenterHTTPClient(newConnect(f.opts.UserCenterOptions))
	rp, err := client.Login(context.Background(), &usercenterv1.LoginRequest{
		Username: f.opts.UserOptions.Username,
		Password: f.opts.UserOptions.Password,
	})
	if err != nil {
		return "", err
	}

	klog.V(4).Infof("Get login token: %s", rp.AccessToken)
	return rp.AccessToken, nil
}

func newConnect(opts *clioptions.ServerOptions, mws ...middleware.Middleware) *transhttp.Client {
	conn, err := transhttp.NewClient(
		context.Background(),
		transhttp.WithEndpoint(opts.Addr),
		transhttp.WithTimeout(opts.Timeout),
		transhttp.WithUserAgent(kubeutil.GetUserAgent("onexctl")),
		transhttp.WithMiddleware(mws...),
	)
	if err != nil {
		panic(err)
	}

	return conn
}
