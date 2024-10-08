// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package resourceclean

import (
	"context"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/store/where"
)

type Miner struct {
	mu     sync.Mutex
	client client.Client
	ds     store.IStore
}

func (c *Miner) Name() string {
	return "miner"
}

func (c *Miner) Initialize(client client.Client, ds store.IStore) {
	c.client = client
	c.ds = ds
}

func (c *Miner) Delete(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.V(4).InfoS("Cleanup miners from miner table")
	_, miners, err := c.ds.Miners().List(ctx, nil)
	if err != nil {
		klog.ErrorS(err, "Failed to list miners")
		return err
	}

	klog.V(4).InfoS("Successfully got miners", "count", len(miners))
	for _, miner := range miners {
		m := v1beta1.Miner{}
		key := client.ObjectKey{Namespace: miner.Namespace, Name: miner.Name}
		if err := c.client.Get(ctx, key, &m); err != nil {
			if apierrors.IsNotFound(err) {
				if derr := c.ds.Miners().Delete(ctx, where.F("namespace", miner.Namespace, "name", miner.Name)); derr != nil {
					klog.V(1).InfoS("Failed to delete miner", "miner", klog.KRef(miner.Namespace, miner.Name), "err", derr)
					continue
				}
				klog.V(4).InfoS("Successfully delete miner", "miner", klog.KRef(miner.Namespace, miner.Name))
			}

			klog.ErrorS(err, "Failed to get miner", "miner", klog.KRef(key.Namespace, key.Name))
			return err
		}
	}

	return nil
}
