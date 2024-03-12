// Copyright 2024 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	known "github.com/superproj/onex/internal/pkg/known/toyblc"
	"github.com/superproj/onex/pkg/log"
)

// TraceID 是一个 Gin 中间件，用来在每一个 HTTP 请求的 context, response 中注入 `Trace-ID` 键值对.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否有 `Trace-ID`，如果有则复用，没有则新建
		traceID := c.Request.Header.Get(known.TraceIDKey)

		if traceID == "" {
			traceID = uuid.New().String()
			// 将 Trace-ID 保存在 HTTP 请求头中，Header 的键为 `Trace-ID`
			c.Request.Header.Set(known.TraceIDKey, traceID)
		}

		// 将 Trace-ID 保存在 HTTP 返回头中，Header 的键为 `Trace-ID`
		c.Writer.Header().Set(known.TraceIDKey, traceID)

		// 将 `trace.id` 保存在 gin.Context 中，方便后边程序使用
		// 使用 `trace.id` 而没有使用 `known.TraceIDKey`，是为了跟其它组件保持一致
		c.Set("trace.id", traceID)

		_ = log.WithContext(c, "trace.id", traceID)

		c.Next()
	}
}
