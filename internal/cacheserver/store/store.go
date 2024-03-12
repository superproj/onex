// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

//go:generate mockgen -destination mock_store.go -package store github.com/superproj/onex/internal/cacheserver/store IStore

import (
	"github.com/dgraph-io/ristretto"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/cacheserver/store/secret"
	"github.com/superproj/onex/pkg/cache"
	ristrettostore "github.com/superproj/onex/pkg/cache/store/ristretto"
)

// ProviderSet is a Wire provider set that initializes new datastore instances
// and binds the IStore interface to the actual datastore type.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

// IStore defines the methods that need to be implemented by the Store layer.
type IStore interface {
	Secrets() *cache.ChainCache[any]
}

// datastore is used to implement the IStore interface.
type datastore struct {
	db     *gorm.DB
	local  cache.Cache[any]
	secret *cache.ChainCache[any]
}

// NewStore creates a new instance of the datastore.
func NewStore(db *gorm.DB, disable bool) *datastore {
	caches := make([]cache.Cache[any], 0)

	// ristretto configuration has been verified in the application, so this is a legal
	// configuration and no error will be returned.
	riscache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})

	risstore := ristrettostore.NewRistretto(riscache)
	local := cache.New[any](risstore)
	if !disable {
		caches = append(caches, local)
	}

	mysqlStore := secret.New(db)
	caches = append(caches, cache.New[any](mysqlStore))

	return &datastore{
		db:     db,
		local:  local,
		secret: cache.NewChain[any](caches...),
	}
}

// Secrets returns a ChainCache for managing secrets.
func (ds *datastore) Secrets() *cache.ChainCache[any] {
	return ds.secret
}
