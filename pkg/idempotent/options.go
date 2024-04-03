// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package idempotent

import (
	"github.com/redis/go-redis/v9"
)

type Options struct {
	redis  redis.UniversalClient
	prefix string
	expire int
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd == nil {
			return
		}

		getOptionsOrSetDefault(options).redis = rd
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		if prefix == "" {
			return
		}

		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithExpire(min int) func(*Options) {
	return func(options *Options) {
		if min <= 0 {
			return
		}

		getOptionsOrSetDefault(options).expire = min
	}
}

// getOptionsOrSetDefault returns the provided options if they are not nil,
// otherwise it returns a default set of options.
func getOptionsOrSetDefault(options *Options) *Options {
	if options != nil {
		return options
	}

	return &Options{
		prefix: "idempotent",
		expire: 60,
	}
}
