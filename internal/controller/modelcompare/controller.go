// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerset

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/superproj/onex/internal/pkg/util/annotations"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	labelsutil "github.com/superproj/onex/internal/pkg/util/labels"
	logutil "github.com/superproj/onex/internal/pkg/util/log"
	"github.com/superproj/onex/internal/pkg/util/patch"
	"github.com/superproj/onex/internal/pkg/util/predicates"
	"github.com/superproj/onex/internal/pkg/util/ssa"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/record"
	retryutil "github.com/superproj/onex/pkg/util/retry"
	stringsutil "github.com/superproj/onex/pkg/util/strings"
)

// MaxConcurrency used to prevent the high load of onex-apiserver caused by excessive concurrency,
// it is necessary to limit the miner create/delete concurrency.
const MaxConcurrency = 30

const controllerName = "modelcompare-controller"

var (
	// mcKind contains the schema.GroupVersionKind for the ModelCompare type.
	mcKind = v1beta1.SchemeGroupVersion.WithKind("ModelCompare")

	// stateConfirmationTimeout is the amount of time allowed to wait for desired state.
	stateConfirmationTimeout = 10 * time.Second

	// stateConfirmationInterval is the amount of time between polling for the desired state.
	// The polling is against a local memory cache.
	stateConfirmationInterval = 100 * time.Millisecond
)

// Reconciler reconciles a ModelCompare object.
type Reconciler struct {
	client    client.Client
	APIReader client.Reader

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
	ssaCache         ssa.Cache
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.ModelCompare{}).
		Owns(&v1beta1.Evaluate{}).
		Watches(
			&v1beta1.Evaluate{},
			handler.EnqueueRequestsFromMapFunc(r.EvaluateToModelCompares)).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))

	r.client = mgr.GetClient()
	r.ssaCache = ssa.NewCache()
	r.APIReader = mgr.GetAPIReader()

	return builder.Complete(r)
}

// Reconcile reads that state of the OneX for a ModelCompare object and makes changes based on the state read
// and what is in the ModelCompare.Spec.
func (r *Reconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	// 1. Fetch the ModelCompare object
	mc := &v1beta1.ModelCompare{}
	if err := r.client.Get(ctx, rq.NamespacedName, mc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ctx, log := logutil.AddOwners(ctx, mc)
	log.V(4).Info("Reconcile modelcompare")
	// Return early if the object is paused.
	if annotations.IsPaused(mc) {
		log.V(2).Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	helper, err := patch.NewHelper(mc, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchModelCompare(ctx, helper, mc); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Ignore deleted ModelCompares, this can happen when foregroundDeletion is enabled
	if !mc.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	result, err := r.reconcile(ctx, mc)
	if err != nil {
		log.Error(err, "Failed to reconcile ModelCompare")
		record.Warnf(mc, "ReconcileError", "%v", err)
	}
	return result, err
}

func patchModelCompare(ctx context.Context, helper *patch.Helper, mc *v1beta1.ModelCompare, options ...patch.Option) error {
	return helper.Patch(ctx, mc, options...)
}

func (r *Reconciler) reconcile(ctx context.Context, mc *v1beta1.ModelCompare) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Reconcile and retrieve the ModelCompare object.
	selectorMap, err := metav1.LabelSelectorAsMap(&mc.Spec.Selector)
	if err != nil {
		log.Error(err, "Failed to convert ModelCompare label selector to a map")
		return ctrl.Result{}, err
	}

	// Get all Evaluates linked to this ModelCompare.
	allEvaluates := &v1beta1.EvaluateList{}
	if err := r.client.List(ctx, allEvaluates, client.InNamespace(mc.Namespace), client.MatchingLabels(selectorMap)); err != nil {
		log.Error(err, "Failed to list evaluates")
		return ctrl.Result{}, err
	}

	// Filter out irrelevant miners (i.e. IsControlledBy something else) and claim orphaned miners.
	// Evaluates in deleted state are deliberately not excluded.
	filteredEvaluates := make([]*v1beta1.Evaluate, 0, len(allEvaluates.Items))
	for idx := range allEvaluates.Items {
		evaluate := &allEvaluates.Items[idx]
		if shouldExcludeEvaluate(mc, evaluate) {
			continue
		}

		// Attempt to adopt evaluate if it meets previous conditions and it has no controller references.
		if metav1.GetControllerOf(evaluate) == nil {
			if err := r.adoptOrphan(ctx, mc, evaluate); err != nil {
				log.Error(err, "Failed to adopt Evaluate", "evaluate", klog.KObj(evaluate))
				record.Warnf(mc, "FailedAdopt", "Failed to adopt Evaluate %q: %v", evaluate.Name, err)
				continue
			}
			log.V(2).Info("Adopted Evaluate", "evaluate", klog.KObj(evaluate))
			record.Eventf(mc, "SuccessfulAdopt", "Adopted Evaluate %q", evaluate.Name)
		}

		filteredEvaluates = append(filteredEvaluates, evaluate)
	}

	result := ctrl.Result{}

	if err := r.syncEvaluates(ctx, mc, filteredEvaluates); err != nil {
		return ctrl.Result{}, err
	}

	syncResult, syncErr := r.syncModelIDs(ctx, mc, filteredEvaluates)
	result = coreutil.LowestNonZeroResult(result, syncResult)

	// Always updates status as miners come up or die.
	if err := r.updateStatus(ctx, mc, filteredEvaluates); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update ModelCompare's Status, err: %w", kerrors.NewAggregate([]error{err, syncErr}))
	}

	if syncErr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync ModelCompare replicas, err: %w", syncErr)
	}

	return result, nil
}

