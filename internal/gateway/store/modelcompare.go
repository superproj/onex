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

// ModelCompareStore defines the modelcompare storage interface.
type ModelCompareStore interface {
	Create(ctx context.Context, ms *model.WalleModelCompare) error
	Delete(ctx context.Context, filters map[string]any) error
	Update(ctx context.Context, ms *model.WalleModelCompare) error
	Get(ctx context.Context, filters map[string]any) (*model.WalleModelCompare, error)
	List(ctx context.Context, namespace string, opts ...meta.ListOption) (int64, []*model.WalleModelCompare, error)
}

// modelCompareStore is a structure which implements the ModelCompareStore interface.
type modelCompareStore struct {
	ds *datastore
}

// newModelCompareStore creates a new modelCompareStore instance with provided datastore.
func newModelCompareStore(ds *datastore) *modelCompareStore {
	return &modelCompareStore{ds}
}

// db is an alias for d.ds.Core(ctx context.Context), a convenience method to get the core database instance.
func (d *modelCompareStore) db(ctx context.Context) *gorm.DB {
	return d.ds.Core(ctx)
}

// Create creates a new modelcompare record in the database.
func (d *modelCompareStore) Create(ctx context.Context, ms *model.WalleModelCompare) error {
	return d.db(ctx).Create(&ms).Error
}

// Delete deletes a modelcompare record from the database based on provided filters.
func (d *modelCompareStore) Delete(ctx context.Context, filters map[string]any) error {
	err := d.db(ctx).Where(filters).Delete(&model.WalleModelCompare{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// Update updates a modelcompare record in the database.
func (d *modelCompareStore) Update(ctx context.Context, ms *model.WalleModelCompare) error {
	return d.db(ctx).Save(ms).Error
}

// Get retrieves a single modelcompare record from the database based on provided filters.
func (d *modelCompareStore) Get(ctx context.Context, filters map[string]any) (*model.WalleModelCompare, error) {
	ms := &model.WalleModelCompare{}
	if err := d.db(ctx).Where(filters).First(&ms).Error; err != nil {
		return nil, err
	}

	return ms, nil
}

// List returns a list of modelcompare records according to the provided query conditions.
func (d *modelCompareStore) List(ctx context.Context, namespace string, opts ...meta.ListOption) (count int64, ret []*model.WalleModelCompare, err error) {
	los := meta.NewListOptions(opts...)
	if namespace != "" {
		los.Filters["namespace"] = namespace
	}
	ans := d.db(ctx).
		Where(los.Filters).
		Offset(los.Offset).
		Limit(los.Limit).
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count)

	return count, ret, ans.Error
}
