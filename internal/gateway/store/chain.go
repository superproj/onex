// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"context"

	"github.com/superproj/onex/internal/gateway/model"
	genericstore "github.com/superproj/onex/pkg/store"
	"github.com/superproj/onex/pkg/store/logger/onex"
	"github.com/superproj/onex/pkg/store/where"
)

// ChainStore defines the interface for managing chains in the database.
type ChainStore interface {
	// Create inserts a new chain into the database.
	Create(ctx context.Context, chain *model.ChainM) error

	// Update modifies an existing chain in the database.
	Update(ctx context.Context, chain *model.ChainM) error

	// Delete removes chains with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a chain with the specified options.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.ChainM, error)

	// List returns a list of chains with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.ChainM, error)

	ChainExpansion
}

// ChainExpansion defines additional methods for chain operations.
type ChainExpansion interface{}

// chainStore implements the ChainStore interface.
type chainStore struct {
	*genericstore.Store[model.ChainM]
}

// Ensure chainStore implements the ChainStore interface.
var _ ChainStore = (*chainStore)(nil)

// newChainStore creates a new chainStore instance with provided datastore.
func newChainStore(ds *datastore) *chainStore {
	return &chainStore{
		Store: genericstore.NewStore[model.ChainM](ds, onex.NewLogger()),
	}
}
