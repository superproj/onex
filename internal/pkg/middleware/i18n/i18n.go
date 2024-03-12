// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package i18n

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"

	"github.com/superproj/onex/pkg/i18n"
)

func Translator(options ...func(*i18n.Options)) middleware.Middleware {
	i := i18n.New(options...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (rp any, err error) {
			var lang language.Tag
			header := make(metadata.MD)
			key := "Accept-Language"
			if tr, ok := transport.FromServerContext(ctx); ok {
				lang = language.Make(tr.RequestHeader().Get(key))
			}
			ii := i.Select(lang)
			header.Set(key, ii.Language().String())
			ctx = metadata.NewOutgoingContext(ctx, header)
			ctx = i18n.NewContext(ctx, ii)
			return handler(ctx, rq)
		}
	}
}
