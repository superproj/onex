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

// MinerStore defines the miner storage interface.
type MinerStore interface {
	Create(ctx context.Context, miner *model.MinerM) error
	Delete(ctx context.Context, filters map[string]any) error
	Update(ctx context.Context, miner *model.MinerM) error
	Get(ctx context.Context, filters map[string]any) (*model.MinerM, error)
	List(ctx context.Context, namespace string, opts ...meta.ListOption) (int64, []*model.MinerM, error)
}

type minerStore struct {
	ds *datastore
}

func newMinerStore(ds *datastore) *minerStore {
	return &minerStore{ds}
}

// db is alias for d.ds.Core(ctx context.Context).
func (d *minerStore) db(ctx context.Context) *gorm.DB {
	return d.ds.Core(ctx)
}

// Create creates a new miner record.
func (d *minerStore) Create(ctx context.Context, miner *model.MinerM) error {
	return d.db(ctx).Create(&miner).Error
}

// Delete delete an miner record.
func (d *minerStore) Delete(ctx context.Context, filters map[string]any) error {
	err := d.db(ctx).Where(filters).Delete(&model.MinerM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// Update updates an miner record.
func (d *minerStore) Update(ctx context.Context, miner *model.MinerM) error {
	return d.db(ctx).Save(miner).Error
}

// Get get an miner record.
func (d *minerStore) Get(ctx context.Context, filters map[string]any) (*model.MinerM, error) {
	miner := &model.MinerM{}
	if err := d.db(ctx).Where(filters).First(&miner).Error; err != nil {
		return nil, err
	}

	return miner, nil
}

// List return miners by specified query conditions.
func (d *minerStore) List(ctx context.Context, namespace string, opts ...meta.ListOption) (count int64, ret []*model.MinerM, err error) {
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
