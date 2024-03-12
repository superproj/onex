// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package fake

import (
	"time"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/pkg/zid"
)

type ByID []*model.OrderM

func (o ByID) Len() int           { return len(o) }
func (o ByID) Less(i, j int) bool { return o[i].ID < o[j].ID }
func (o ByID) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

// FakeOrders returns fake order data.
func FakeOrders(count int) map[string]*model.OrderM {
	// init some order records
	orders := make(map[string]*model.OrderM)
	for i := 0; i < count; i++ {
		orderM := &model.OrderM{
			ID:        int64(i),
			OrderID:   zid.Order.New(uint64(i)),
			Customer:  gofakeit.Name(),
			Product:   gofakeit.Fruit(),
			Quantity:  gofakeit.Int64(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		orders[orderM.OrderID] = orderM
	}

	return orders
}
