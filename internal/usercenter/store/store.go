// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

//go:generate mockgen -self_package github.com/superproj/onex/internal/usercenter/store -destination mock_store.go -package store github.com/superproj/onex/internal/usercenter/store IStore,SecretStore,UserStore

import (
	"context"
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is a Wire provider set that initializes new datastore instances
// and binds the IStore interface to the actual datastore type.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

// Singleton instance variables.
var (
	once sync.Once
	S    *datastore
)

// transactionKey is an unique key used in context to store
// transaction instances to be shared between multiple operations.
type transactionKey struct{}

// IStore is an interface that represents methods
// required to be implemented by a Store implementation.
type IStore interface {
	DB(ctx context.Context) *gorm.DB
	TX(context.Context, func(ctx context.Context) error) error
	Users() UserStore
	Secrets() SecretStore
}

// datastore is an implementation of IStore that provides methods
// to perform operations on a database using gorm library.
type datastore struct {
	// core is the main database instance.
	// The `core` name indicates this is the main database.
	core *gorm.DB

	// Additional database instances can be added as needed.
	// In the example below, a fake database instance is added:
	// fake *gorm.DB
}

// Ensure datastore implements IStore.
var _ IStore = (*datastore)(nil)

// NewStore initializes a new datastore instance using the provided DB gorm instance.
// It also creates a singleton instance for the datastore.
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

// FakeDB is an empty method to demonstrate how to handle multiple database instances.
// This method should be implemented to return an actual fake DB instance.
func (ds *datastore) FakeDB(ctx context.Context) *gorm.DB { return nil }

// TX starts a transaction using the main DB context
// and passes the transactional context to the provided function.
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// Users returns an initialized instance of UserStore.
func (ds *datastore) Users() UserStore {
	return newUserStore(ds)
}

// Secrets returns an initialized instance of SecretStore.
func (ds *datastore) Secrets() SecretStore {
	return newSecretStore(ds)
}
