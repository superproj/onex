// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package tracing wires OpenTelemetry tracing into controller processes.
package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Setup initializes the global OpenTelemetry tracer provider, exporting spans to
// the given OTLP endpoint. It returns an http.RoundTripper wrapper that
// instruments client-go requests with spans. When endpoint is empty, tracing is
// disabled and the returned wrapper is nil.
func Setup(ctx context.Context, serviceName, endpoint string) (func(http.RoundTripper) http.RoundTripper, error) {
	if endpoint == "" {
		return nil, nil
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return func(rt http.RoundTripper) http.RoundTripper {
		return otelhttp.NewTransport(rt)
	}, nil
}
