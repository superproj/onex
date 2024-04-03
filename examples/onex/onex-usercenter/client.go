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

	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

func main() {
	callHTTP()
	callGRPC()
}

func callHTTP() {
	conn, err := transhttp.NewClient(
		context.Background(),
		transhttp.WithMiddleware(
			recovery.Recovery(),
		),
		transhttp.WithEndpoint("onex.usercenter.superproj.com:18080"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewUserCenterHTTPClient(conn)
	reply, err := client.Login(context.Background(), &v1.LoginRequest{Username: "colin", Password: "onex(#)666"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[http] Login %s\n", reply.RefreshToken)

	// returns error
	_, err = client.Login(context.Background(), &v1.LoginRequest{Username: "colin", Password: "badpassword"})
	if err != nil {
		log.Printf("[http] Login error: %v\n", err)
	}
	if errors.IsBadRequest(err) {
		log.Printf("[http] Login error is invalid argument: %v\n", err)
	}
}

func callGRPC() {
	conn, err := transgrpc.DialInsecure(
		context.Background(),
		transgrpc.WithEndpoint("onex.usercenter.superproj.com:18080"),
		transgrpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewUserCenterClient(conn)
	reply, err := client.Login(context.Background(), &v1.LoginRequest{Username: "colin", Password: "onex(#)666"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[grpc] Login %+v\n", reply.RefreshToken)

	// returns error
	_, err = client.Login(context.Background(), &v1.LoginRequest{Username: "colin", Password: "badpassword"})
	if err != nil {
		log.Printf("[grpc] Login error: %v\n", err)
	}
	if errors.IsBadRequest(err) {
		log.Printf("[grpc] Login error is invalid argument: %v\n", err)
	}
}
