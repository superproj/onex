// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package fake

import (
	"context"
	"sort"
	"sync"

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/fakeserver/store"
)

var (
	once sync.Once
	s    *datastore
)

// datastore 是 IStore 的一个具体实现.
type datastore struct {
	sync.RWMutex
	orders map[string]*model.OrderM
	maxIdx int64
}

// 确保 datastore 实现了 store.IStore 接口.
var _ store.IStore = (*datastore)(nil)

// NewStore 创建一个 store.IStore 类型的实例.
func NewStore(count int) *datastore {
	// 确保 s 只被初始化一次
	once.Do(func() {
		orders := FakeOrders(count)
		s = &datastore{
			orders: orders,
			maxIdx: int64(len(orders) - 1),
		}
	})

	return s
}

func (ds *datastore) DB(ctx context.Context) *gorm.DB {
	return nil
}

// TX is a method to execute a function inside a transaction, it takes a context and a function as parameters.
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return nil
}

// Orders 返回一个实现了 OrderStore 接口的实例.
func (ds *datastore) Orders() store.OrderStore {
	return newOrders(ds)
}

func (ds *datastore) GetIndex() int64 {
	ds.maxIdx++
	return ds.maxIdx
}

func (ds *datastore) List() []*model.OrderM {
	list := make([]*model.OrderM, 0, len(ds.orders))
	for _, order := range ds.orders {
		list = append(list, order)
	}

	sort.Sort(ByID(list))
	return list
}
