// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cache

import (
	"context"
	"sync"
	"time"
)

// loadableKeyValue represents a key-value pair to be loaded into the cache.
type loadableKeyValue[T any] struct {
	key   any
	value T
}

// LoadFunction is a function type for loading data into the cache.
type LoadFunction[T any] func(ctx context.Context, key any) (T, error)

// LoadableCache represents a cache that uses a function to load data.
type LoadableCache[T any] struct {
	loadFunc   LoadFunction[T]
	cache      Cache[T]
	setChannel chan *loadableKeyValue[T]
	wg         *sync.WaitGroup
}

// NewLoadable instanciates a new cache that uses a function to load data.
func NewLoadable[T any](loadFunc LoadFunction[T], cache Cache[T]) *LoadableCache[T] {
	loadable := &LoadableCache[T]{
		loadFunc:   loadFunc,
		cache:      cache,
		setChannel: make(chan *loadableKeyValue[T], 10000),
		wg:         &sync.WaitGroup{},
	}

	loadable.wg.Add(1)
	go loadable.Sync()

	return loadable
}

// Sync processes items in the setChannel and sets them in the cache.
func (c *LoadableCache[T]) Sync() {
	defer c.wg.Done()

	for item := range c.setChannel {
		c.Set(context.Background(), item.key, item.value)
	}
}

// Get returns the obj stored in cache if it exists.
func (c *LoadableCache[T]) Get(ctx context.Context, key any) (T, error) {
	var err error

	obj, err := c.cache.Get(ctx, key)
	if err == nil {
		return obj, nil
	}

	// Unable to find in cache, try to load it from load function
	obj, err = c.loadFunc(ctx, key)
	if err != nil {
		return obj, err
	}

	// Then, put it back in cache
	c.setChannel <- &loadableKeyValue[T]{key, obj}

	return obj, err
}

// GetWithTTL retrieves the object from the cache with its time to live (TTL) or loads it using the load function if not found.
func (c *LoadableCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	var err error

	obj, ttl, err := c.cache.GetWithTTL(ctx, key)
	if err == nil {
		return obj, ttl, nil
	}

	// Unable to find in cache, try to load it from load function
	obj, err = c.loadFunc(ctx, key)
	if err != nil {
		return obj, 0, err
	}

	// Then, put it back in cache
	c.setChannel <- &loadableKeyValue[T]{key, obj}

	return obj, ttl, err
}

// Set sets a value in available caches.
func (c *LoadableCache[T]) Set(ctx context.Context, key any, obj T) error {
	return c.cache.Set(ctx, key, obj)
}

// SetWithTTL sets a value in the cache with a specified time to live (TTL).
func (c *LoadableCache[T]) SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error {
	return c.cache.SetWithTTL(ctx, key, obj, ttl)
}

// Del removes a value from cache.
func (c *LoadableCache[T]) Del(ctx context.Context, key any) error {
	return c.cache.Del(ctx, key)
}

// Clear resets all cache data.
func (c *LoadableCache[T]) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

// Wait waits for all operations to finish.
func (c *LoadableCache[T]) Wait(ctx context.Context) {
	c.cache.Wait(ctx)
}

// Close closes the setChannel and waits for all operations to finish.
func (c *LoadableCache[T]) Close() error {
	close(c.setChannel)
	c.wg.Wait()

	return nil
}
