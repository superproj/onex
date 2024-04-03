// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package gocache

import (
	"context"
	"time"

	"github.com/superproj/onex/pkg/cache/store"
)

const (
	// GoCacheType represents the storage type as a string value.
	GoCacheType = "go-cache"
	// GoCacheTagPattern represents the tag pattern to be used as a key in specified storage.
	GoCacheTagPattern = "gocache_tag_%s"
)

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client.
type GoCacheClientInterface interface {
	Get(k string) (any, bool)
	GetWithExpiration(k string) (any, time.Time, bool)
	Set(k string, x any, d time.Duration)
	Delete(k string)
	Flush()
}

// GoCacheStore is a store for GoCache (memory) library.
type GoCacheStore struct {
	client GoCacheClientInterface
}

// NewGoCache creates a new store to GoCache (memory) library instance.
func NewGoCache(client GoCacheClientInterface) *GoCacheStore {
	return &GoCacheStore{
		client: client,
	}
}

// Get returns data stored from a given key.
func (s *GoCacheStore) Get(_ context.Context, key any) (any, error) {
	keyStr := key.(string)
	value, exists := s.client.Get(keyStr)
	if !exists {
		return value, store.ErrKeyNotFound
	}

	return value, nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL.
func (s *GoCacheStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	data, t, exists := s.client.GetWithExpiration(key.(string))
	if !exists {
		return data, 0, store.ErrKeyNotFound
	}
	duration := time.Until(t)
	return data, duration, nil
}

// Set defines data in GoCache memoey cache for given key identifier.
func (s *GoCacheStore) Set(ctx context.Context, key any, value any) error {
	s.client.Set(key.(string), value, 0)
	return nil
}

func (s *GoCacheStore) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error {
	s.client.Set(key.(string), value, ttl)
	return nil
}

// Delete removes data in GoCache memoey cache for given key identifier.
func (s *GoCacheStore) Del(_ context.Context, key any) error {
	s.client.Delete(key.(string))
	return nil
}

// Clear resets all data in the store.
func (s *GoCacheStore) Clear(_ context.Context) error {
	s.client.Flush()
	return nil
}

func (s *GoCacheStore) Wait(_ context.Context) {
}
