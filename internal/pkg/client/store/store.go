// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"sync"

	"github.com/google/wire"

	gwstore "github.com/superproj/onex/internal/gateway/store"
)

// ProviderSet is store providers.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(Interface), new(*datastore)))

var (
	once sync.Once
	S    *datastore
)

// Interface defines the storage interface.
type Interface interface {
	Gateway() gwstore.IStore
}

type datastore struct {
	gw gwstore.IStore
}

var _ Interface = (*datastore)(nil)

func (ds *datastore) Gateway() gwstore.IStore {
	return ds.gw
}

func NewStore(gw gwstore.IStore) *datastore {
	once.Do(func() {
		S = &datastore{gw: gw, uc: uc}
	})

	return S
}
