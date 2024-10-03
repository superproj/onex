// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	jwtauthn "github.com/superproj/onex/pkg/authn/jwt"
)

func main() {
	headers := make(map[string]any)
	headers["kid"] = "8b5228a5-b3d2-4165-aaac-58a052629846"

	opts := []jwtauthn.Option{
		jwtauthn.WithSigningMethod(jwt.GetSigningMethod("HS256")),
		jwtauthn.WithIssuer("examples/onex/onex-usercenter/token"),
		jwtauthn.WithTokenHeader(headers),
		jwtauthn.WithExpired(2 * time.Hour),
		jwtauthn.WithSigningKey([]byte("98506f66-6247-49c1-88c8-3c7a4c8489b9")),
	}
	j, err := jwtauthn.New(nil, opts...).Sign(context.Background(), "")
	if err != nil {
		panic(err)
	}

	fmt.Println(j.GetToken())
}
