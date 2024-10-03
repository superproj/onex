// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	durationpb "google.golang.org/protobuf/types/known/durationpb"

	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
)

const (
	// gRPC 服务地址
	address = "127.0.0.1:57090"
)

func main() {
	// 建立到 gRPC 服务的连接
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到 gRPC 服务: %v", err)
	}
	defer conn.Close()

	key := "exampletest"

	// 创建 gRPC 客户端
	client := v1.NewCacheServerClient(conn)

	// 调用 gRPC 服务的方法
	setrq := &v1.SetSecretRequest{Key: key, Name: key, Expire: durationpb.New(30 * time.Second)}
	_, err = client.SetSecret(context.Background(), setrq)
	if err != nil {
		log.Fatalf("调用 gRPC 服务失败: %v", err)
	}

	rq := &v1.GetSecretRequest{Key: key}
	rp, err := client.GetSecret(context.Background(), rq)
	if err != nil {
		log.Fatalf("调用 gRPC 服务失败: %v", err)
	}

	// 处理 gRPC 服务的响应
	log.Printf("收到 gRPC 服务的响应: `%s`", rp.String())
}
