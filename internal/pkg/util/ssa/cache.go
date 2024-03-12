// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package ssa

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/superproj/onex/internal/pkg/util/hash"
)

const (
	// ttl is the duration for which we keep the keys in the cache.
	ttl = 10 * time.Minute

	// expirationInterval is the interval in which we will remove expired keys
	// from the cache.
	expirationInterval = 10 * time.Hour
)

// Cache caches SSA request results.
// Specifically we only use it to cache that a certain request
// doesn't have to be repeated anymore because there was no diff.
type Cache interface {
	// Add adds the given key to the Cache.
	// Note: keys expire after the ttl.
	Add(key string)

	// Has checks if the given key (still) exists in the Cache.
	// Note: keys expire after the ttl.
	Has(key string) bool
}

// NewCache creates a new cache.
func NewCache() Cache {
	r := &ssaCache{
		Store: cache.NewTTLStore(func(obj any) (string, error) {
			// We only add strings to the cache, so it's safe to cast to string.
			return obj.(string), nil
		}, ttl),
	}
	go func() {
		for {
			// Call list to clear the cache of expired items.
			// We have to do this periodically as the cache itself only expires
			// items lazily. If we don't do this the cache grows indefinitely.
			r.List()

			time.Sleep(expirationInterval)
		}
	}()
	return r
}

type ssaCache struct {
	cache.Store
}

// Add adds the given key to the Cache.
// Note: keys expire after the ttl.
func (r *ssaCache) Add(key string) {
	// Note: We can ignore the error here because by only allowing strings
	// and providing the corresponding keyFunc ourselves we can guarantee that
	// the error never occurs.
	_ = r.Store.Add(key)
}

// Has checks if the given key (still) exists in the Cache.
// Note: keys expire after the ttl.
func (r *ssaCache) Has(key string) bool {
	// Note: We can ignore the error here because GetByKey never returns an error.
	_, exists, _ := r.Store.GetByKey(key)
	return exists
}

// ComputeRequestIdentifier computes a request identifier for the cache.
// The identifier is unique for a specific request to ensure we don't have to re-run the request
// once we found out that it would not produce a diff.
// The identifier consists of: gvk, namespace, name and resourceVersion of the original object and a hash of the modified
// object. This ensures that we re-run the request as soon as either original or modified changes.
func ComputeRequestIdentifier(scheme *runtime.Scheme, original, modified client.Object) (string, error) {
	modifiedObjectHash, err := hash.Compute(modified)
	if err != nil {
		return "", errors.Wrapf(err, "failed to calculate request identifier: failed to compute hash for modified object")
	}

	gvk, err := apiutil.GVKForObject(original, scheme)
	if err != nil {
		return "", errors.Wrapf(err, "failed to calculate request identifier: failed to get GroupVersionKind of original object %s", klog.KObj(original))
	}

	return fmt.Sprintf("%s.%s.%s.%d", gvk.String(), klog.KObj(original), original.GetResourceVersion(), modifiedObjectHash), nil
}
