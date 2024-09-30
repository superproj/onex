// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package mysql

import (
	"context"
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/store"
)

// ProviderSet is a Wire provider set for creating a new datastore instance.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(store.IStore), new(*datastore)))

var (
	// Singleton instance variables for the datastore.
	once sync.Once
	// Global variable to hold the singleton datastore instance.
	S *datastore
)

// transactionKey is used as a key for storing the transaction context in context.Context.
type transactionKey struct{}

// datastore represents the main database instance and any additional instances.
type datastore struct {
	// core is the main database instance.
	// The `core` name indicates this is the main database.
	core *gorm.DB

	// Additional database instances can be added as needed.
	// In the example below, a fake database instance is added:
	// fake *gorm.DB
}

// Ensure that datastore implements the IStore interface.
var _ store.IStore = (*datastore)(nil)

// NewStore creates a new instance of datastore.
func NewStore(db *gorm.DB) *datastore {
	once.Do(func() {
		S = &datastore{db}
	})

	return S
}

// DB retrieves the current database instance from the context or returns the main instance.
func (ds *datastore) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if ok {
		return tx
	}

	return ds.core
}

// Orders returns a new instance of the OrderStore.
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// Orders returns a new instance of the OrderStore.
func (ds *datastore) Orders() store.OrderStore {
	return newOrders(ds)
}
