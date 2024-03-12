// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/superproj/onex/internal/usercenter/model"
)

// secretStore provides methods to interact with secrets in the database.
type secretStore struct {
	db *gorm.DB
}

// New creates a new instance of secretStore.
func New(db *gorm.DB) *secretStore {
	return &secretStore{db: db}
}

// Get retrieves a secret by its key.
func (s *secretStore) Get(ctx context.Context, key any) (any, error) {
	secret := &model.SecretM{}
	if err := s.db.Where(model.SecretM{SecretID: key.(string)}).First(&secret).Error; err != nil {
		return nil, err
	}
	return secret, nil
}

// GetWithTTL retrieves a secret by its key along with its time to live (TTL).
func (s *secretStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	ttl := time.Until(time.Unix(value.(*model.SecretM).Expires, 0))

	return value, ttl, nil
}

// Set stores a secret with the given key and value.
func (s *secretStore) Set(ctx context.Context, key any, value any) error {
	secret := value.(*model.SecretM)

	err := s.db.Where(model.SecretM{SecretID: secret.SecretID}).
		Assign(secret).
		FirstOrCreate(secret).
		Error

	return err
}

// SetWithTTL stores a secret with the given key, value, and time to live (TTL).
func (s *secretStore) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error {
	return s.Set(ctx, key, value)
}

// Del deletes a secret by its key.
func (s *secretStore) Del(ctx context.Context, key any) error {
	err := s.db.Where(model.SecretM{SecretID: key.(string)}).Delete(&model.SecretM{}).Error
	// If error is not a "record not found" error, return the error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// Clear is not supported for secretStore.
func (s *secretStore) Clear(ctx context.Context) error {
	return nil
}

// Wait waits for all operations to finish.
func (s *secretStore) Wait(ctx context.Context) {}
