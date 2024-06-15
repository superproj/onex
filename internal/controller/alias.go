// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package controller

import (
	"context"

	"k8s.io/client-go/metadata"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	evaluatecontroller "github.com/superproj/onex/internal/controller/evaluate"
	modelcomparecontroller "github.com/superproj/onex/internal/controller/modelcompare"
	namespacecontroller "github.com/superproj/onex/internal/controller/namespace"
	synccontroller "github.com/superproj/onex/internal/controller/sync"
	"github.com/superproj/onex/internal/gateway/store"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

// Following types provides access to reconcilers implemented in internal/controller, thus
// allowing users to provide a single binary "batteries included" with OneX and providers of choice.

type ModelCompareReconciler struct{}

func (r *ModelCompareReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return (&modelcomparecontroller.Reconciler{
		//WatchFilterValue: r.WatchFilterValue,
	}).SetupWithManager(ctx, mgr, options)
}

// SyncReconciler sync onex resource to database.
type SyncReconciler struct {
	Store store.IStore
}

func (r *SyncReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	// setup chainSync controller
	if err := (&synccontroller.ModelCompareSyncReconciler{
		Store: r.Store,
	}).SetupWithManager(ctx, mgr, options); err != nil {
		return err
	}

	/*
		if err := (&synccontroller.EvaluateSyncReconciler{
			Store: r.Store,
		}).SetupWithManager(ctx, mgr, options); err != nil {
			return err
		}
	*/

	return nil
}

// EvaluateReconciler evaluate model traing.
type EvaluateReconciler struct{}

func (r *EvaluateReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	// setup chainSync controller
	if err := (&evaluatecontroller.TextReconciler{}).SetupWithManager(ctx, mgr, options); err != nil {
		return err
	}

	/*
		// For future use.
		// setup imageReconcile controller
		if err := (&evaluatecontroller.ImageReconciler{ }).SetupWithManager(ctx, mgr, options); err != nil {
			return err
		}
	*/

	return nil
}

// NamespacedResourcesDeleterReconciler is a reconciler used to delete a namespace with all resources in it.
type NamespacedResourcesDeleterReconciler struct {
	Client         clientset.Interface
	MetadataClient metadata.Interface
}

func (r *NamespacedResourcesDeleterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return namespacecontroller.NewNamespacedResourcesDeleter(mgr, r.Client, r.MetadataClient).SetupWithManager(mgr, options)
}
