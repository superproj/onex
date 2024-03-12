// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package biz

//go:generate mockgen -destination mock_biz.go -package biz github.com/superproj/onex/internal/fakeserver/biz IBiz

import (
	"github.com/superproj/onex/internal/fakeserver/biz/order"
	"github.com/superproj/onex/internal/fakeserver/store"
)

// IBiz 定义了 Biz 层需要实现的方法.
type IBiz interface {
	Orders() order.OrderBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	ds store.IStore
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

// Orders 返回一个实现了 OrderBiz 接口的实例.
func (b *biz) Orders() order.OrderBiz {
	return order.New(b.ds)
}
