// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockCache[T any] struct {
	storage map[any]T
}

func (m *mockCache[T]) Set(ctx context.Context, key any, obj T) error {
	m.storage[key] = obj
	return nil
}

func (m *mockCache[T]) Get(ctx context.Context, key any) (T, error) {
	obj, exists := m.storage[key]
	if !exists {
		return *new(T), fmt.Errorf("key not found")
	}
	return obj, nil
}

func (m *mockCache[T]) SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error {
	m.storage[key] = obj
	// Simulate TTL behavior
	go func() {
		time.Sleep(ttl)
		delete(m.storage, key)
	}()
	return nil
}

func (m *mockCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	obj, exists := m.storage[key]
	if !exists {
		return *new(T), 0, fmt.Errorf("key not found")
	}
	// Simulate TTL behavior
	return obj, 10 * time.Minute, nil
}

func (m *mockCache[T]) Del(ctx context.Context, key any) error {
	delete(m.storage, key)
	return nil
}

func (m *mockCache[T]) Clear(ctx context.Context) error {
	m.storage = make(map[any]T)
	return nil
}

func (m *mockCache[T]) Wait(ctx context.Context) {
	// No-op for the mock implementation
}

func TestCacheSetAndGet(t *testing.T) {
	// Create a mock cache
	cache := &mockCache[any]{storage: make(map[any]any)}

	// Test Set and Get operations
	key := "testKey"
	value := "testValue"
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Errorf("Error setting value in cache: %v", err)
	}

	retrievedValue, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Errorf("Error getting value from cache: %v", err)
	}
	if retrievedValue != value {
		t.Errorf("Retrieved value does not match the expected value")
	}
}

func TestCacheSetWithTTLAndGetWithTTL(t *testing.T) {
	// Create a mock cache
	cache := &mockCache[any]{storage: make(map[any]any)}

	// Test SetWithTTL and GetWithTTL operations
	key := "testKey"
	value := "testValue"
	ttl := 5 * time.Second
	err := cache.SetWithTTL(context.Background(), key, value, ttl)
	if err != nil {
		t.Errorf("Error setting value with TTL in cache: %v", err)
	}

	retrievedValue, retrievedTTL, err := cache.GetWithTTL(context.Background(), key)
	if err != nil {
		t.Errorf("Error getting value with TTL from cache: %v", err)
	}
	if retrievedValue != value {
		t.Errorf("Retrieved value does not match the expected value")
	}
	if retrievedTTL < ttl {
		t.Errorf("Retrieved TTL is less than the expected TTL")
	}
}

func TestCacheDel(t *testing.T) {
	// Create a mock cache
	cache := &mockCache[any]{storage: make(map[any]any)}

	// Test Del operation
	key := "testKey"
	value := "testValue"
	cache.Set(context.Background(), key, value)

	err := cache.Del(context.Background(), key)
	if err != nil {
		t.Errorf("Error deleting value from cache: %v", err)
	}

	_, err = cache.Get(context.Background(), key)
	if err == nil {
		t.Errorf("Retrieved value after deletion, expected key to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	// Create a mock cache
	cache := &mockCache[any]{storage: make(map[any]any)}

	// Test Clear operation
	key := "testKey"
	value := "testValue"
	cache.Set(context.Background(), key, value)

	err := cache.Clear(context.Background())
	if err != nil {
		t.Errorf("Error clearing the cache: %v", err)
	}

	_, err = cache.Get(context.Background(), key)
	if err == nil {
		t.Errorf("Retrieved value after clearing, expected cache to be empty")
	}
}
