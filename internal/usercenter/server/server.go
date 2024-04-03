// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package server

import (
	"context"
	"encoding/json"

	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"golang.org/x/text/language"

	onexmetrics "github.com/superproj/onex/internal/pkg/metrics"
	"github.com/superproj/onex/internal/pkg/middleware/authn/jwt"
	i18nmw "github.com/superproj/onex/internal/pkg/middleware/i18n"
	"github.com/superproj/onex/internal/pkg/middleware/logging"
	"github.com/superproj/onex/internal/pkg/middleware/tracing"
	"github.com/superproj/onex/internal/pkg/middleware/validate"
	"github.com/superproj/onex/internal/usercenter/locales"
	"github.com/superproj/onex/pkg/authn"
	"github.com/superproj/onex/pkg/i18n"
	"github.com/superproj/onex/pkg/log"
)

// ProviderSet defines a wire provider set.
var ProviderSet = wire.NewSet(NewServers, NewGRPCServer, NewHTTPServer, NewMiddlewares)

// NewServers is a wire provider function that creates and returns a slice of transport servers.
func NewServers(hs *http.Server, gs *grpc.Server) []transport.Server {
	return []transport.Server{hs, gs}
}

// NewMiddlewares return middlewares used by grpc and http server both.
func NewMiddlewares(logger krtlog.Logger, a authn.Authenticator, v validate.IValidator) []middleware.Middleware {
	return []middleware.Middleware{
		recovery.Recovery(
			recovery.WithHandler(func(ctx context.Context, rq, err any) error {
				data, _ := json.Marshal(rq)
				log.C(ctx).Errorw(err.(error), "Catching a panic", "rq", string(data))
				return nil
			}),
		),
		metrics.Server(
			metrics.WithSeconds(prom.NewHistogram(onexmetrics.KratosMetricSeconds)),
			metrics.WithRequests(prom.NewCounter(onexmetrics.KratosServerMetricRequests)),
		),
		i18nmw.Translator(i18n.WithLanguage(language.English), i18n.WithFS(locales.Locales)),
		// circuitbreaker.Client(),
		ratelimit.Server(),
		tracing.Server(),
		metadata.Server(),
		selector.Server(jwt.Server(a)).Match(NewWhiteListMatcher()).Build(),
		validate.Validator(v),
		logging.Server(logger),
	}
}
