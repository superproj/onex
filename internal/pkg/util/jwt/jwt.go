// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package jwt

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	// bearerWord the bearer key word for authorization.
	bearerWord string = "Bearer"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	authorizationKey string = "Authorization"
)

func TokenFromServerContext(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		auths := strings.SplitN(tr.RequestHeader().Get(authorizationKey), " ", 2)
		if len(auths) == 2 && strings.EqualFold(auths[0], bearerWord) {
			return auths[1]
		}
	}

	if md, ok := metadata.FromServerContext(ctx); ok {
		return md.Get("x-md-global-jwt")
	}

	return ""
}
