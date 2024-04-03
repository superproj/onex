// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package cache is a complete Go cache library that brings you
// multiple ways of managing your caches.
package cache

//go:generate mockgen -destination mock_cache.go -package cache github.com/superproj/onex/pkg/cache KeyGetter

import (
	"context"
	"crypto"
	"fmt"
	"reflect"
	"time"
)

// CacheType represents the type of cache, such as "noop", "l2", "chain", or "loadable".
type CacheType string

const (
	NoopCacheType     CacheType = "noop"
	L2CacheType       CacheType = "l2"
	ChainCacheType    CacheType = "chain"
	LoadableCacheType CacheType = "loadable"
)

func (ct CacheType) String() string {
	return string(ct)
}

// KeyFunc knows how to make a key from an object. Implementations should be deterministic.
type KeyFunc func(obj any) (string, error)

// Cache represents the interface for all caches (aggregates, metric, memory, redis, ...)
type Cache[T any] interface {
	// Set stores the object with the given key in the cache.
	Set(ctx context.Context, key any, obj T) error
	// Get retrieves the object from the cache based on the given key.
	Get(ctx context.Context, key any) (T, error)
	// SetWithTTL stores the object with the given key and time-to-live (TTL) in the cache.
	SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error
	// GetWithTTL retrieves the object and its time-to-live (TTL) from the cache based on the given key.
	GetWithTTL(ctx context.Context, key any) (T, time.Duration, error)
	// Del deletes the object from the cache based on the given key.
	Del(ctx context.Context, key any) error
	// Clear clears the cache.
	Clear(ctx context.Context) error
	// Wait waits for any pending operations to complete.
	Wait(ctx context.Context)
}

// KeyGetter is an interface for objects that can provide a cache key.
type KeyGetter interface {
	CacheKey() string
}

// keyFunc returns the cache key for the given key object by returning
// the key if type is string or by computing a checksum of key structure
// if its type is other than string.
func keyFunc(key any) string {
	switch typed := key.(type) {
	case string:
		return typed
	case KeyGetter:
		return typed.CacheKey()
	default:
		// hashes a given object into a string
		digester := crypto.MD5.New()
		fmt.Fprint(digester, reflect.TypeOf(typed))
		fmt.Fprint(digester, typed)
		hash := digester.Sum(nil)
		return fmt.Sprintf("%x", hash)
	}
}
