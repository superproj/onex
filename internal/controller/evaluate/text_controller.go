// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:dupl
package evaluate

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/superproj/onex/internal/pkg/util/annotations"
	cmerrors "github.com/superproj/onex/pkg/errors"
	"github.com/superproj/onex/pkg/record"
	//"github.com/superproj/onex/internal/pkg/util/conditions"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	"github.com/superproj/onex/internal/pkg/util/patch"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const eventControllerName = "controller-manager.evaluate"

// TextReconciler sync a Event object to database.
type TextReconciler struct {
	client client.Client
}

// SetupWithManager sets up the controller with the Manager.
func (r *TextReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Evaluate{}).
		WithOptions(options).
		Named(eventControllerName)

	r.client = mgr.GetClient()

	return builder.Complete(r)
}

func (r *TextReconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the Evaluate instance
	eva := &v1beta1.Evaluate{}
	if err := r.client.Get(ctx, rq.NamespacedName, eva); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if eva.Spec.Provider != "text" {
		return ctrl.Result{}, nil
	}

	// AddOwners adds the owners of Chain as k/v pairs to the logger.
	log.V(4).Info("Reconcile evaluate")

	// Return early if the object is paused.
	if annotations.IsPaused(eva) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	helper, err := patch.NewHelper(eva, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the Miner object and status after each reconciliation.
		r.reconcilePhase(ctx, eva)

		// Always attempt to patch the object and status after each reconciliation.
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := helper.Patch(ctx, eva, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(eva, v1beta1.EvaluateFinalizer) {
		controllerutil.AddFinalizer(eva, v1beta1.EvaluateFinalizer)
		return ctrl.Result{}, nil
	}

	if !eva.GetDeletionTimestamp().IsZero() {
		// 这里执行删除逻辑
		controllerutil.RemoveFinalizer(eva, v1beta1.EvaluateFinalizer)
		return r.reconcileDelete(ctx, eva)
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, eva)
}

func (r *TextReconciler) reconcile(ctx context.Context, eva *v1beta1.Evaluate) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	if eva.Status.Phase == string(v1beta1.EvaluatePhaseFailed) {
		log.V(1).Info("Evaluate has gone `Failed` phase. It won't reconcile")
		return ctrl.Result{}, nil
	}

	/*
		if eva.Status.Phase == string(v1beta1.EvaluatePhaseSucceeded) {
			log.V(1).Info("Evaluate has gone `Succeeded` phase. It won't reconcile")
			return ctrl.Result{}, nil
		}
	*/

	// 模拟训练任务
	duration := time.Now().Sub(eva.GetCreationTimestamp().Time)
	if duration.Seconds() > 10 && duration.Seconds() < 30 {
		eva.Status.Phase = string(v1beta1.EvaluatePhaseEvaluating)
		return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
	}

	if duration.Seconds() > 30 && eva.Spec.ModelID != 1004 {
		eva.Status.Phase = string(v1beta1.EvaluatePhaseSucceeded)
		eva.Status.ArthurID = ptr.To("4001")
		eva.Status.Addresses = v1beta1.EvaluateAddresses{
			HDFSRoot:          "hdfs://testroot",
			HDFSPtPath:        "hdfs://test-pt-path",
			TOSTrainDataPath:  "cos://test-train-data-path",
			TOSTestDataPath:   "cos://test-test-data-path",
			TOSTrainDataCount: 31,
			TOSTestDataConut:  16,
		}
		record.Eventf(eva, "SuccessfulCreate", "Created evaluate %q", eva.Name)
		return ctrl.Result{}, nil
	}
	if duration.Seconds() > 30 && eva.Spec.ModelID == 1004 {
		eva.Status.Phase = string(v1beta1.EvaluatePhaseFailed)
		eva.Status.FailureReason = ptr.To(cmerrors.InsufficientResourcesMinerError)
		eva.Status.FailureMessage = ptr.To("Cannot found a useable GPU resource")
		record.Warnf(eva, "FailedCreate", "Failed to create evaluate %q", eva.Name)
		conditions.MarkFalse(eva, v1beta1.InfrastructureReadyCondition, v1beta1.PodStartupTimeoutReason, v1beta1.ConditionSeverityError, "")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
}

func (r *TextReconciler) reconcileDelete(ctx context.Context, eva *v1beta1.Evaluate) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *TextReconciler) reconcilePhase(_ context.Context, eva *v1beta1.Evaluate) {
	if eva.Status.Phase == string(v1beta1.EvaluatePhaseSucceeded) || eva.Status.Phase == string(v1beta1.EvaluatePhaseFailed) {
		eva.Status.EndedAt = ptr.To(metav1.Now())
	}
}
