// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:dupl
package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/superproj/onex/pkg/log"
)

// Server is an server logging middleware.
func Server(logger krtlog.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if tr, ok := transport.FromServerContext(ctx); ok {
				kind = tr.Kind().String()
				operation = tr.Operation()
			}
			reply, err = handler(ctx, rq)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			_ = log.C(ctx).Log(level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"args", extractArgs(rq),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// Client is a client logging middleware.
func Client(logger krtlog.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if tr, ok := transport.FromClientContext(ctx); ok {
				kind = tr.Kind().String()
				operation = tr.Operation()
			}
			reply, err = handler(ctx, rq)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			_ = log.C(ctx).Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"args", extractArgs(rq),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the rq.
func extractArgs(rq any) string {
	if stringer, ok := rq.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", rq)
}

// extractError returns the string of the error.
func extractError(err error) (krtlog.Level, string) {
	if err != nil {
		return krtlog.LevelError, fmt.Sprintf("%+v", err)
	}
	return krtlog.LevelInfo, ""
}
