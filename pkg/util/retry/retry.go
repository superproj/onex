// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package retry

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	backoffSteps    = 10
	backoffFactor   = 1.25
	backoffDuration = 5
	backoffJitter   = 1.0
)

// Retry retries a given function with exponential backoff.
func Retry(fn wait.ConditionFunc, initialBackoffSec int) error {
	if initialBackoffSec <= 0 {
		initialBackoffSec = backoffDuration
	}
	backoffConfig := wait.Backoff{
		Steps:    backoffSteps,
		Factor:   backoffFactor,
		Duration: time.Duration(initialBackoffSec) * time.Second,
		Jitter:   backoffJitter,
	}
	retryErr := wait.ExponentialBackoff(backoffConfig, fn)
	if retryErr != nil {
		return retryErr
	}
	return nil
}

// Poll tries a condition func until it returns true, an error, or the timeout
// is reached.
func Poll(interval, timeout time.Duration, condition wait.ConditionFunc) error {
	return wait.Poll(interval, timeout, condition)
}

// PollImmediate tries a condition func until it returns true, an error, or the timeout
// is reached.
func PollImmediate(interval, timeout time.Duration, condition wait.ConditionFunc) error {
	return wait.PollImmediate(interval, timeout, condition)
}
