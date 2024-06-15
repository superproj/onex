// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"context"
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is a set that initializes the Store and binds it to the IStore interface.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

var (
	once sync.Once
	S    *datastore
)

type transactionKey struct{}

// IStore is an interface defining the required methods for a Store.
type IStore interface {
	TX(context.Context, func(ctx context.Context) error) error
	ModelCompares() ModelCompareStore
}

// datastore is a concrete implementation of IStore interface.
type datastore struct {
	// core is the main database, use the name `core` to indicate that this is the main database.
	core *gorm.DB

	// You can add more database instances as needed. For example, a fake database instance is added below:
	// fake *gorm.DB
}

// Verify that datastore implements IStore interface.
var _ IStore = (*datastore)(nil)

// NewStore initializes a new datastore by using the given gorm.DB and returns it.
func NewStore(db *gorm.DB) *datastore {
	once.Do(func() {
		S = &datastore{db}
	})

	return S
}

// Core returns the core gorm.DB from the datastore. If there is an ongoing transaction,
// the transaction's gorm.DB is returned instead.
func (ds *datastore) Core(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if ok {
		return tx
	}

	return ds.core
}

// FakeDB is used to demonstrate multiple database instances. It returns a nil gorm.DB, indicating a fake database.
func (ds *datastore) FakeDB(ctx context.Context) *gorm.DB { return nil }

// TX is a method to execute a function inside a transaction, it takes a context and a function as parameters.
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// ModelCompares returns a ModelCompareStore that interacts with datastore.
func (ds *datastore) ModelCompares() ModelCompareStore {
	return newModelCompareStore(ds)
}
