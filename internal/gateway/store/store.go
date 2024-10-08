// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

//go:generate mockgen -self_package github.com/superproj/onex/internal/gateway/store -destination mock_store.go -package store github.com/superproj/onex/internal/gateway/store IStore,ChainStore,MinerStore,MinerSetStore

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
	DB(ctx context.Context) *gorm.DB
	TX(context.Context, func(ctx context.Context) error) error
	Chains() ChainStore
	Miners() MinerStore
	MinerSets() MinerSetStore
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

// DB retrieves the current database instance from the context or returns the main instance.
func (ds *datastore) DB(ctx context.Context) *gorm.DB {
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

// Chains returns a ChainStore that interacts with datastore.
func (ds *datastore) Chains() ChainStore {
	return newChainStore(ds)
}

// MinerSets returns a MinerSetStore that interacts with datastore.
func (ds *datastore) MinerSets() MinerSetStore {
	return newMinerSetStore(ds)
}

// Miners returns a MinerStore that interacts with datastore.
func (ds *datastore) Miners() MinerStore {
	return newMinerStore(ds)
}