func (r *Reconciler) syncModelIDs(ctx context.Context, mc *v1beta1.ModelCompare, evaluates []*v1beta1.Evaluate) (ctrl.Result, error) {
	if len(mc.Spec.ModelIDs) == 0 {
		return ctrl.Result{}, fmt.Errorf("the modelIDs field in Spec for modelcompare %v is nil, this should not be allowed", mc.Name)
	}

	existingEvaluates := make(map[int64]*v1beta1.Evaluate, 0)
	for _, evaluate := range evaluates {
		existingEvaluates[evaluate.Spec.ModelID] = evaluate
	}

	addedList := make([]*v1beta1.Evaluate, 0)
	for _, modelID := range mc.Spec.ModelIDs {
		// 如果存在则，不执行任何操作
		if _, ok := existingEvaluates[modelID]; ok {
			continue
		}

		// 如果不存在则创建
		eva, err := r.createEvaluate(ctx, mc, modelID)
		if err != nil {
			return ctrl.Result{}, err
		}
		addedList = append(addedList, eva)
	}
	if err := r.waitForEvaluateCreation(ctx, addedList); err != nil {
		return ctrl.Result{}, err
	}

	currentModel := make(map[int64]struct{}, 0)
	deletedList := make([]*v1beta1.Evaluate, 0)
	for _, modelID := range mc.Spec.ModelIDs {
		currentModel[modelID] = struct{}{}
	}
	for modelID, eva := range existingEvaluates {
		if _, ok := currentModel[modelID]; ok {
			continue
		}

		// 存在则删除
		if err := r.deleteEvaluate(ctx, mc, eva); err != nil {
			return ctrl.Result{}, err
		}
		deletedList = append(deletedList, eva)

	}

	if err := r.waitForEvaluateDeletion(ctx, deletedList); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) syncEvaluates(ctx context.Context, mc *v1beta1.ModelCompare, evaluates []*v1beta1.Evaluate) error {
	// Also equal to `log := klog.FromContext(ctx)`.
	// ctrl.LoggerFrom and klog.FromContext(ctx) use the same context key: github.com/go-logr/logr.contextKey
	log := ctrl.LoggerFrom(ctx)

	for i := range evaluates {
		eva := evaluates[i]
		// If the evaluate is already being deleted, we don't need to update it.
		if !eva.DeletionTimestamp.IsZero() {
			continue
		}

		// Cleanup managed fields of all Miners.
		// We do this so that Miners that were created/patched before the controller adopted Server-Side-Apply (SSA)
		// (< v1.4.0) can also work with SSA. Otherwise, fields would be co-owned by our "old" "manager" and
		// "capi-minerset" and then we would not be able to e.g. drop labels and annotations.
		if err := ssa.CleanUpManagedFieldsForSSAAdoption(ctx, r.client, eva, controllerName); err != nil {
			return fmt.Errorf("failed to update evaluate: failed to adjust the managedFields of the Evaluate %q, err: %w", eva.Name, err)
		}

		// Update Evaluate to propagate in-place mutable fields from the ModelCompare.
		updatedEvaluate := r.computeDesiredEvaluate(mc, eva.Spec.ModelID, eva)
		err := ssa.Patch(ctx, r.client, controllerName, updatedEvaluate, ssa.WithCachingProxy{Cache: r.ssaCache, Original: eva})
		if err != nil {
			log.Error(err, "failed to update Evaluate", "Evaluate", klog.KObj(updatedEvaluate))
			return fmt.Errorf("failed to update Evaluate %q, err: %w", klog.KObj(updatedEvaluate), err)
		}
		evaluates[i] = updatedEvaluate
	}

	return nil
}

