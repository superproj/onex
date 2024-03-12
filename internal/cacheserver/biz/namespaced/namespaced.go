// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package namespaced

//go:generate mockgen -destination mock_namespaced.go -package namespaced github.com/superproj/onex/internal/cacheserver/biz/namespaced NamespacedBiz

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	durationpb "google.golang.org/protobuf/types/known/durationpb"

	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
	"github.com/superproj/onex/pkg/cache"
)

// NamespacedBiz defines the methods that need to be implemented for namespaced cache operations.
type NamespacedBiz interface {
	Set(ctx context.Context, key string, value *any.Any, ttl *durationpb.Duration) error
	Get(ctx context.Context, key string) (*v1.GetResponse, error)
	Del(ctx context.Context, key string) error
}

// NamespacedKey represents a key with a namespace.
type NamespacedKey struct {
	Namespace string
	Key       string
}

// namespacedBiz is a concrete implementation of NamespacedBiz.
type namespacedBiz struct {
	cache     cache.Cache[*any.Any]
	namespace string
}

// Ensure that namespacedBiz implements the NamespacedBiz interface.
var _ NamespacedBiz = (*namespacedBiz)(nil)

// CacheKey returns the cache key for the NamespacedKey.
func (k NamespacedKey) CacheKey() string {
	return fmt.Sprintf("namespace:%s:%s", k.Namespace, k.Key)
}

// New creates a new namespacedBiz instance with the specified cache and namespace.
func New(cache cache.Cache[*any.Any], namespace string) *namespacedBiz {
	return &namespacedBiz{cache: cache, namespace: namespace}
}

// Set stores a value with the given key and time to live (TTL) in the namespaced cache.
func (b *namespacedBiz) Set(ctx context.Context, key string, value *any.Any, ttl *durationpb.Duration) error {
	return b.cache.SetWithTTL(ctx, NamespacedKey{b.namespace, key}, value, ttl.AsDuration())
}

// Get retrieves a value from the namespaced cache by its key.
func (b *namespacedBiz) Get(ctx context.Context, key string) (*v1.GetResponse, error) {
	value, ttl, err := b.cache.GetWithTTL(ctx, NamespacedKey{b.namespace, key})
	if err != nil {
		return nil, err
	}

	return &v1.GetResponse{Value: value, Expire: durationpb.New(ttl)}, nil
}

// Del deletes a value from the namespaced cache by its key.
func (b *namespacedBiz) Del(ctx context.Context, key string) error {
	return b.cache.Del(ctx, NamespacedKey{b.namespace, key})
}
