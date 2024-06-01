// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerset

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/superproj/onex/internal/pkg/util/annotations"
	"github.com/superproj/onex/internal/pkg/util/collections"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	labelsutil "github.com/superproj/onex/internal/pkg/util/labels"
	logutil "github.com/superproj/onex/internal/pkg/util/log"
	minerutil "github.com/superproj/onex/internal/pkg/util/miner"
	"github.com/superproj/onex/internal/pkg/util/patch"
	"github.com/superproj/onex/internal/pkg/util/predicates"
	"github.com/superproj/onex/internal/pkg/util/ssa"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/record"
	retryutil "github.com/superproj/onex/pkg/util/retry"
)

// MaxConcurrency used to prevent the high load of onex-apiserver caused by excessive concurrency,
// it is necessary to limit the miner create/delete concurrency.
const MaxConcurrency = 30

const controllerName = "minerset-controller"

var (
	// msKind contains the schema.GroupVersionKind for the MinerSet type.
	msKind = v1beta1.SchemeGroupVersion.WithKind("MinerSet")

	// stateConfirmationTimeout is the amount of time allowed to wait for desired state.
	stateConfirmationTimeout = 10 * time.Second

	// stateConfirmationInterval is the amount of time between polling for the desired state.
	// The polling is against a local memory cache.
	stateConfirmationInterval = 100 * time.Millisecond
)

// Reconciler reconciles a MinerSet object.
type Reconciler struct {
	client    client.Client
	APIReader client.Reader

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
	ssaCache         ssa.Cache
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.MinerSet{}).
		Owns(&v1beta1.Miner{}).
		Watches(
			&v1beta1.Miner{},
			handler.EnqueueRequestsFromMapFunc(r.MinerToMinerSets)).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))

	r.client = mgr.GetClient()
	r.ssaCache = ssa.NewCache()
	r.APIReader = mgr.GetAPIReader()

	return builder.Complete(r)
}

