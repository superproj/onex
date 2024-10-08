// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package fake

import (
	"context"

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/fakeserver/store"
	"github.com/superproj/onex/internal/pkg/zid"
	"github.com/superproj/onex/pkg/store/where"
)

// OrderStore 接口的实现.
type orders struct {
	ds *datastore
}

// 确保 orders 实现了 OrderStore 接口.
var _ store.OrderStore = (*orders)(nil)

func newOrders(ds *datastore) *orders {
	return &orders{ds}
}

// Create 插入一条 order 记录.
func (o *orders) Create(ctx context.Context, order *model.OrderM) error {
	o.ds.Lock()
	defer o.ds.Unlock()

	order.ID = o.ds.GetIndex()
	order.OrderID = zid.Order.New(uint64(order.ID))

	o.ds.orders[order.OrderID] = order
	return nil
}

// Get 根据用户名查询指定 order 的数据库记录.
func (o *orders) Get(ctx context.Context, opts *where.WhereOptions) (*model.OrderM, error) {
	o.ds.Lock()
	defer o.ds.Unlock()

	orderID := opts.Filters["order_id"].(string)
	order, ok := o.ds.orders[orderID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	return order, nil
}

// Update 更新一条 order 数据库记录.
func (o *orders) Update(ctx context.Context, order *model.OrderM) error {
	o.ds.Lock()
	defer o.ds.Unlock()

	if _, ok := o.ds.orders[order.OrderID]; !ok {
		return gorm.ErrRecordNotFound
	}

	o.ds.orders[order.OrderID] = order

	return nil
}

// List 根据 offset 和 limit 返回 order 列表.
func (o *orders) List(ctx context.Context, opts *where.WhereOptions) (count int64, ret []*model.OrderM, err error) {
	o.ds.Lock()
	defer o.ds.Unlock()

	all := o.ds.List()

	offset, limit := opts.Offset, opts.Limit
	if offset >= len(all) {
		return 0, []*model.OrderM{}, nil
	}

	if offset+limit > len(all) {
		limit = len(all) - offset
	}

	return int64(len(all)), all[offset : offset+limit], nil
}

// Delete 根据 orderID 删除数据库 order 记录.
func (o *orders) Delete(ctx context.Context, opts *where.WhereOptions) error {
	o.ds.Lock()
	defer o.ds.Unlock()

	orderID := opts.Filters["order_id"].(string)
	delete(s.orders, orderID)

	return nil
}
