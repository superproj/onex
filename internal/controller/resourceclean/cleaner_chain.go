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

type Chain struct {
	mu     sync.Mutex
	client client.Client
	ds     store.IStore
}

func (c *Chain) Name() string {
	return "chain"
}

func (c *Chain) Initialize(client client.Client, ds store.IStore) {
	c.client = client
	c.ds = ds
}

func (c *Chain) Delete(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.V(4).InfoS("Cleanup chains from chain table")
	_, chains, err := c.ds.Chains().List(ctx, nil)
	if err != nil {
		klog.ErrorS(err, "Failed to list chains")
		return err
	}

	klog.V(4).InfoS("Successfully got chains", "count", len(chains))
	for _, chain := range chains {
		ch := v1beta1.Chain{}
		key := client.ObjectKey{Namespace: chain.Namespace, Name: chain.Name}
		if err := c.client.Get(ctx, key, &ch); err != nil {
			if apierrors.IsNotFound(err) {
				if derr := c.ds.Chains().Delete(ctx, where.F("namespace", chain.Namespace, "name", chain.Name)); derr != nil {
					klog.V(1).InfoS("Failed to delete chain", "chain", klog.KRef(chain.Namespace, chain.Name), "err", derr)
					continue
				}
				klog.V(4).InfoS("Successfully delete chain", "chain", klog.KRef(chain.Namespace, chain.Name))
			}

			klog.ErrorS(err, "Failed to get chain", "chain", klog.KRef(key.Namespace, key.Name))
			return err
		}
	}

	return nil
}