// Reconcile reads that state of the OneX for a MinerSet object and makes changes based on the state read
// and what is in the MinerSet.Spec.
func (r *Reconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	// 1. Fetch the MinerSet object
	ms := &v1beta1.MinerSet{}
	if err := r.client.Get(ctx, rq.NamespacedName, ms); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2.  AddOwners adds the owners of MinerSet as k/v pairs to the logger.
	ctx, log := logutil.AddOwners(ctx, ms)
	log = log.WithValues("Chain", klog.KRef(ms.ObjectMeta.Namespace, ms.Spec.Template.Spec.ChainName))
	ctx = ctrl.LoggerInto(ctx, log)

	chain, err := coreutil.GetChainByName(ctx, r.client, metav1.NamespaceSystem, ms.Spec.Template.Spec.ChainName)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.V(4).Info("Reconcile minerset")
	// Return early if the object is paused.
	if annotations.IsPaused(ms) {
		log.V(2).Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	helper, err := patch.NewHelper(ms, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchMinerSet(ctx, helper, ms); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Ignore deleted MinerSets, this can happen when foregroundDeletion is enabled
	if !ms.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	result, err := r.reconcile(ctx, chain, ms)
	if err != nil {
		log.Error(err, "Failed to reconcile MinerSet")
		record.Warnf(ms, "ReconcileError", "%v", err)
	}
	return result, err
}

func patchMinerSet(ctx context.Context, helper *patch.Helper, ms *v1beta1.MinerSet, options ...patch.Option) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	conditions.SetSummary(ms,
		conditions.WithConditions(
			v1beta1.MinersCreatedCondition,
			v1beta1.ResizedCondition,
			v1beta1.MinersReadyCondition,
		),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	options = append(options,
		patch.WithOwnedConditions{Conditions: []v1beta1.ConditionType{
			v1beta1.ReadyCondition,
			v1beta1.MinersCreatedCondition,
			v1beta1.ResizedCondition,
			v1beta1.MinersReadyCondition,
		}},
	)
	return helper.Patch(ctx, ms, options...)
}

func (r *Reconciler) reconcile(ctx context.Context, chain *v1beta1.Chain, ms *v1beta1.MinerSet) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Reconcile and retrieve the MinerSet object.
	if ms.Labels == nil {
		ms.Labels = make(map[string]string)
	}
	ms.Labels[v1beta1.ChainNameLabel] = ms.Spec.Template.Spec.ChainName

	// If the miner set is a stand alone one, meaning not originated from a MinerDeployment, then set it as directly
	// TODO: support for owners across namespaces
	/*
		if r.shouldAdopt(ms) {
			ms.SetOwnerReferences(coreutil.EnsureOwnerRef(ms.GetOwnerReferences(), metav1.OwnerReference{
				APIVersion: v1beta1.SchemeGroupVersion.String(),
				Kind:       v1beta1.SchemeGroupVersion.WithKind("Chain").Kind,
				Name:       chain.Name,
				UID:        chain.UID,
			}))
		}
	*/

	// Reconcile and retrieve the MinerSet object.
	selectorMap, err := metav1.LabelSelectorAsMap(&ms.Spec.Selector)
	if err != nil {
		log.Error(err, "Failed to convert MinerSet label selector to a map")
		return ctrl.Result{}, err
	}

	// Get all Miners linked to this MinerSet.
	allMiners := &v1beta1.MinerList{}
	if err := r.client.List(ctx, allMiners, client.InNamespace(ms.Namespace), client.MatchingLabels(selectorMap)); err != nil {
		log.Error(err, "Failed to list miners")
		return ctrl.Result{}, err
	}

	// Filter out irrelevant miners (i.e. IsControlledBy something else) and claim orphaned miners.
	// Miners in deleted state are deliberately not excluded.
	filteredMiners := make([]*v1beta1.Miner, 0, len(allMiners.Items))
	for idx := range allMiners.Items {
		miner := &allMiners.Items[idx]
		if shouldExcludeMiner(ms, miner) {
			continue
		}

		// Attempt to adopt miner if it meets previous conditions and it has no controller references.
		if metav1.GetControllerOf(miner) == nil {
			if err := r.adoptOrphan(ctx, ms, miner); err != nil {
				log.Error(err, "Failed to adopt Miner", "miner", klog.KObj(miner))
				record.Warnf(ms, "FailedAdopt", "Failed to adopt Miner %q: %v", miner.Name, err)
				continue
			}
			log.V(2).Info("Adopted Miner", "miner", klog.KObj(miner))
			record.Eventf(ms, "SuccessfulAdopt", "Adopted Miner %q", miner.Name)
		}

		filteredMiners = append(filteredMiners, miner)
	}

	result := ctrl.Result{}
	unHealthyResult, err := r.reconcileUnhealthyMiners(ctx, filteredMiners)
	if err != nil {
		return ctrl.Result{}, err
	}
	result = coreutil.LowestNonZeroResult(result, unHealthyResult)

	if err := r.syncMiners(ctx, ms, filteredMiners); err != nil {
		return ctrl.Result{}, err
	}

	syncResult, syncErr := r.syncReplicas(ctx, ms, filteredMiners)
	result = coreutil.LowestNonZeroResult(result, syncResult)

	// Always updates status as miners come up or die.
	if err := r.updateStatus(ctx, ms, filteredMiners); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update MinerSet's Status, err: %w", kerrors.NewAggregate([]error{err, syncErr}))
	}

	if syncErr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to sync MinerSet replicas, err: %w", syncErr)
	}

	var replicas int32
	if ms.Spec.Replicas != nil {
		replicas = *ms.Spec.Replicas
	}

	// Resync the MinerSet after MinReadySeconds as a last line of defense to guard against clock-skew.
	// Clock-skew is an issue as it may impact whether an available replica is counted as a ready replica.
	// A replica is available if the amount of time since last transition exceeds MinReadySeconds.
	// If there was a clock skew, checking whether the amount of time since last transition to ready state
	// exceeds MinReadySeconds could be incorrect.
	// To avoid an available replica stuck in the ready state, we force a reconcile after MinReadySeconds,
	// at which point it should confirm any available replica to be available.
	if ms.Spec.MinReadySeconds > 0 &&
		ms.Status.ReadyReplicas == replicas &&
		ms.Status.AvailableReplicas != replicas {
		minReadyResult := ctrl.Result{RequeueAfter: time.Duration(ms.Spec.MinReadySeconds) * time.Second}
		result = coreutil.LowestNonZeroResult(result, minReadyResult)
		return result, nil
	}

	// Quickly reconcile until the pods become Ready.
	if ms.Status.ReadyReplicas != replicas {
		log.V(4).Info("Some miners are not ready yet, rqueuing until they are ready")
		result = coreutil.LowestNonZeroResult(result, ctrl.Result{RequeueAfter: 15 * time.Second})
		return result, nil
	}

	return result, nil
}

