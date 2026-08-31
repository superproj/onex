// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package ratelimiter

import (
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DefaultControllerRateLimiter returns the rate limiter shared by all OneX controllers.
func DefaultControllerRateLimiter() workqueue.TypedRateLimiter[reconcile.Request] {
	return workqueue.NewTypedMaxOfRateLimiter(
		// Per-item exponential backoff: retries start at 200ms and grow to at most 1h,
		// so a persistently failing object is retried less and less frequently.
		workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](200*time.Millisecond, 1*time.Hour),
		// Overall token bucket: 5000 qps with a burst of 10000. This is the global,
		// not per-item, cap on retry throughput.
		&workqueue.TypedBucketRateLimiter[reconcile.Request]{Limiter: rate.NewLimiter(rate.Limit(5000), 10000)},
	)
}
