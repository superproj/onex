// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package cleaner

import (
	"context"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

type ModelCompare struct {
	mu     sync.Mutex
	client client.Client
	ds     store.IStore
}

func (c *ModelCompare) Name() string {
	return "modelcompare"
}

func (c *ModelCompare) Initialize(client client.Client, ds store.IStore) {
	c.client = client
	c.ds = ds
}

func (c *ModelCompare) Sync(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.V(4).InfoS("Cleanup modelcompares from modelcompare table")
	_, modelcompares, err := c.ds.ModelCompares().List(ctx, "")
	if err != nil {
		klog.ErrorS(err, "Failed to list modelcompares")
		return err
	}

	klog.V(4).InfoS("Successfully got modelcompares", "count", len(modelcompares))
	for _, modelcompare := range modelcompares {
		ms := v1beta1.ModelCompare{}
		key := client.ObjectKey{Namespace: modelcompare.Namespace, Name: modelcompare.Name}
		if err := c.client.Get(ctx, key, &ms); err != nil {
			if apierrors.IsNotFound(err) {
				filter := map[string]any{"namespace": modelcompare.Namespace, "name": modelcompare.Name}
				if derr := c.ds.ModelCompares().Delete(ctx, filter); derr != nil {
					klog.V(1).InfoS("Failed to delete modelcompare", "modelcompare", klog.KRef(modelcompare.Namespace, modelcompare.Name), "err", derr)
					continue
				}
				klog.V(4).InfoS("Successfully delete modelcompare", "modelcompare", klog.KRef(modelcompare.Namespace, modelcompare.Name))
			}

			klog.ErrorS(err, "Failed to get modelcompare", "modelcompare", klog.KRef(key.Namespace, key.Name))
			return err
		}
	}

	return nil
}