// syncReplicas scales Miner resources up or down.
func (r *Reconciler) syncReplicas(ctx context.Context, ms *v1beta1.MinerSet, miners []*v1beta1.Miner) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	if ms.Spec.Replicas == nil {
		return ctrl.Result{}, fmt.Errorf("the Replicas field in Spec for minerset %v is nil, this should not be allowed", ms.Name)
	}
	diff := len(miners) - int(*(ms.Spec.Replicas))
	switch {
	case diff < 0:
		diff *= -1
		log.V(2).Info(fmt.Sprintf("MinerSet is scaling up to %d replicas by creating %d miners", *(ms.Spec.Replicas), diff), "miners", len(miners))
		if ms.Annotations != nil {
			if _, ok := ms.Annotations[v1beta1.DisableMinerCreateAnnotation]; ok {
				log.V(2).Info("Automatic creation of new miners disabled for miner set")
				return ctrl.Result{}, nil
			}
		}

		minerList, err := r.createMiners(ctx, ms, concurrencyNum(diff))
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, r.waitForMinerCreation(ctx, minerList)
	case diff > 0:
		log.V(2).Info(
			fmt.Sprintf("MinerSet is scaling down to %d replicas by deleting %d machines", *(ms.Spec.Replicas), diff),
			"miners", len(miners),
			"deletePolicy", ms.Spec.DeletePolicy,
		)

		deletePriorityFunc, err := getDeletePriorityFunc(ms)
		if err != nil {
			log.Error(err, "Unable to obtain delete priority function")
			return ctrl.Result{}, err
		}

		minersToDelete := getMinersToDeletePrioritized(miners, diff, deletePriorityFunc)

		minersToConcurrencyDelete := minersToDelete[0:concurrencyNum(len(minersToDelete))]

		// In order to reduce the pressure on the onex-apiserver, for each minerset,
		// only `MaxDeleteConcurrency` miners are allowed to be deleted at the same time
		if err := r.deleteMiners(ctx, ms, minersToConcurrencyDelete); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, r.waitForMinerDeletion(ctx, minersToConcurrencyDelete)
	}

	return ctrl.Result{}, nil
}

