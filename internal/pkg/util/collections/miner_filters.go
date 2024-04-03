// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package collections

import (
	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/internal/pkg/util/conditions"
	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Func is the functon definition for a filter.
type Func func(miner *v1beta1.Miner) bool

// And returns a filter that returns true if all of the given filters returns true.
func And(filters ...Func) Func {
	return func(miner *v1beta1.Miner) bool {
		for _, f := range filters {
			if !f(miner) {
				return false
			}
		}
		return true
	}
}

// Or returns a filter that returns true if any of the given filters returns true.
func Or(filters ...Func) Func {
	return func(miner *v1beta1.Miner) bool {
		for _, f := range filters {
			if f(miner) {
				return true
			}
		}
		return false
	}
}

// Not returns a filter that returns the opposite of the given filter.
func Not(mf Func) Func {
	return func(miner *v1beta1.Miner) bool {
		return !mf(miner)
	}
}

// HasControllerRef is a filter that returns true if the miner has a controller ref.
func HasControllerRef(miner *v1beta1.Miner) bool {
	if miner == nil {
		return false
	}
	return metav1.GetControllerOf(miner) != nil
}

// OwnedMiners returns a filter to find all miners owned by specified owner.
// Usage: GetFilteredMinersForCluster(ctx, client, cluster, OwnedMiners(controlPlane)).
func OwnedMiners(owner client.Object) func(miner *v1beta1.Miner) bool {
	return func(miner *v1beta1.Miner) bool {
		if miner == nil {
			return false
		}
		return coreutil.IsOwnedByObject(miner, owner)
	}
}

// ActiveMiners returns a filter to find all active miners.
// Usage: GetFilteredMinersForCluster(ctx, client, cluster, ActiveMiners).
func ActiveMiners(miner *v1beta1.Miner) bool {
	if miner == nil {
		return false
	}
	return miner.DeletionTimestamp.IsZero()
}

// HasDeletionTimestamp returns a filter to find all miners that have a deletion timestamp.
func HasDeletionTimestamp(miner *v1beta1.Miner) bool {
	if miner == nil {
		return false
	}
	return !miner.DeletionTimestamp.IsZero()
}

// HasUnhealthyCondition returns a filter to find all miners that have a MinerHealthCheckSucceeded condition set to False,
// indicating a problem was detected on the miner, and the MinerOwnerRemediated condition set, indicating that KCP is
// responsible of performing remediation as owner of the miner.
func HasUnhealthyCondition(miner *v1beta1.Miner) bool {
	if miner == nil {
		return false
	}
	return conditions.IsFalse(miner, v1beta1.MinerHealthCheckSucceededCondition) && conditions.IsFalse(miner, v1beta1.MinerOwnerRemediatedCondition)
}

// IsReady returns a filter to find all miners with the ReadyCondition equals to True.
func IsReady() Func {
	return func(miner *v1beta1.Miner) bool {
		if miner == nil {
			return false
		}
		return conditions.IsTrue(miner, v1beta1.ReadyCondition)
	}
}

// ShouldRolloutAfter returns a filter to find all miners where
// CreationTimestamp < rolloutAfter < reconciliationTIme.
func ShouldRolloutAfter(reconciliationTime, rolloutAfter *metav1.Time) Func {
	return func(miner *v1beta1.Miner) bool {
		if miner == nil {
			return false
		}
		return miner.CreationTimestamp.Before(rolloutAfter) && rolloutAfter.Before(reconciliationTime)
	}
}

// HasAnnotationKey returns a filter to find all miners that have the
// specified Annotation key present.
func HasAnnotationKey(key string) Func {
	return func(miner *v1beta1.Miner) bool {
		if miner == nil || miner.Annotations == nil {
			return false
		}
		if _, ok := miner.Annotations[key]; ok {
			return true
		}
		return false
	}
}

// WithVersion returns a filter to find miner that have a non empty and valid version.
func WithVersion() Func {
	return func(miner *v1beta1.Miner) bool {
		if miner == nil {
			return false
		}
		if miner.Spec.MinerType == "" {
			return false
		}
		if _, err := semver.ParseTolerant(miner.Spec.MinerType); err != nil {
			return false
		}
		return true
	}
}
