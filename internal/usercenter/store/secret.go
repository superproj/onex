// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"context"

	"github.com/superproj/onex/internal/usercenter/model"
	genericstore "github.com/superproj/onex/pkg/store"
	"github.com/superproj/onex/pkg/store/logger/onex"
	"github.com/superproj/onex/pkg/store/where"
)

// SecretStore defines the interface for managing secrets in the database.
type SecretStore interface {
	// Create inserts a new secret into the database.
	Create(ctx context.Context, secret *model.SecretM) error

	// Update modifies an existing secret in the database.
	Update(ctx context.Context, secret *model.SecretM) error

	// Delete removes secrets with the specified options.
	Delete(ctx context.Context, opts *where.WhereOptions) error

	// Get retrieves a secret with the specified options.
	Get(ctx context.Context, opts *where.WhereOptions) (*model.SecretM, error)

	// List returns a list of secrets with the specified options.
	List(ctx context.Context, opts *where.WhereOptions) (int64, []*model.SecretM, error)

	SecretExpansion
}

// SecretExpansion defines additional methods for secret operations.
type SecretExpansion interface{}

// secretStore implements the SecretStore interface.
type secretStore struct {
	*genericstore.Store[model.SecretM]
}

// Ensure secretStore implements the SecretStore interface.
var _ SecretStore = (*secretStore)(nil)

// newSecretStore creates a new secretStore instance with provided datastore.
func newSecretStore(ds *datastore) *secretStore {
	return &secretStore{
		Store: genericstore.NewStore[model.SecretM](ds, onex.NewLogger()),
	}
}