// syncMiners updates Miners to propagate in-place mutable fields
// from the MinerSet.
// Note: It also cleans up managed fields of all Miners so that Miners that were
// created/patched before (< v1.4.0) the controller adopted Server-Side-Apply (SSA) can also work with SSA.
// Otherwise fields would be co-owned by our "old" "manager" and "onex-minerset" and then we would not be
// able to e.g. drop labels and annotations.
func (r *Reconciler) syncMiners(ctx context.Context, ms *v1beta1.MinerSet, miners []*v1beta1.Miner) error {
	// Also equal to `log := klog.FromContext(ctx)`.
	// ctrl.LoggerFrom and klog.FromContext(ctx) use the same context key: github.com/go-logr/logr.contextKey
	log := ctrl.LoggerFrom(ctx)

	for i := range miners {
		m := miners[i]
		// If the miner is already being deleted, we don't need to update it.
		if !m.DeletionTimestamp.IsZero() {
			continue
		}

		// Cleanup managed fields of all Miners.
		// We do this so that Miners that were created/patched before the controller adopted Server-Side-Apply (SSA)
		// (< v1.4.0) can also work with SSA. Otherwise, fields would be co-owned by our "old" "manager" and
		// "capi-minerset" and then we would not be able to e.g. drop labels and annotations.
		if err := ssa.CleanUpManagedFieldsForSSAAdoption(ctx, r.client, m, controllerName); err != nil {
			return fmt.Errorf("failed to update miner: failed to adjust the managedFields of the Miner %q, err: %w", m.Name, err)
		}

		// Update Miner to propagate in-place mutable fields from the MinerSet.
		updatedMiner := r.computeDesiredMiner(ms, m)
		err := ssa.Patch(ctx, r.client, controllerName, updatedMiner, ssa.WithCachingProxy{Cache: r.ssaCache, Original: m})
		if err != nil {
			log.Error(err, "failed to update Miner", "Miner", klog.KObj(updatedMiner))
			return fmt.Errorf("failed to update Miner %q, err: %w", klog.KObj(updatedMiner), err)
		}
		miners[i] = updatedMiner
	}

	return nil
}

func (r *Reconciler) reconcileUnhealthyMiners(ctx context.Context, filteredMiners []*v1beta1.Miner) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	var errs []error
	for _, miner := range filteredMiners {
		// filteredMiners contains miners in deleting status to calculate correct status.
		// skip remediation for those in deleting status.
		if !miner.DeletionTimestamp.IsZero() {
			continue
		}

		if !conditions.IsFalse(miner, v1beta1.MinerOwnerRemediatedCondition) {
			continue
		}

		log.V(2).Info("Deleting miner because marked as unhealthy by the MinerHealthCheck controller", "miner", klog.KObj(miner))
		patch := client.MergeFrom(miner.DeepCopy())
		if err := r.client.Delete(ctx, miner); err != nil {
			log.Error(err, "Failed to delete", "miner", klog.KObj(miner))
			errs = append(errs, err)
			continue
		}
		conditions.MarkTrue(miner, v1beta1.MinerOwnerRemediatedCondition)
		if err := r.client.Status().Patch(ctx, miner, patch); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to update status", "miner", klog.KObj(miner))
			errs = append(errs, err)
		}
	}

	if err := kerrors.NewAggregate(errs); err != nil {
		log.Error(err, "Failed while deleting unhealthy miners")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// computeDesiredMiner computes the desired Miner.
// This Miner will be used during reconciliation to:
// * create a Miner
// * update an existing Miner
// Because we are using Server-Side-Apply we always have to calculate the full object.
// There are small differences in how we calculate the Miner depending on if it
// is a create or update. Example: for a new Miner we have to calculate a new name,
// while for an existing Miner we have to use the name of the existing Miner.
func (r *Reconciler) computeDesiredMiner(ms *v1beta1.MinerSet, existingMiner *v1beta1.Miner) *v1beta1.Miner {
	gv := v1beta1.SchemeGroupVersion
	desiredMiner := &v1beta1.Miner{
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("Miner").Kind,
			APIVersion: gv.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-", ms.Name)),
			Namespace: ms.Namespace,
			// Note: By setting the ownerRef on creation we signal to the Miner controller that this is not a stand-alone Miner.
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ms, msKind)},
			Labels:          ms.Spec.Template.Labels,
			Annotations:     ms.Spec.Template.Annotations,
			Finalizers:      []string{v1beta1.MinerFinalizer},
		},
		Spec: *ms.Spec.Template.Spec.DeepCopy(),
	}

	// If we are updating an existing Miner reuse the name and uid
	// from the existingMiner.
	// Note: we use UID to force SSA to update the existing Miner and to not accidentally create a new Miner.
	// infrastructureRef and bootstrap.configRef remain the same for an existing Miner.
	if existingMiner != nil {
		desiredMiner.SetName(existingMiner.Name)
		desiredMiner.SetUID(existingMiner.UID)
	}

	// Set Labels
	desiredMiner.Labels = minerLabelsFromMinerSet(ms)

	// Set Annotations
	desiredMiner.Annotations = minerAnnotationsFromMinerSet(ms)

	// Set all other in-place mutable fields.
	desiredMiner.Spec.PodDeletionTimeout = ms.Spec.Template.Spec.PodDeletionTimeout

	return desiredMiner
}

