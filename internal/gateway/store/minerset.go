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

	"github.com/superproj/onex/internal/gateway/model"
	"github.com/superproj/onex/internal/pkg/meta"
)

// MinerSetStore defines the minerset storage interface.
type MinerSetStore interface {
	Create(ctx context.Context, ms *model.MinerSetM) error
	Delete(ctx context.Context, filters map[string]any) error
	Update(ctx context.Context, ms *model.MinerSetM) error
	Get(ctx context.Context, filters map[string]any) (*model.MinerSetM, error)
	List(ctx context.Context, namespace string, opts ...meta.ListOption) (int64, []*model.MinerSetM, error)
}

// minerSetStore is a structure which implements the MinerSetStore interface.
type minerSetStore struct {
	ds *datastore
}

// newMinerSetStore creates a new minerSetStore instance with provided datastore.
func newMinerSetStore(ds *datastore) *minerSetStore {
	return &minerSetStore{ds}
}

// db is an alias for d.ds.Core(ctx context.Context), a convenience method to get the core database instance.
func (d *minerSetStore) db(ctx context.Context) *gorm.DB {
	return d.ds.Core(ctx)
}

// Create creates a new minerset record in the database.
func (d *minerSetStore) Create(ctx context.Context, ms *model.MinerSetM) error {
	return d.db(ctx).Create(&ms).Error
}

// Delete deletes a minerset record from the database based on provided filters.
func (d *minerSetStore) Delete(ctx context.Context, filters map[string]any) error {
	err := d.db(ctx).Where(filters).Delete(&model.MinerSetM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// Update updates a minerset record in the database.
func (d *minerSetStore) Update(ctx context.Context, ms *model.MinerSetM) error {
	return d.db(ctx).Save(ms).Error
}

// Get retrieves a single minerset record from the database based on provided filters.
func (d *minerSetStore) Get(ctx context.Context, filters map[string]any) (*model.MinerSetM, error) {
	ms := &model.MinerSetM{}
	if err := d.db(ctx).Where(filters).First(&ms).Error; err != nil {
		return nil, err
	}

	return ms, nil
}

// List returns a list of minerset records according to the provided query conditions.
func (d *minerSetStore) List(ctx context.Context, namespace string, opts ...meta.ListOption) (count int64, ret []*model.MinerSetM, err error) {
	los := meta.NewListOptions(opts...)
	if namespace != "" {
		los.Filters["namespace"] = namespace
	}
	ans := d.db(ctx).
		Where(los.Filters).
		Offset(los.Offset).
		Limit(los.Limit).
		Order("id desc").
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count)

	return count, ret, ans.Error
}
