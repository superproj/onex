// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/spf13/pflag"

	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

var token = pflag.StringP("token", "t", "", "Access token used to access onex-gateway.")

func withToken(token string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (any, error) {
			if clientContext, ok := transport.FromClientContext(ctx); ok {
				clientContext.RequestHeader().Set("Authorization", fmt.Sprintf("Bearer %s", token))
				return handler(ctx, rq)
			}
			return nil, fmt.Errorf("Wrong context for middleware")
		}
	}
}

func main() {
	pflag.Parse()

	conn, err := transhttp.NewClient(
		context.Background(),
		transhttp.WithEndpoint("onex.usercenter.superproj.com:51843"),
		transhttp.WithTimeout(30*time.Second),
		transhttp.WithUserAgent("examples/onex/onex-usercenter/client"),
		transhttp.WithMiddleware(withToken(*token)),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewGatewayHTTPClient(conn)
	mss, err := client.ListMinerSet(context.Background(), &v1.ListMinerSetRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, ms := range mss.MinerSets {
		fmt.Println(ms.Name)
	}
}
