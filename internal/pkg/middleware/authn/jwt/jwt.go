// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package jwt

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/pkg/authn"
	"github.com/superproj/onex/pkg/log"
)

const (
	// reason holds the error reason.
	reason string = "UNAUTHORIZED"

	// bearerWord the bearer key word for authorization.
	bearerWord string = "Bearer"

	// bearerFormat authorization token format.
	bearerFormat string = "Bearer %s"

	// authorizationKey holds the key used to store the JWT Token in the rquest tokenHeader.
	authorizationKey string = "Authorization"
)

var (
	ErrMissingJwtToken = errors.Unauthorized(reason, "JWT token is missing")
	ErrWrongContext    = errors.Unauthorized(reason, "Wrong context for middleware")
)

// Server is a server auth middleware. Check the token and extract the info from token.
func Server(a authn.Authenticator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (any, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				auths := strings.SplitN(tr.RequestHeader().Get(authorizationKey), " ", 2)
				if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
					return nil, ErrMissingJwtToken
				}

				accessToken := auths[1]
				claims, err := a.ParseClaims(ctx, accessToken)
				if err != nil {
					return nil, err
				}

				ctx = onexx.NewContext(ctx, claims)
				ctx = onexx.NewUserID(ctx, claims.Subject)
				ctx = onexx.NewAccessToken(ctx, accessToken)
				ctx = log.WithContext(ctx, "user.id", claims.Subject)
				return handler(ctx, rq)
			}
			return nil, ErrWrongContext
		}
	}
}

// WithSignToken is a client jwt middleware used to sign a jwt token.
func WithSignToken(a authn.Authenticator, userID string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (any, error) {
			accessToken, err := a.Sign(ctx, userID)
			if err != nil {
				return nil, err
			}

			if clientContext, ok := transport.FromClientContext(ctx); ok {
				clientContext.RequestHeader().Set(authorizationKey, fmt.Sprintf(bearerFormat, accessToken))
				return handler(ctx, rq)
			}
			return nil, ErrWrongContext
		}
	}
}

// WithToken is a client jwt middleware.
func WithToken(token string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (any, error) {
			if clientContext, ok := transport.FromClientContext(ctx); ok {
				clientContext.RequestHeader().Set(authorizationKey, fmt.Sprintf(bearerFormat, token))
				return handler(ctx, rq)
			}
			return nil, ErrWrongContext
		}
	}
}
