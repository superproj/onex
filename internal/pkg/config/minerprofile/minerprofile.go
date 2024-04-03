// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerprofile

import (
	"context"
	"encoding/json"
	"sync"
	"unsafe"

	"github.com/dgraph-io/ristretto"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	"github.com/superproj/onex/internal/pkg/config"
	"github.com/superproj/onex/pkg/cache"
	ristrettostore "github.com/superproj/onex/pkg/cache/store/ristretto"
)

// MinerProfile represents the profile of a miner.
type MinerProfile struct {
	CPU              resource.Quantity `json:"cpu,omitempty"`
	Memory           resource.Quantity `json:"memory,omitempty"`
	MiningDifficulty int               `json:"miningDifficulty,omitempty"`
}

// cacher is a struct for caching MinerProfile data.
type cacher struct {
	mu     sync.Mutex
	client any
	data   *cache.DelegateCache[*MinerProfile]
}

var (
	g    = new(cacher)
	once = sync.Once{}
)

// defaultCostFn is a function to calculate the cost of a cache item.
var defaultCostFn = func(value any) int64 {
	size := int64(unsafe.Sizeof(MinerProfile{}))
	//nolint: staticcheck
	if mp, ok := value.(MinerProfile); ok {
		size += int64(unsafe.Sizeof(mp.CPU))
		size += int64(unsafe.Sizeof(mp.Memory))
		size += int64(unsafe.Sizeof(mp.MiningDifficulty))
	}

	return size
}

// loadMinerProfileData loads MinerProfile data into the cache.
func (c *cacher) loadMinerProfileData() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Fetch MinerProfile data from the configuration.
	cm, err := config.MinerTypesName.GetConfig(c.client)
	if err != nil {
		return err
	}

	// Configure the ristretto cache.
	cfg := &ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		// Metrics:     config.MetricsDurationSec > 0,
		Cost: defaultCostFn,
		OnEvict: func(item *ristretto.Item) {
			klog.V(4).InfoS("Cache evict", "item", item, "cost", item.Cost, "expiration", item.Expiration)
		},
		OnReject: func(item *ristretto.Item) {
			klog.V(4).InfoS("Cache reject", "item", item, "cost", item.Cost, "expiration", item.Expiration)
		},
	}
	riscache, err := ristretto.NewCache(cfg)
	if err != nil {
		return err
	}
	risstore := ristrettostore.NewRistretto(riscache)
	cache := cache.New[*MinerProfile](risstore)

	// Populate the cache with MinerProfile data.
	for k, v := range cm.Data {
		var profile MinerProfile
		if err := json.Unmarshal([]byte(v), &profile); err != nil {
			return err
		}

		cache.Set(context.TODO(), k, &profile)
	}

	c.data = cache
	return nil
}

// Init initializes the cache with MinerProfile data.
func Init(ctx context.Context, client any) (err error) {
	once.Do(func() {
		tc := cacher{client: client}

		if err = tc.loadMinerProfileData(); err != nil {
			return
		}

		g = &tc
	})
	return
}

// GetMinerProfile retrieves a MinerProfile from the cache.
func GetMinerProfile(key string) (*MinerProfile, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if obj, err := g.data.Get(context.TODO(), key); err == nil {
		return obj, true
	}

	// If the item is not found in the cache, reload data and try again.
	if err := g.loadMinerProfileData(); err != nil {
		return nil, false
	}

	if obj, err := g.data.Get(context.TODO(), key); err == nil {
		return obj, true
	}

	return nil, false
}
