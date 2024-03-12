// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cache

import (
	"context"
	"time"

	"github.com/superproj/onex/pkg/cache/store"
)

// DelegateCache is a representative cache used to represent the store. By representing
// the store, different stores can be encapsulated into a unified cache and perform some unified operations.
type DelegateCache[T any] struct {
	store store.Store
}

// New instantiates a new delegate cache entry.
func New[T any](store store.Store) *DelegateCache[T] {
	return &DelegateCache[T]{store: store}
}

// Get returns the obj stored in cache if it exists.
func (c *DelegateCache[T]) Get(ctx context.Context, key any) (T, error) {
	value, err := c.store.Get(ctx, keyFunc(key))
	if err != nil {
		return *new(T), err
	}

	if v, ok := value.(T); ok {
		return v, nil
	}

	return *new(T), nil
}

// GetWithTTL returns the obj stored in cache and its corresponding TTL.
func (c *DelegateCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	value, duration, err := c.store.GetWithTTL(ctx, keyFunc(key))
	if err != nil {
		return *new(T), duration, err
	}

	if v, ok := value.(T); ok {
		return v, duration, nil
	}

	return *new(T), duration, nil
}

// Set populates the cache item using the given key.
func (c *DelegateCache[T]) Set(ctx context.Context, key any, obj T) error {
	return c.store.Set(ctx, keyFunc(key), obj)
}

// SetWithTTL populates the cache item using the given key with a specified TTL.
func (c *DelegateCache[T]) SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error {
	return c.store.SetWithTTL(ctx, keyFunc(key), obj, ttl)
}

// Del removes the cache item using the given key.
func (c *DelegateCache[T]) Del(ctx context.Context, key any) error {
	return c.store.Del(ctx, keyFunc(key))
}

// Clear resets all cache data.
func (c *DelegateCache[T]) Clear(ctx context.Context) error {
	return c.store.Clear(ctx)
}

// Wait waits for all cache operations to complete.
func (c *DelegateCache[T]) Wait(ctx context.Context) {
	c.store.Wait(ctx)
}
