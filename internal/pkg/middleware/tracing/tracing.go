// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package tracing

import (
	"context"

	krtlog "github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/superproj/onex/pkg/log"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	tracerName     string
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

// WithPropagator with tracer propagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opts *options) {
		opts.propagator = propagator
	}
}

// WithTracerProvider with tracer provider.
// Deprecated: use otel.SetTracerProvider(provider) instead.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *options) {
		opts.tracerProvider = provider
	}
}

// WithTracerName with tracer name.
func WithTracerName(tracerName string) Option {
	return func(opts *options) {
		opts.tracerName = tracerName
	}
}

// Server returns a new server middleware for OpenTelemetry.
func Server(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				var span trace.Span
				ctx, span = tracer.Start(ctx, tr.Operation(), tr.RequestHeader())

				traceID := span.SpanContext().TraceID().String()
				tr.ReplyHeader().Set("X-Trace-ID", traceID)
				ctx = log.WithContext(ctx, "trace.id", traceID)

				setServerSpan(ctx, span, rq)

				defer func() { tracer.End(ctx, span, reply, err) }()
			}
			return handler(ctx, rq)
		}
	}
}

// Client returns a new client middleware for OpenTelemetry.
func Client(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindClient, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				var span trace.Span
				ctx, span = tracer.Start(ctx, tr.Operation(), tr.RequestHeader())
				setClientSpan(ctx, span, rq)
				defer func() { tracer.End(ctx, span, reply, err) }()
			}
			return handler(ctx, rq)
		}
	}
}

// TraceID returns a traceid valuer.
func TraceID() krtlog.Valuer {
	return func(ctx context.Context) any {
		if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
			return span.TraceID().String()
		}
		return ""
	}
}

// SpanID returns a spanid valuer.
func SpanID() krtlog.Valuer {
	return func(ctx context.Context) any {
		if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
			return span.SpanID().String()
		}
		return ""
	}
}
