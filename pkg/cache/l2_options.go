// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cache

import (
	"github.com/dgraph-io/ristretto"
)

// L2Option represents a cache option function.
type L2Option func(o *L2Options)

// L2Options represents the options for L2 cache configuration.
type L2Options struct {
	// Disable local cache. To enable or disable the local cache,
	// you need to restart the service.
	Disable bool

	// NumCounters determines the number of counters (keys) to keep that hold
	// access frequency information. It's generally a good idea to have more
	// counters than the max cache capacity, as this will improve eviction
	// accuracy and subsequent hit ratios.
	//
	// For example, if you expect your cache to hold 1,000,000 items when full,
	// NumCounters should be 10,000,000 (10x). Each counter takes up roughly
	// 3 bytes (4 bits for each counter * 4 copies plus about a byte per
	// counter for the bloom filter). Note that the number of counters is
	// internally rounded up to the nearest power of 2, so the space usage
	// may be a little larger than 3 bytes * NumCounters.
	NumCounters int64
	// MaxCost can be considered as the cache capacity, in whatever units you
	// choose to use.
	//
	// For example, if you want the cache to have a max capacity of 100MB, you
	// would set MaxCost to 100,000,000 and pass an item's number of bytes as
	// the `cost` parameter for calls to Set. If new items are accepted, the
	// eviction process will take care of making room for the new item and not
	// overflowing the MaxCost value.
	MaxCost int64
	// BufferItems determines the size of Get buffers.
	//
	// Unless you have a rare use case, using `64` as the BufferItems value
	// results in good performance.
	BufferItems int64
	// Metrics determines whether cache statistics are kept during the cache's
	// lifetime. There *is* some overhead to keeping statistics, so you should
	// only set this flag to true when testing or throughput performance isn't a
	// major factor.
	Metrics bool
}

// L2WithNumCounters sets the number of counters for L2 cache.
func L2WithNumCounters(numCounters int64) L2Option {
	return func(opts *L2Options) {
		opts.NumCounters = numCounters
	}
}

// L2WithDisableCache enables or disables the local cache for L2.
func L2WithDisableCache(disable bool) L2Option {
	return func(opts *L2Options) {
		opts.Disable = disable
	}
}

// L2WithMetrics sets whether cache statistics are kept during the cache's lifetime for L2 cache.
func L2WithMetrics(enable bool) L2Option {
	return func(opts *L2Options) {
		opts.Metrics = enable
	}
}

// NewL2Options instantiates a new L2Options with default values.
func NewL2Options() *L2Options {
	return &L2Options{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		Metrics:     false,
	}
}

// ApplyTo applies the L2Options to a ristretto.Config.
func (o *L2Options) ApplyTo(cfg *ristretto.Config) {
	cfg.NumCounters = o.NumCounters
	cfg.MaxCost = o.MaxCost
	cfg.BufferItems = o.BufferItems
	cfg.Metrics = o.Metrics
}
