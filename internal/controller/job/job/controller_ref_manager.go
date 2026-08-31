// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"

	coreutil "github.com/onexstack/onex/internal/pkg/util/core"
)

// ControllerRefManager claims (adopts) and releases (orphans) child objects on
// behalf of the owning Job, using the object's controller reference to track
// ownership. It is the onex port of the BaseControllerRefManager logic in
// k8s.io/kubernetes/pkg/controller.
type ControllerRefManager struct {
	client         client.Client
	controller     *v1beta1.Job
	controllerKind schema.GroupVersionKind
}

// NewControllerRefManager returns a ControllerRefManager for the given owner.
func NewControllerRefManager(c client.Client, controller *v1beta1.Job, kind schema.GroupVersionKind) *ControllerRefManager {
	return &ControllerRefManager{client: c, controller: controller, controllerKind: kind}
}

// ClaimObject decides whether a child should be processed by the Job, adopting
// or releasing the child's controller reference as necessary. It returns true
// when the child is matched (already owned, or newly adopted).
func (m *ControllerRefManager) ClaimObject(ctx context.Context, obj client.Object, match func(metav1.Object) bool) (bool, error) {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef != nil {
		if controllerRef.UID != m.controller.GetUID() {
			// Owned by a different controller; ignore.
			return false, nil
		}
		if match(obj) {
			return true, nil
		}
		// Owned by us but no longer matches; release (orphan) it.
		return false, m.releaseObject(ctx, obj)
	}

	// Not owned by anyone.
	if !match(obj) {
		return false, nil
	}
	return true, m.adoptObject(ctx, obj)
}

func (m *ControllerRefManager) adoptObject(ctx context.Context, obj client.Object) error {
	patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	ref := *metav1.NewControllerRef(m.controller, m.controllerKind)
	obj.SetOwnerReferences(coreutil.EnsureOwnerRef(obj.GetOwnerReferences(), ref))
	return m.client.Patch(ctx, obj, patch)
}

func (m *ControllerRefManager) releaseObject(ctx context.Context, obj client.Object) error {
	patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	ref := *metav1.NewControllerRef(m.controller, m.controllerKind)
	obj.SetOwnerReferences(coreutil.RemoveOwnerRef(obj.GetOwnerReferences(), ref))
	return m.client.Patch(ctx, obj, patch)
}
