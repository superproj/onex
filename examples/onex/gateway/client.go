// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"log"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/superproj/onex/internal/pkg/middleware/authn/jwt"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

func main() {
	callHTTP()
	// callGRPC()
}

func callHTTP() {
	conn, err := transhttp.NewClient(
		context.Background(),
		transhttp.WithEndpoint("onex.gateway.superproj.com:18080"),
		transhttp.WithMiddleware(
			jwt.WithToken("eyJhbGciOiJIUzUxMiIsImtpZCI6ImU2NTExMDc2LTlkZWUtNGIwNS04ODk5LTA4MDA0NDdkYjI4MSIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ6ZXJvLXVzZXJjZW50ZXIiLCJzdWIiOiJ1c2VyLWFkbWluIiwiZXhwIjoxNjk3MDM5ODYwLCJuYmYiOjE2OTcwMzI2NjAsImlhdCI6MTY5NzAzMjY2MH0.cYku0agwihDsrBpUbJ66n6mGu7_vREsBYLICY-bUislXDz7ydeuWqctKIfDkWaihk0jWnD_t54p37OYwLYtAmw"),
		),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewGatewayHTTPClient(conn)
	reply, err := client.GetMinerSet(context.Background(), &v1.GetMinerSetRequest{Name: "test"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[http] GetMinerSet %s\n", reply.Spec.DisplayName)
	if errors.IsBadRequest(err) {
		log.Printf("[http] Login error is invalid argument: %v\n", err)
	}
}

func callGRPC() {
	conn, err := transgrpc.DialInsecure(
		context.Background(),
		transgrpc.WithEndpoint("onex.gateway.superproj.com:39090"),
		transgrpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewGatewayClient(conn)
	reply, err := client.GetMinerSet(context.Background(), &v1.GetMinerSetRequest{Name: "test"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[grpc] GetMinerSet %+v\n", reply.Spec.DisplayName)
	if errors.IsBadRequest(err) {
		log.Printf("[grpc] Login error is invalid argument: %v\n", err)
	}
}