// getNewMiner creates a new Miner object. The name of the newly created resource is going
// to be created by the API server, we set the generateName field.
func (r *Reconciler) getNewMiner(ms *v1beta1.MinerSet) *v1beta1.Miner {
	gv := v1beta1.SchemeGroupVersion
	miner := &v1beta1.Miner{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    fmt.Sprintf("%s-", ms.Name),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ms, msKind)},
			Namespace:       ms.Namespace,
			Labels:          ms.Spec.Template.Labels,
			Annotations:     ms.Spec.Template.Annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("Miner").Kind,
			APIVersion: gv.String(),
		},
		Spec: ms.Spec.Template.Spec,
	}
	// miner.Spec.ClusterName = ms.Spec.ClusterName
	if miner.Labels == nil {
		miner.Labels = make(map[string]string)
	}
	return miner
}

// minerLabelsFromMinerSet computes the labels the Miner created from this MinerSet should have.
func minerLabelsFromMinerSet(ms *v1beta1.MinerSet) map[string]string {
	minerLabels := map[string]string{}
	// Note: We can't just set `ms.Spec.Template.Labels` directly and thus "share" the labels
	// map between Miner and ms.Spec.Template.Labels. This would mean that adding the
	// MinerSetNameLabel and MinerDeploymentNameLabel later on the Miner would also add the labels
	// to ms.Spec.Template.Labels and thus modify the labels of the MinerSet.
	for k, v := range ms.Spec.Template.Labels {
		minerLabels[k] = v
	}
	// Always set the MinerSetNameLabel.
	// Note: If a client tries to create a MinerSet without a selector, the MinerSet webhook
	// will add this label automatically. But we want this label to always be present even if the MinerSet
	// has a selector which doesn't include it. Therefore, we have to set it here explicitly.
	minerLabels[v1beta1.MinerSetNameLabel] = labelsutil.MustFormatValue(ms.Name)
	return minerLabels
}

// minerAnnotationsFromMinerSet computes the annotations the Miner created from this MinerSet should have.
func minerAnnotationsFromMinerSet(ms *v1beta1.MinerSet) map[string]string {
	annotations := map[string]string{}
	for k, v := range ms.Spec.Template.Annotations {
		annotations[k] = v
	}
	return annotations
}

// shouldExcludeMiner returns true if the miner should be filtered out, false otherwise.
func shouldExcludeMiner(ms *v1beta1.MinerSet, miner *v1beta1.Miner) bool {
	if metav1.GetControllerOf(miner) != nil && !metav1.IsControlledBy(miner, ms) {
		return true
	}

	return false
}

// adoptOrphan sets the MinerSet as a controller OwnerReference to the Miner.
func (r *Reconciler) adoptOrphan(ctx context.Context, ms *v1beta1.MinerSet, miner *v1beta1.Miner) error {
	patch := client.MergeFrom(miner.DeepCopy())
	newRef := *metav1.NewControllerRef(ms, msKind)
	miner.OwnerReferences = append(miner.OwnerReferences, newRef)
	return r.client.Patch(ctx, miner, patch)
}

