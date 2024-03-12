// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package mysql

import (
	"sync"

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/store"
)

var (
	once sync.Once
	// 全局变量，保存已被初始化的 *datastore 实例.
	s *datastore
)

// datastore 是 IStore 的一个具体实现.
type datastore struct {
	db *gorm.DB
}

// 确保 datastore 实现了 store.IStore 接口.
var _ store.IStore = (*datastore)(nil)

// NewStore 创建一个 store.IStore 类型的实例.
func NewStore(db *gorm.DB) *datastore {
	// 确保 s 只被初始化一次
	once.Do(func() {
		s = &datastore{db}
	})

	return s
}

// Orders 返回一个实现了 OrderStore 接口的实例.
func (ds *datastore) Orders() store.OrderStore {
	return newOrders(ds.db)
}
