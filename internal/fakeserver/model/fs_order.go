// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package model

import "time"

type OrderM struct {
	ID        int64     `gorm:"column:id;primary_key"` //
	OrderID   string    `gorm:"column:order_id"`       //
	Customer  string    `gorm:"column:customer"`       //
	Product   string    `gorm:"column:product"`        //
	Quantity  int64     `gorm:"column:quantity"`       //
	CreatedAt time.Time `gorm:"column:created_at"`     //
	UpdatedAt time.Time `gorm:"column:updated_at"`     //
}

// TableName sets the insert table name for this struct type.
func (o *OrderM) TableName() string {
	return "fs_order"
}
