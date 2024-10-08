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

// MinerSetStore defines the interface for managing minerSets in the database.
type MinerSetStore interface {
	// Create inserts a new minerSet into the database.
	Create(ctx context.Context, minerSet *model.MinerSetM) error

	// Update modifies an existing minerSet in the database.
	Update(ctx context.Context, minerSet *model.MinerSetM) error

	// Delete removes minerSets with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a minerSet with the specified options.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.MinerSetM, error)

	// List returns a list of minerSets with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.MinerSetM, error)

	MinerSetExpansion
}

// MinerSetExpansion defines additional methods for minerSet operations.
type MinerSetExpansion interface{}

// minerSetStore implements the MinerSetStore interface.
type minerSetStore struct {
	*genericstore.Store[model.MinerSetM]
}

// Ensure minerSetStore implements the MinerSetStore interface.
var _ MinerSetStore = (*minerSetStore)(nil)

// newMinerSetStore creates a new minerSetStore instance with provided datastore.
func newMinerSetStore(ds *datastore) *minerSetStore {
	return &minerSetStore{
		Store: genericstore.NewStore[model.MinerSetM](ds, onex.NewLogger()),
	}
}
