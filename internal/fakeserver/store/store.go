// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

//go:generate mockgen -destination mock_store.go -package store github.com/superproj/onex/internal/fakeserver/store IStore,OrderStore

import (
	"context"
	"sync"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/pkg/meta"
)

// IStore 定义了 Store 层需要实现的方法.
type IStore interface {
	Orders() OrderStore
}

// OrderStore 定义了 order 模块在 store 层所实现的方法.
type OrderStore interface {
	Create(ctx context.Context, order *model.OrderM) error
	Get(ctx context.Context, orderID string) (*model.OrderM, error)
	Update(ctx context.Context, order *model.OrderM) error
	List(ctx context.Context, opts ...meta.ListOption) (int64, []*model.OrderM, error)
	Delete(ctx context.Context, orderID string) error
}

var (
	once sync.Once
	// 全局变量，方便其它包直接调用已初始化好的 S 实例.
	S IStore
)

// SetStore set the onex-fakeserver store instance in a global variable `S`.
// Direct use the global `S` is not recommended as this may make dependencies and calls unclear.
func SetStore(store IStore) {
	once.Do(func() {
		S = store
	})
}
