// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package idempotent

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/superproj/onex/internal/pkg/idempotent"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/api/zerrors"
)

func idempotentBlacklist() selector.MatchFunc {
	blacklist := make(map[string]struct{})
	blacklist[v1.OperationGatewayCreateMiner] = struct{}{}
	blacklist[v1.OperationGatewayCreateMinerSet] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := blacklist[operation]; ok {
			return true
		}
		return false
	}
}

func Idempotent(idt *idempotent.Idempotent) middleware.Middleware {
	return selector.Server(
		func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, rq any) (rp any, err error) {
				if tr, ok := transport.FromServerContext(ctx); ok {
					token := tr.RequestHeader().Get("X-Idempotent-ID")
					if token != "" {
						if idt.Check(ctx, token) {
							return handler(ctx, rq)
						}
						return nil, zerrors.ErrorIdempotentTokenExpired("idempotent token is invalid")
					}
				}

				return nil, zerrors.ErrorIdempotentMissingToken("idempotent token is missing")
			}
		},
	).Match(idempotentBlacklist()).Build()
}
