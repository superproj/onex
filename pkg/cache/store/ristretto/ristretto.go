// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package ristretto

import (
	"context"
	"fmt"
	"time"

	"github.com/superproj/onex/pkg/cache/store"
)

const (
	// RistrettoType represents the storage type as a string value.
	RistrettoType = "ristretto"
	// RistrettoTagPattern represents the tag pattern to be used as a key in specified storage.
	RistrettoTagPattern = "gocache_tag_%s"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client.
type RistrettoClientInterface interface {
	Get(key any) (any, bool)
	Set(key, value any, cost int64) bool
	SetWithTTL(key, value any, cost int64, ttl time.Duration) bool
	Del(key any)
	Clear()
	Wait()
}

// RistrettoStore is a store for Ristretto (memory) library.
type RistrettoStore struct {
	client RistrettoClientInterface
}

// NewRistretto creates a new store to Ristretto (memory) library instance.
func NewRistretto(client RistrettoClientInterface) *RistrettoStore {
	return &RistrettoStore{
		client: client,
	}
}

// Get returns data stored from a given key.
func (s *RistrettoStore) Get(_ context.Context, key any) (any, error) {
	var err error

	value, exists := s.client.Get(key)
	if !exists {
		err = store.ErrKeyNotFound
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL.
func (s *RistrettoStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, 0, err
}

// Set defines data in Ristretto memory cache for given key identifier.
func (s *RistrettoStore) Set(_ context.Context, key any, value any) error {
	if set := s.client.Set(key, value, 0); !set {
		return fmt.Errorf("an error has occurred while setting value '%v' on key '%v'", value, key)
	}

	return nil
}

func (s *RistrettoStore) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error {
	if set := s.client.SetWithTTL(key, value, 0, ttl); !set {
		return fmt.Errorf("an error has occurred while setting value '%v' on key '%v'", value, key)
	}

	return nil
}

// Delete removes data in Ristretto memory cache for given key identifier.
func (s *RistrettoStore) Del(_ context.Context, key any) error {
	s.client.Del(key)
	return nil
}

// Clear resets all data in the store.
func (s *RistrettoStore) Clear(_ context.Context) error {
	s.client.Clear()
	return nil
}

func (s *RistrettoStore) Wait(_ context.Context) {
	s.client.Wait()
}
