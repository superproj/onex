// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package predicates implements predicate utilities.
package predicates

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// MinerSetCreateNotPaused returns a predicate that returns true for a create event when a minerset has Spec.Paused set as false
// it also returns true if the resource provided is not a MinerSet to allow for use with controller-runtime NewControllerManagedBy.
func MinerSetCreateNotPaused(logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			log := logger.WithValues("predicate", "MinerSetCreateNotPaused", "eventType", "create")

			c, ok := e.Object.(*v1beta1.MinerSet)
			if !ok {
				log.V(4).Info("Expected MinerSet", "type", fmt.Sprintf("%T", e.Object))
				return false
			}
			log = log.WithValues("MinerSet", klog.KObj(c))

			// Only need to trigger a reconcile if the MinerSet.Spec.Paused is false
			if !isPaused(c) {
				log.V(6).Info("MinerSet is not paused, allowing further processing")
				return true
			}

			log.V(4).Info("MinerSet is paused, blocking further processing")
			return false
		},
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

// MinerSetUpdateUnpaused returns a predicate that returns true for an update event when a minerset has Spec.Paused changed from true to false
// it also returns true if the resource provided is not a MinerSet to allow for use with controller-runtime NewControllerManagedBy.
func MinerSetUpdateUnpaused(logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			log := logger.WithValues("predicate", "MinerSetUpdateUnpaused", "eventType", "update")

			oldMinerSet, ok := e.ObjectOld.(*v1beta1.MinerSet)
			if !ok {
				log.V(4).Info("Expected MinerSet", "type", fmt.Sprintf("%T", e.ObjectOld))
				return false
			}
			log = log.WithValues("MinerSet", klog.KObj(oldMinerSet))

			newMinerSet := e.ObjectNew.(*v1beta1.MinerSet)

			if isPaused(oldMinerSet) && !isPaused(newMinerSet) {
				log.V(4).Info("MinerSet was unpaused, allowing further processing")
				return true
			}

			// This predicate always work in "or" with Paused predicates
			// so the logs are adjusted to not provide false negatives/verbosity al V<=5.
			log.V(6).Info("MinerSet was not unpaused, blocking further processing")
			return false
		},
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

// MinerSetUnpaused returns a Predicate that returns true on MinerSet creation events where MinerSet.Spec.Paused is false
// and Update events when MinerSet.Spec.Paused transitions to false.
// This implements a common requirement for many minerset-api and provider controllers (such as MinerSet Infrastructure
// controllers) to resume reconciliation when the MinerSet is unpaused.
// Example use:
//
//	err := controller.Watch(
//	    &source.Kind{Type: &v1beta1.MinerSet{}},
//	    &handler.EnqueueRequestsFromMapFunc{
//	        ToRequests: minersetToMiners,
//	    },
//	    predicates.MinerSetUnpaused(r.Log),
//	)
func MinerSetUnpaused(logger logr.Logger) predicate.Funcs {
	log := logger.WithValues("predicate", "MinerSetUnpaused")

	// Use any to ensure we process either create or update events we care about
	return Any(log, MinerSetCreateNotPaused(log), MinerSetUpdateUnpaused(log))
}

func isPaused(ms *v1beta1.MinerSet) bool {
	_, ok := ms.GetAnnotations()[v1beta1.PausedAnnotation]
	return ok
}
