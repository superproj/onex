// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package log provides log utils.
package log

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// AddOwners adds the owners of an Object based on OwnerReferences as k/v pairs to the logger in ctx.
func AddOwners(ctx context.Context, obj metav1.Object) (context.Context, logr.Logger) {
	log := ctrl.LoggerFrom(ctx)

	owners := []owner{}
	for _, ownerRef := range obj.GetOwnerReferences() {
		owners = append(owners, owner{
			Kind:      ownerRef.Kind,
			Namespace: obj.GetNamespace(),
			Name:      ownerRef.Name,
		})
	}

	// Add owners as k/v pairs.
	keysAndValues := []any{}
	addedKinds := sets.Set[string]{}
	for _, owner := range owners {
		// Don't add duplicate kinds.
		if addedKinds.Has(owner.Kind) {
			continue
		}

		keysAndValues = append(keysAndValues, owner.Kind, klog.KRef(owner.Namespace, owner.Name))
		addedKinds.Insert(owner.Kind)
	}

	log = log.WithValues(keysAndValues...)
	ctx = ctrl.LoggerInto(ctx, log)
	return ctx, log
}

// AddWithMinerSetOwners adds the owners of an Object based on OwnerReferences as k/v pairs to the logger in ctx.
// Note: If an owner is a MinerSet we also add the owners from the MinerSet OwnerReferences.
func AddWithMinerSetOwners(ctx context.Context, c client.Client, obj metav1.Object) (context.Context, logr.Logger, error) {
	log := ctrl.LoggerFrom(ctx)

	owners, err := getOwners(ctx, c, obj)
	if err != nil {
		return nil, logr.Logger{}, errors.Wrapf(err, "failed to add object hierarchy to logger")
	}

	// Add owners as k/v pairs.
	keysAndValues := []any{}
	addedKinds := sets.Set[string]{}
	for _, owner := range owners {
		// Don't add duplicate kinds.
		if addedKinds.Has(owner.Kind) {
			continue
		}

		keysAndValues = append(keysAndValues, owner.Kind, klog.KRef(owner.Namespace, owner.Name))
		addedKinds.Insert(owner.Kind)
	}
	log = log.WithValues(keysAndValues...)

	ctx = ctrl.LoggerInto(ctx, log)
	return ctx, log, nil
}

// owner represents an owner of an object.
type owner struct {
	Kind      string
	Name      string
	Namespace string
}

// getOwners returns owners of an Object based on OwnerReferences.
// Note: If an owner is a MinerSet we also return the owners from the MinerSet OwnerReferences.
func getOwners(ctx context.Context, c client.Client, obj metav1.Object) ([]owner, error) {
	owners := []owner{}
	for _, ownerRef := range obj.GetOwnerReferences() {
		owners = append(owners, owner{
			Kind:      ownerRef.Kind,
			Namespace: obj.GetNamespace(),
			Name:      ownerRef.Name,
		})

		// continue if the ownerRef does not point to a MinerSet.
		if ownerRef.Kind != "MinerSet" {
			continue
		}

		// get owners of the MinerSet.
		var ms v1beta1.MinerSet
		if err := c.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: ownerRef.Name}, &ms); err != nil {
			// continue if the MinerSet doesn't exist.
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, errors.Wrapf(err, "failed to get owners: failed to get MinerSet %s", klog.KRef(obj.GetNamespace(), ownerRef.Name))
		}

		for _, ref := range ms.GetOwnerReferences() {
			owners = append(owners, owner{
				Kind:      ref.Kind,
				Namespace: obj.GetNamespace(),
				Name:      ref.Name,
			})
		}
	}

	return owners, nil
}