func (r *Reconciler) waitForMinerCreation(ctx context.Context, minerList []*v1beta1.Miner) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(minerList); i++ {
		miner := minerList[i]
		pollErr := retryutil.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			key := client.ObjectKey{Namespace: miner.Namespace, Name: miner.Name}
			if err := r.client.Get(ctx, key, &v1beta1.Miner{}); err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}

			return true, nil
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for miner object to be created", "miner", klog.KObj(miner))
			return pollErr
		}
	}

	return nil
}

func (r *Reconciler) waitForMinerDeletion(ctx context.Context, minerList []*v1beta1.Miner) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(minerList); i++ {
		miner := minerList[i]
		pollErr := retryutil.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			m := &v1beta1.Miner{}
			key := client.ObjectKey{Namespace: miner.Namespace, Name: miner.Name}
			err := r.client.Get(ctx, key, m)
			if apierrors.IsNotFound(err) || !m.DeletionTimestamp.IsZero() {
				return true, nil
			}
			return false, err
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for miner object to be deleted", "miner", klog.KObj(miner))
			return pollErr
		}
	}
	return nil
}

// MinerToMinerSets is a handler.ToRequestsFunc to be used to enqueue rquests for reconciliation
// for MinerSets that might adopt an orphaned Miner.
func (r *Reconciler) MinerToMinerSets(ctx context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	m, ok := o.(*v1beta1.Miner)
	if !ok {
		panic(fmt.Sprintf("Expected a Miner but got a %T", o))
	}

	log := ctrl.LoggerFrom(ctx, "Miner", klog.KObj(m)) // TODO: test here

	// Check if the controller reference is already set and
	// return an empty result when one is found.
	for _, ref := range m.ObjectMeta.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return result
		}
	}

	mss, err := r.getMinerSetsForMiner(ctx, m)
	if err != nil {
		log.Error(err, "Failed getting MinerSets for Miner")
		return nil
	}
	if len(mss) == 0 {
		return nil
	}

	for _, ms := range mss {
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(ms)})
	}

	return result
}

