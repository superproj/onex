// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/fakeserver/store"
	"github.com/superproj/onex/internal/pkg/meta"
)

// OrderStore 接口的实现.
type orders struct {
	db *gorm.DB
}

// 确保 orders 实现了 OrderStore 接口.
var _ store.OrderStore = (*orders)(nil)

func newOrders(db *gorm.DB) *orders {
	return &orders{db}
}

// Create 插入一条 order 记录.
func (o *orders) Create(ctx context.Context, order *model.OrderM) error {
	return o.db.Create(&order).Error
}

// Get 根据用户名查询指定 order 的数据库记录.
func (o *orders) Get(ctx context.Context, orderID string) (*model.OrderM, error) {
	var order model.OrderM
	if err := o.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// Update 更新一条 order 数据库记录.
func (o *orders) Update(ctx context.Context, order *model.OrderM) error {
	return o.db.Save(order).Error
}

// List 根据 offset 和 limit 返回 order 列表.
func (o *orders) List(ctx context.Context, opts ...meta.ListOption) (count int64, ret []*model.OrderM, err error) {
	options := meta.NewListOptions(opts...)

	ans := o.db.
		Where(options.Filters).
		Offset(options.Offset).
		Limit(options.Limit).
		Order("id desc").
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count)

	return count, ret, ans.Error
}

// Delete 根据 orderID 删除数据库 order 记录.
func (o *orders) Delete(ctx context.Context, orderID string) error {
	err := o.db.Where("order_id = ?", orderID).Delete(&model.OrderM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}
