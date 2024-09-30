// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/superproj/onex/pkg/store/where"
)

type DBProvider interface {
	DB(ctx context.Context) *gorm.DB
}

type Store[T any] struct {
	storage DBProvider
}

func NewStore[T any](storage DBProvider) *Store[T] {
	return &Store[T]{storage}
}

// db is an alias for accessing the Core method of the datastore using the provided context.
func (s *Store[T]) db(ctx context.Context, wheres ...where.Where) *gorm.DB {
	storage := s.storage.DB(ctx)
	for _, whr := range wheres {
		storage = whr.Where(storage)
	}

	return storage
}

func (s *Store[T]) Create(ctx context.Context, obj *T) error {
	return s.db(ctx).Create(obj).Error
}

func (s *Store[T]) Update(ctx context.Context, obj *T) error {
	return s.db(ctx).Save(obj).Error
}

func (s *Store[T]) Delete(ctx context.Context, opts *where.WhereOptions) error {
	err := s.db(ctx, opts).Delete(new(T)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *Store[T]) Get(ctx context.Context, opts *where.WhereOptions) (*T, error) {
	var obj T
	if err := s.db(ctx, opts).First(&obj).Error; err != nil {
		return nil, err
	}

	return &obj, nil
}

func (s *Store[T]) List(ctx context.Context, opts *where.WhereOptions) (count int64, ret []*T, err error) {
	err = s.db(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	return
}
