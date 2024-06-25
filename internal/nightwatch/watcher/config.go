// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package watcher provides functions used by all watchers.
package watcher

import (
	"github.com/superproj/onex/internal/pkg/client/store"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

// Config aggregates the configurations of all watchers and serves as a configuration aggregator.
type Config struct {
	// The purpose of nightwatch is to handle asynchronous tasks on the onex platform
	// in a unified manner, so a store aggregation type is needed here.
	Store store.Interface

	// Client is the client for onex-apiserver.
	Client clientset.Interface

	// Then maximum concurrency event of user watcher.
	UserWatcherMaxWorkers int64
}
