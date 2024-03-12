// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	"context"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/internal/gateway/store"
)

// ICleaner interface defines the methods required for a cleaner implementation:.
type ICleaner interface {
	// Name returns the name of the cleaner.
	Name() string

	// Sync performs the synchronization operation.
	Sync(ctx context.Context) error

	// Initialize initializes the cleaner with the provided client and store client.
	Initialize(client client.Client, storeClient store.IStore)
}

// Cleaner is a struct that represents a set of cleaners used to clean deleted resources from onex db.
type Cleaner struct {
	cleaners []ICleaner
}

// newCleaner return a cleaner set used to clean deleted resources from onex db.
func newCleaner(client client.Client, ds store.IStore, cleaners ...ICleaner) *Cleaner {
	for _, cleaner := range cleaners {
		cleaner.Initialize(client, ds)
	}

	return &Cleaner{cleaners}
}

// Start starts the Cleaner and runs the SyncAll() method periodically.
func (c *Cleaner) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.SyncAll(ctx)
		}
	}
}

// SyncAll runs the Sync() method of all registered cleaners.
func (c *Cleaner) SyncAll(ctx context.Context) {
	for _, cleaner := range c.cleaners {
		go func(cleaner ICleaner) {
			if err := cleaner.Sync(ctx); err != nil {
				klog.ErrorS(err, "Failed to sync", "cleaner", cleaner.Name())
			}
		}(cleaner)
	}
}
