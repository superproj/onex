// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"context"
	"errors"

	"gorm.io/gorm"

	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/pkg/where"
	"github.com/superproj/onex/internal/usercenter/model"
)

// SecretStore defines the interface for managing secrets in the database.
type SecretStore interface {
	// Create inserts a new secret into the database.
	Create(ctx context.Context, secret *model.SecretM) error

	// Update modifies an existing secret in the database.
	Update(ctx context.Context, secret *model.SecretM) error

	// Get retrieves a secret by userID and secret name.
	Get(ctx context.Context, opts ...where.GetOption) (*model.SecretM, error)

	// List returns a list of secrets with the specified options.
	List(ctx context.Context, opts ...where.ListOption) (int64, []*model.SecretM, error)

	// Delete removes jobs by userID and a list of names.
	Delete(ctx context.Context, userID string, names []string) error
}

// secretStore is an implementation of the SecretStore interface
// that manages the secret model in a datastore.
type secretStore struct {
	ds *datastore
}

// newSecretStore initializes a new secretStore instance using the provided datastore.
func newSecretStore(ds *datastore) *secretStore {
	return &secretStore{ds}
}

// db is an alias for accessing the Core method of the datastore using the provided context.
func (d *secretStore) db(ctx context.Context, opts where.Where) *gorm.DB {
	return opts.Where(d.ds.Core(ctx))
}

// Create adds a new secret record in the datastore.
func (d *secretStore) Create(ctx context.Context, secret *model.SecretM, opts where.CreateOptions) error {
	return d.db(ctx, opts).Create(&secret).Error
}

// Delete removes a secret record from the datastore based on userID and name.
func (d *secretStore) Delete(ctx context.Context, opts ...where.DeleteOption) error {
	err := d.db(ctx).Where(model.SecretM{UserID: userID, Name: name}).Delete(&model.SecretM{}).Error
	// If error is not a "record not found" error, return the error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// Update modifies an existing secret record in the datastore.
func (d *secretStore) Update(ctx context.Context, secret *model.SecretM) error {
	return d.db(ctx).Save(secret).Error
}

// Get retrieves a secret record from the datastore based on userID and name.
func (d *secretStore) Get(ctx context.Context, userID string, name string) (*model.SecretM, error) {
	secret := &model.SecretM{}
	if err := d.db(ctx).Where(model.SecretM{UserID: userID, Name: name}).First(&secret).Error; err != nil {
		return nil, err
	}

	return secret, nil
}

// List returns a list of secret records that match the specified query conditions.
// It returns the total count of records and a slice of secret records.
// The query dynamically applies filters, offset, limit, and order, based on provided list options.
func (d *secretStore) List(ctx context.Context, userID string, opts ...meta.ListOption) (count int64, ret []*model.SecretM, err error) {
	// Initialize and configure list options
	o := meta.NewListOptions(opts...)
	// List secrets for all users by default.
	if userID != "" {
		o.Filters["user_id"] = userID
	}

	// Build query with filters, offset, limit, and order, and execute
	ans := d.db(ctx).
		Not("name", known.TemporaryKeyName).
		Where(o.Filters).
		Offset(o.Offset).
		Limit(o.Limit).
		Order("id desc").
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count)

	return count, ret, ans.Error
}
