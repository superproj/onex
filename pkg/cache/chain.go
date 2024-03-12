// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// chainKeyValue represents the key-value pair with TTL and cache ID.
type chainKeyValue[T any] struct {
	key   any
	value T
	ttl   time.Duration
	id    string
}

type cacheWrapper[T any] struct {
	Cache[T]
	id string
}

// ChainCache represents the configuration needed by a cache aggregator.
type ChainCache[T any] struct {
	caches     []*cacheWrapper[T]
	setChannel chan *chainKeyValue[T]
}

// NewChain instantiates a new cache aggregator.
func NewChain[T any](caches ...Cache[T]) *ChainCache[T] {
	wrappers := make([]*cacheWrapper[T], 0, len(caches))
	for _, c := range caches {
		wrappers = append(wrappers, &cacheWrapper[T]{
			Cache: c,
			id:    uuid.New().String(),
		})
	}
	chain := &ChainCache[T]{
		caches:     wrappers,
		setChannel: make(chan *chainKeyValue[T], 10000),
	}

	go chain.Sync()

	return chain
}

// Sync synchronizes a value in available caches, until a given cache layer.
func (c *ChainCache[T]) Sync() {
	for item := range c.setChannel {
		for _, cache := range c.caches {
			if item.id == cache.id {
				break
			}

			cache.SetWithTTL(context.Background(), item.key, item.value, item.ttl)
		}
	}
}

// Get returns the obj stored in cache if it exists.
func (c *ChainCache[T]) Get(ctx context.Context, key any) (T, error) {
	obj, _, err := c.GetWithTTL(ctx, key)
	return obj, err
}

// GetWithTTL returns the object and its TTL from the first cache where it exists.
func (c *ChainCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	var obj T
	var err error
	var ttl time.Duration

	for _, cache := range c.caches {
		obj, ttl, err = cache.GetWithTTL(ctx, key)
		if err == nil {
			// Set the value back until this cache layer.
			c.setChannel <- &chainKeyValue[T]{key, obj, ttl, cache.id}
			return obj, ttl, nil
		}
	}

	return obj, ttl, err
}

// Set sets a value in available caches.
func (c *ChainCache[T]) Set(ctx context.Context, key any, obj T) error {
	errs := []error{}
	for _, cache := range c.caches {
		if err := cache.Set(ctx, key, obj); err != nil {
			errs = append(errs, fmt.Errorf("unable to set item into cache %w", err))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	errStr := ""
	for k, v := range errs {
		errStr += fmt.Sprintf("error %d of %d: %v", k+1, len(errs), v.Error())
	}
	return errors.New(errStr)
}

// SetWithTTL sets a value in available caches with a specified TTL.
func (c *ChainCache[T]) SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error {
	errs := []error{}
	for _, cache := range c.caches {
		if err := cache.SetWithTTL(ctx, key, obj, ttl); err != nil {
			errs = append(errs, fmt.Errorf("unable to set item into cache: %w", err))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	errStr := ""
	for k, v := range errs {
		errStr += fmt.Sprintf("error %d of %d: %v", k+1, len(errs), v.Error())
	}
	return errors.New(errStr)
}

// Del removes a value from all available caches.
func (c *ChainCache[T]) Del(ctx context.Context, key any) error {
	for _, cache := range c.caches {
		cache.Del(ctx, key)
	}

	return nil
}

// Clear resets all cache data.
func (c *ChainCache[T]) Clear(ctx context.Context) error {
	for _, cache := range c.caches {
		cache.Clear(ctx)
	}

	return nil
}

// Wait waits for all cache operations to complete.
func (c *ChainCache[T]) Wait(ctx context.Context) {
	for _, cache := range c.caches {
		cache.Wait(ctx)
	}
}