func (r *Reconciler) getMinerSetsForMiner(ctx context.Context, m *v1beta1.Miner) ([]*v1beta1.MinerSet, error) {
	if len(m.Labels) == 0 {
		return nil, fmt.Errorf("miner %v has no labels, this is unexpected", client.ObjectKeyFromObject(m))
	}

	msList := &v1beta1.MinerSetList{}
	if err := r.client.List(ctx, msList, client.InNamespace(m.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list MinerSets, err: %w", err)
	}

	var mss []*v1beta1.MinerSet
	for idx := range msList.Items {
		ms := &msList.Items[idx]
		if labelsutil.HasMatchingLabels(ms.Spec.Selector, m.Labels) {
			mss = append(mss, ms)
		}
	}

	return mss, nil
}

// shouldAdopt returns true if the MinerSet should be adopted as a stand-alone MinerSet directly owned by the Chain.
func (r *Reconciler) shouldAdopt(ms *v1beta1.MinerSet) bool {
	// if the MinerSet is controlled by a MinerDeployment, or if it is a stand-alone MinerSet directly owned by the Chain, then no-op.
	if coreutil.HasOwner(ms.GetOwnerReferences(), v1beta1.SchemeGroupVersion.String(), []string{"MinerDeployment"}) {
		return false
	}

	// If the MinerSet is originated by a MinerDeployment object, it should not be adopted directly by the Chain as a stand-alone MinerSet.
	// Note: this is rquired because after restore from a backup both the MinerSet controller and the
	// MinerDeployment controller are racing to adopt MinerSets.
	if _, ok := ms.Labels[v1beta1.MinerDeploymentNameLabel]; ok {
		return false
	}

	return true
}

// updateStatus updates the Status field for the MinerSet
// It checks for the current state of the replicas and updates the Status of the MinerSet.
func (r *Reconciler) updateStatus(ctx context.Context, ms *v1beta1.MinerSet, filteredMiners []*v1beta1.Miner) error {
	log := ctrl.LoggerFrom(ctx)

	newStatus := ms.Status.DeepCopy()

	/*
		// Copy label selector to its status counterpart in string format.
		// This is necessary for CRDs including scale subresources.
		selector, err := metav1.LabelSelectorAsSelector(&ms.Spec.Selector)
		if err != nil {
			return errors.Wrapf(err, "failed to update status for MinerSet %s/%s", ms.Namespace, ms.Name)
		}
		newStatus.Selector = selector.String()
	*/

	// Count the number of miners that have labels matching the labels of the miner
	// template of the replica set, the matching miners may have more
	// labels than are in the template. Because the label of minerTemplateSpec is
	// a superset of the selector of the replica set, so the possible
	// matching miners must be part of the filteredMiners.
	fullyLabeledReplicasCount := 0
	readyReplicasCount := 0
	availableReplicasCount := 0
	desiredReplicas := *ms.Spec.Replicas
	templateLabel := labels.Set(ms.Spec.Template.Labels).AsSelectorPreValidated()

	for _, miner := range filteredMiners {
		log := log.WithValues("Miner", klog.KObj(miner))

		if templateLabel.Matches(labels.Set(miner.Labels)) {
			fullyLabeledReplicasCount++
		}

		if miner.Status.PodRef == nil {
			log.V(4).Info("Waiting for the miner controller to set status.PodRef on the Miner")
			continue
		}

		if minerutil.IsMinerReady(miner) {
			readyReplicasCount++
			if minerutil.IsMinerAvailable(miner, ms.Spec.MinReadySeconds, metav1.Now()) {
				availableReplicasCount++
			}
		} else if miner.GetDeletionTimestamp().IsZero() {
			log.V(2).Info("Waiting for the Kubernetes pod on the miner to report ready state")
		}
	}

	newStatus.Replicas = int32(len(filteredMiners))
	newStatus.FullyLabeledReplicas = int32(fullyLabeledReplicasCount)
	newStatus.ReadyReplicas = int32(readyReplicasCount)
	newStatus.AvailableReplicas = int32(availableReplicasCount)

	// Copy the newly calculated status into the minerset
	if ms.Status.Replicas != newStatus.Replicas ||
		ms.Status.FullyLabeledReplicas != newStatus.FullyLabeledReplicas ||
		ms.Status.ReadyReplicas != newStatus.ReadyReplicas ||
		ms.Status.AvailableReplicas != newStatus.AvailableReplicas ||
		ms.Generation != ms.Status.ObservedGeneration {
		log.V(4).Info("Updating status: " +
			fmt.Sprintf("replicas %d->%d (need %d), ", ms.Status.Replicas, newStatus.Replicas, desiredReplicas) +
			fmt.Sprintf("fullyLabeledReplicas %d->%d, ", ms.Status.FullyLabeledReplicas, newStatus.FullyLabeledReplicas) +
			fmt.Sprintf("readyReplicas %d->%d, ", ms.Status.ReadyReplicas, newStatus.ReadyReplicas) +
			fmt.Sprintf("availableReplicas %d->%d, ", ms.Status.AvailableReplicas, newStatus.AvailableReplicas) +
			fmt.Sprintf("sequence No: %v->%v", ms.Status.ObservedGeneration, newStatus.ObservedGeneration))

		// Save the generation number we acted on, otherwise we might wrongfully indicate
		// that we've seen a spec update when we retry.
		newStatus.ObservedGeneration = ms.Generation
		newStatus.DeepCopyInto(&ms.Status)
	}
	switch {
	// We are scaling up
	case newStatus.Replicas < desiredReplicas:
		conditions.MarkFalse(
			ms,
			v1beta1.ResizedCondition,
			v1beta1.ScalingUpReason,
			v1beta1.ConditionSeverityWarning,
			"Scaling up MinerSet to %d replicas (actual %d)", desiredReplicas, newStatus.Replicas,
		)
	// We are scaling down
	case newStatus.Replicas > desiredReplicas:
		conditions.MarkFalse(
			ms,
			v1beta1.ResizedCondition,
			v1beta1.ScalingDownReason,
			v1beta1.ConditionSeverityWarning,
			"Scaling down MinerSet to %d replicas (actual %d)", desiredReplicas, newStatus.Replicas,
		)
		// This means that there was no error in generating the desired number of miner objects
		conditions.MarkTrue(ms, v1beta1.MinersCreatedCondition)
	default:
		// Make sure last resize operation is marked as completed.
		// NOTE: we are checking the number of miners ready so we report resize completed only when the miners
		// are actually provisioned (vs reporting completed immediately after the last miner object is created). This convention is also used by KCP.
		if newStatus.ReadyReplicas == newStatus.Replicas {
			if conditions.IsFalse(ms, v1beta1.ResizedCondition) {
				log.V(2).Info("All the replicas are ready", "replicas", newStatus.ReadyReplicas)
			}
			conditions.MarkTrue(ms, v1beta1.ResizedCondition)
		}
		// This means that there was no error in generating the desired number of miner objects
		conditions.MarkTrue(ms, v1beta1.MinersCreatedCondition)
	}

	// Aggregate the operational state of all the miners; while aggregating we are adding the
	// source ref (reason@miner/name) so the problem can be easily tracked down to its source miner.
	conditions.SetAggregate(
		ms,
		v1beta1.MinersReadyCondition,
		collections.FromMiners(filteredMiners...).ConditionGetters(),
		conditions.AddSourceRef(),
		conditions.WithStepCounterIf(false),
	)

	return nil
}

func (r *Reconciler) createMiners(ctx context.Context, ms *v1beta1.MinerSet, concurrent int) ([]*v1beta1.Miner, error) {
	log := ctrl.LoggerFrom(ctx)

	var (
		miners []*v1beta1.Miner
		mu     sync.Mutex
	)

	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < concurrent; i++ {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				miner := r.computeDesiredMiner(ms, nil)

				// Create the Miner.
				if err := ssa.Patch(ctx, r.client, controllerName, miner); err != nil {
					record.Warnf(ms, "FailedCreate", "Failed to create miner %q: %v", miner.Name, err)
					conditions.MarkFalse(ms, v1beta1.MinersCreatedCondition, v1beta1.MinerCreationFailedReason,
						v1beta1.ConditionSeverityError, err.Error())

					log.Error(err, "Unable to create Miner")

					// Try to cleanup the external objects if the Miner creation failed.
					// ...

					return err
				}

				log.V(2).Info("Created miner", "miner", klog.KObj(miner))
				record.Eventf(ms, "SuccessfulCreate", "Created miner %q", miner.Name)

				mu.Lock()
				miners = append(miners, miner)
				mu.Unlock()

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return miners, nil
}

func (r *Reconciler) deleteMiners(ctx context.Context, ms *v1beta1.MinerSet, minersToDelete []*v1beta1.Miner) error {
	log := ctrl.LoggerFrom(ctx)

	eg, ctx := errgroup.WithContext(ctx)

	for _, miner := range minersToDelete {
		m := miner
		if !m.GetDeletionTimestamp().IsZero() {
			continue
		}

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				if err := r.client.Delete(ctx, m); err != nil {
					log.Error(err, "Unable to delete Miner", "miner", klog.KObj(m))
					record.Warnf(ms, "FailedDelete", "Failed to delete miner %q: %v", m.Name, err)
					return err
				}
				log.V(2).Info("Deleted miner", "miner", klog.KObj(m))
				record.Eventf(ms, "SuccessfulDelete", "Deleted miner %q", m.Name)
				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func concurrencyNum(actualNum int) int {
	if actualNum <= MaxConcurrency {
		return actualNum
	}

	return MaxConcurrency
}
