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

// MinerStore defines the interface for managing miners in the database.
type MinerStore interface {
	// Create inserts a new miner into the database.
	Create(ctx context.Context, miner *model.MinerM) error

	// Update modifies an existing miner in the database.
	Update(ctx context.Context, miner *model.MinerM) error

	// Delete removes miners with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a miner with the specified options.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.MinerM, error)

	// List returns a list of miners with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.MinerM, error)

	MinerExpansion
}

// MinerExpansion defines additional methods for miner operations.
type MinerExpansion interface{}

// minerStore implements the MinerStore interface.
type minerStore struct {
	*genericstore.Store[model.MinerM]
}

// Ensure minerStore implements the MinerStore interface.
var _ MinerStore = (*minerStore)(nil)

// newMinerStore creates a new minerStore instance with provided datastore.
func newMinerStore(ds *datastore) *minerStore {
	return &minerStore{
		Store: genericstore.NewStore[model.MinerM](ds, onex.NewLogger()),
	}
}
