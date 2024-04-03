// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package model

import (
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/zid"
)

// AfterCreate runs after creating a OrderM database record and updates the OrderID field.
func (o *OrderM) AfterCreate(tx *gorm.DB) (err error) {
	o.OrderID = zid.Order.New(uint64(o.ID)) // Generate and set a new order ID.

	return tx.Save(o).Error // Save the updated order record to the database.
}
