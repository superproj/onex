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

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/pkg/store/where"
)

// Singleton instance variables for the store.
var (
	once sync.Once
	// Global variable to hold the store instance
	S IStore
)

// IStore defines the interface for the store layer, specifying the methods that need to be implemented.
type IStore interface {
	DB(ctx context.Context) *gorm.DB
	TX(ctx context.Context, fn func(ctx context.Context) error) error
	Orders() OrderStore
}

// OrderStore defines the interface for order-related operations.
type OrderStore interface {
	// Create adds a new order to the store.
	Create(ctx context.Context, order *model.OrderM) error

	// Update modifies an existing order in the store.
	Update(ctx context.Context, order *model.OrderM) error

	// Delete removes orders from the store based on the provided conditions.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a single order from the store based on the provided conditions.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.OrderM, error)

	// List retrieves a list of orders from the store based on the provided conditions.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.OrderM, error)

	OrderExpansion
}

// OrderExpansion defines additional methods for order operations.
type OrderExpansion interface{}

// SetStore set the onex-fakeserver store instance in a global variable `S`.
// Direct use the global `S` is not recommended as this may make dependencies and calls unclear.
func SetStore(store IStore) {
	once.Do(func() {
		S = store
	})
}