func (r *Reconciler) computeDesiredEvaluate(mc *v1beta1.ModelCompare, modelID int64, existingEvaluate *v1beta1.Evaluate) *v1beta1.Evaluate {
	gv := v1beta1.SchemeGroupVersion
	desiredEvaluate := &v1beta1.Evaluate{
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("Evaluate").Kind,
			APIVersion: gv.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-", mc.Name)),
			Namespace: mc.Namespace,
			// Note: By setting the ownerRef on creation we signal to the Evaluate controller that this is not a stand-alone Evaluate.
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(mc, mcKind)},
			Labels:          mc.Spec.Template.Labels,
			Annotations:     mc.Spec.Template.Annotations,
			Finalizers:      []string{v1beta1.EvaluateFinalizer},
		},
		Spec: *mc.Spec.Template.Spec.DeepCopy(),
	}

	if existingEvaluate != nil {
		desiredEvaluate.SetName(existingEvaluate.Name)
		desiredEvaluate.SetUID(existingEvaluate.UID)
	}

	// Set ModelID
	desiredEvaluate.Spec.ModelID = modelID

	// Set Labels
	desiredEvaluate.Labels = evaluateLabelsFromModelCompare(mc)

	// Set Annotations
	desiredEvaluate.Annotations = evaluateAnnotationsFromModelCompare(mc)

	return desiredEvaluate
}

// evaluateLabelsFromModelCompare computes the labels the Miner created from this ModelCompare should have.
func evaluateLabelsFromModelCompare(mc *v1beta1.ModelCompare) map[string]string {
	evaluateLabels := map[string]string{}
	for k, v := range mc.Spec.Template.Labels {
		evaluateLabels[k] = v
	}
	evaluateLabels[v1beta1.ModelCompareNameLabel] = labelsutil.MustFormatValue(mc.Name)
	return evaluateLabels
}

// evaluateAnnotationsFromModelCompare computes the annotations the Miner created from this ModelCompare should have.
func evaluateAnnotationsFromModelCompare(mc *v1beta1.ModelCompare) map[string]string {
	annotations := map[string]string{}
	for k, v := range mc.Spec.Template.Annotations {
		annotations[k] = v
	}
	return annotations
}

// shouldExcludeMiner returns true if the miner should be filtered out, false otherwise.
func shouldExcludeMiner(ms *v1beta1.ModelCompare, miner *v1beta1.Miner) bool {
	if metav1.GetControllerOf(miner) != nil && !metav1.IsControlledBy(miner, ms) {
		return true
	}

	return false
}

// shouldExcludeEvaluate returns true if the miner should be filtered out, false otherwise.
func shouldExcludeEvaluate(mc *v1beta1.ModelCompare, evaluate *v1beta1.Evaluate) bool {
	if metav1.GetControllerOf(evaluate) != nil && !metav1.IsControlledBy(evaluate, mc) {
		return true
	}

	return false
}

// adoptOrphan sets the ModelCompare as a controller OwnerReference to the Evaluate.
func (r *Reconciler) adoptOrphan(ctx context.Context, mc *v1beta1.ModelCompare, evaluate *v1beta1.Evaluate) error {
	patch := client.MergeFrom(evaluate.DeepCopy())
	newRef := *metav1.NewControllerRef(mc, mcKind)
	evaluate.OwnerReferences = append(evaluate.OwnerReferences, newRef)
	return r.client.Patch(ctx, evaluate, patch)
}

func (r *Reconciler) waitForEvaluateCreation(ctx context.Context, evaList []*v1beta1.Evaluate) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(evaList); i++ {
		eva := evaList[i]

		pollErr := retryutil.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			key := client.ObjectKey{Namespace: eva.Namespace, Name: eva.Name}
			if err := r.client.Get(ctx, key, &v1beta1.Evaluate{}); err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}

			return true, nil
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for evaluate object to be created", "evaluate", klog.KObj(eva))
			return pollErr
		}
	}

	return nil
}

func (r *Reconciler) waitForEvaluateDeletion(ctx context.Context, evaList []*v1beta1.Evaluate) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(evaList); i++ {
		eva := evaList[i]
		pollErr := retryutil.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			e := &v1beta1.Evaluate{}
			key := client.ObjectKey{Namespace: eva.Namespace, Name: eva.Name}
			err := r.client.Get(ctx, key, e)
			if apierrors.IsNotFound(err) || !eva.DeletionTimestamp.IsZero() {
				return true, nil
			}
			return false, err
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for evaluate object to be deleted", "evaluate", klog.KObj(eva))
			return pollErr
		}
	}

	return nil
}

func (r *Reconciler) EvaluateToModelCompares(ctx context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	eva, ok := o.(*v1beta1.Evaluate)
	if !ok {
		panic(fmt.Sprintf("Expected a Miner but got a %T", o))
	}

	log := ctrl.LoggerFrom(ctx, "Evaluate", klog.KObj(eva)) // TODO: test here

	// Check if the controller reference is already set and
	// return an empty result when one is found.
	for _, ref := range eva.ObjectMeta.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return result
		}
	}

	mcs, err := r.getModelComparesForEvaluate(ctx, eva)
	if err != nil {
		log.Error(err, "Failed getting ModelCompares for Evaluate")
		return nil
	}
	if len(mcs) == 0 {
		return nil
	}

	for _, mc := range mcs {
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
	}

	return result
}

func (r *Reconciler) getModelComparesForEvaluate(ctx context.Context, eva *v1beta1.Evaluate) ([]*v1beta1.ModelCompare, error) {
	if len(eva.Labels) == 0 {
		return nil, fmt.Errorf("evaluate %v has no labels, this is unexpected", client.ObjectKeyFromObject(eva))
	}

	mcList := &v1beta1.ModelCompareList{}
	if err := r.client.List(ctx, mcList, client.InNamespace(eva.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ModelCompares, err: %w", err)
	}

	var mcs []*v1beta1.ModelCompare
	for idx := range mcList.Items {
		mc := &mcList.Items[idx]
		if labelsutil.HasMatchingLabels(mc.Spec.Selector, eva.Labels) {
			mcs = append(mcs, mc)
		}
	}

	return mcs, nil
}

// updateStatus updates the Status field for the ModelCompare
// It checks for the current state of the replicas and updates the Status of the ModelCompare.
func (r *Reconciler) updateStatus(ctx context.Context, mc *v1beta1.ModelCompare, filteredEvaluates []*v1beta1.Evaluate) error {
	newStatus := mc.Status.DeepCopy()
	evaluatePhase := make([]string, 0)
	for i := range filteredEvaluates {
		evaluatePhase = append(evaluatePhase, filteredEvaluates[i].Status.Phase)
	}

	newStatus.Phase = string(calculateComparePhase(evaluatePhase))

	if newStatus.Phase == string(v1beta1.EvaluatePhaseFailed) || newStatus.Phase == string(v1beta1.EvaluatePhaseSucceeded) {
		newStatus.EndedAt = ptr.To(metav1.Now())
	}

	if mc.Status.Phase != newStatus.Phase {
		newStatus.DeepCopyInto(&mc.Status)
	}

	return nil
}

func (r *Reconciler) createEvaluate(ctx context.Context, mc *v1beta1.ModelCompare, modelID int64) (*v1beta1.Evaluate, error) {
	log := ctrl.LoggerFrom(ctx)

	evaluate := r.computeDesiredEvaluate(mc, modelID, nil)

	// Create the Evaluate.
	if err := ssa.Patch(ctx, r.client, controllerName, evaluate); err != nil {
		record.Warnf(mc, "FailedCreate", "Failed to create evaluate %q: %v", evaluate.Name, err)
		conditions.MarkFalse(mc, v1beta1.MinersCreatedCondition, v1beta1.MinerCreationFailedReason,
			v1beta1.ConditionSeverityError, err.Error())

		log.Error(err, "Unable to create Evaluate")

		return nil, err
	}

	record.Eventf(mc, "SuccessfulCreate", "Created evaluate %q", evaluate.Name)
	return evaluate, nil
}

func (r *Reconciler) deleteEvaluate(ctx context.Context, mc *v1beta1.ModelCompare, eva *v1beta1.Evaluate) error {
	log := ctrl.LoggerFrom(ctx)

	if !eva.GetDeletionTimestamp().IsZero() {
		return nil
	}

	if err := r.client.Delete(ctx, eva); err != nil {
		log.Error(err, "Unable to delete Evaluate", "evaluate", klog.KObj(eva))
		record.Warnf(mc, "FailedDelete", "Failed to delete evaluate %q: %v", eva.Name, err)
		return err
	}
	log.V(2).Info("Deleted evaluate", "evaluate", klog.KObj(eva))
	record.Eventf(mc, "SuccessfulDelete", "Deleted evaluate %q", eva.Name)
	return nil
}

func calculateComparePhase(evaluatePhase []string) v1beta1.EvaluatePhase {
	// 注意：判断优先级不能乱
	if stringsutil.StringIn(string(v1beta1.EvaluatePhaseFailed), evaluatePhase) {
		return v1beta1.EvaluatePhaseFailed
	}

	if stringsutil.StringIn(string(v1beta1.EvaluatePhasePending), evaluatePhase) {
		return v1beta1.EvaluatePhasePending
	}

	if stringsutil.StringIn(string(v1beta1.EvaluatePhasePrepared), evaluatePhase) {
		return v1beta1.EvaluatePhasePrepared
	}

	if stringsutil.StringIn(string(v1beta1.EvaluatePhaseEvaluating), evaluatePhase) {
		return v1beta1.EvaluatePhaseEvaluating
	}

	return v1beta1.EvaluatePhaseSucceeded
}
