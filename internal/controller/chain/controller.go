// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package chain

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/superproj/onex/internal/controller/apis/config"
	"github.com/superproj/onex/internal/pkg/known"
	"github.com/superproj/onex/internal/pkg/util/annotations"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	logutil "github.com/superproj/onex/internal/pkg/util/log"
	"github.com/superproj/onex/internal/pkg/util/patch"
	"github.com/superproj/onex/internal/pkg/util/predicates"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/record"
)

const controllerName = "controller-manager.chain"

// chainKind contains the schema.GroupVersionKind for the Chain type.
var chainKind = v1beta1.SchemeGroupVersion.WithKind("Chain")

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=miners;miners/status;miners/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// Reconciler reconciles a Chain object.
type Reconciler struct {
	client          client.Client
	APIReader       client.Reader
	ComponentConfig *config.ChainControllerConfiguration

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Chain{}).
		Owns(&v1beta1.Miner{}).
		Watches(
			&v1beta1.Miner{},
			handler.EnqueueRequestsFromMapFunc(r.MinerToChains)).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))

	r.client = mgr.GetClient()

	return builder.Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	// Fetch the Chain instance
	ch := &v1beta1.Chain{}
	if err := r.client.Get(ctx, rq.NamespacedName, ch); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// AddOwners adds the owners of Chain as k/v pairs to the logger.
	ctx, log := logutil.AddOwners(ctx, ch)
	log.V(4).Info("Reconcile chain")

	// Return early if the object is paused.
	if annotations.IsPaused(ch) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	helper, err := patch.NewHelper(ch, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := helper.Patch(ctx, ch, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(ch, v1beta1.ChainFinalizer) {
		controllerutil.AddFinalizer(ch, v1beta1.ChainFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle deletion reconciliation loop.
	if !ch.GetDeletionTimestamp().IsZero() {
		controllerutil.RemoveFinalizer(ch, v1beta1.ChainFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, ch)
}

func (r *Reconciler) reconcile(ctx context.Context, ch *v1beta1.Chain) (ctrl.Result, error) {
	if ch.Spec.Image == "" {
		ch.Spec.Image = r.ComponentConfig.Image
	}

	phases := []func(context.Context, *v1beta1.Chain) (ctrl.Result, error){
		r.reconcileConfigMap,
		r.reconcileMiner,
	}

	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, ch)
		if err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			break
		}
		res = coreutil.LowestNonZeroResult(res, phaseResult)
	}

	return res, kerrors.NewAggregate(errs)
}

func (r *Reconciler) reconcileConfigMap(ctx context.Context, ch *v1beta1.Chain) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	reconciled, err := r.IsConfigMapReconciled(ctx, ch)
	if err != nil {
		return ctrl.Result{}, err
	}
	if reconciled {
		return ctrl.Result{}, nil
	}

	gv := v1beta1.SchemeGroupVersion
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    fmt.Sprintf("%s-", ch.Name),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ch, chainKind)},
			Namespace:       ch.Namespace,
			Labels:          map[string]string{v1beta1.ChainNameLabel: ch.Name},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("ConfigMap").Kind,
			APIVersion: gv.String(),
		},
		Data: map[string]string{
			known.GenericKeystoreFile: known.GenericKeystoreValue,
		},
	}

	if err := r.client.Create(ctx, cm); err != nil {
		record.Warnf(ch, "FailedCreate", "Failed to create configMap %q: %v", cm.Name, err)
		conditions.MarkFalse(ch, v1beta1.ConfigMapsCreatedCondition, v1beta1.ConfigMapCreationFailedReason,
			v1beta1.ConditionSeverityError, err.Error())

		log.Error(err, "Unable to create configMap")
		return ctrl.Result{}, err
	}

	ch.Status.ConfigMapRef = &v1beta1.LocalObjectReference{Name: cm.Name}

	log.V(2).Info("Created configMap", "configMap", klog.KObj(cm))
	record.Eventf(ch, "SuccessfulCreate", "Created configMap %q", cm.Name)

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileMiner(ctx context.Context, ch *v1beta1.Chain) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	reconciled, err := r.IsMinerReconciled(ctx, ch)
	if err != nil {
		return ctrl.Result{}, err
	}
	if reconciled {
		return ctrl.Result{}, nil
	}

	gv := v1beta1.SchemeGroupVersion
	miner := &v1beta1.Miner{
		ObjectMeta: metav1.ObjectMeta{
			// GenerateName:    ch.Name + "-",
			Name:            ch.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ch, chainKind)},
			Namespace:       ch.Namespace,
			Labels:          map[string]string{v1beta1.ChainNameLabel: ch.Name},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("Miner").Kind,
			APIVersion: gv.String(),
		},
		Spec: v1beta1.MinerSpec{
			MinerType: ch.Spec.MinerType,
			ChainName: ch.Name,
		},
	}

	if err := r.client.Create(ctx, miner); err != nil {
		record.Warnf(ch, "FailedCreate", "Failed to create miner %q: %v", miner.Name, err)
		conditions.MarkFalse(ch, v1beta1.MinersCreatedCondition, v1beta1.MinerCreationFailedReason,
			v1beta1.ConditionSeverityError, err.Error())

		log.Error(err, "Unable to create Miner")
		return ctrl.Result{}, err
	}

	if ch.Status.MinerRef == nil {
		ch.Status.MinerRef = &v1beta1.LocalObjectReference{Name: miner.Name}
	}

	log.V(2).Info("Created miner", "miner", klog.KObj(miner))
	record.Eventf(ch, "SuccessfulCreate", "Created miner %q", miner.Name)
	conditions.MarkTrue(ch, v1beta1.MinersCreatedCondition)

	return ctrl.Result{}, nil
}

func (r *Reconciler) IsMinerReconciled(ctx context.Context, ch *v1beta1.Chain) (bool, error) {
	log := ctrl.LoggerFrom(ctx)

	mList := &v1beta1.MinerList{}
	selectorMap := map[string]string{v1beta1.ChainNameLabel: ch.Name}
	if err := r.client.List(ctx, mList, client.InNamespace(ch.Namespace), client.MatchingLabels(selectorMap)); err != nil {
		log.Error(err, "Failed to list miners")
		return true, err
	}

	return len(mList.Items) != 0, nil
}

func (r *Reconciler) IsConfigMapReconciled(ctx context.Context, ch *v1beta1.Chain) (bool, error) {
	log := ctrl.LoggerFrom(ctx)

	cmList := &corev1.ConfigMapList{}
	selectorMap := map[string]string{v1beta1.ChainNameLabel: ch.Name}
	if err := r.client.List(ctx, cmList, client.InNamespace(ch.Namespace), client.MatchingLabels(selectorMap)); err != nil {
		log.Error(err, "Failed to list configMaps")
		return true, err
	}

	return len(cmList.Items) != 0, nil
}

// MinerToChains is a handler.ToRequestsFunc to be used to enqueue rquests for reconciliation
// for Chains that might adopt an orphaned Miner.
func (r *Reconciler) MinerToChains(ctx context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	m, ok := o.(*v1beta1.Miner)
	if !ok {
		panic(fmt.Sprintf("Expected a Miner but got a %T", o))
	}

	chainName, ok := m.Labels[v1beta1.ChainNameLabel]
	if !ok {
		klog.V(1).InfoS("Miner has no chain name label", "miner", klog.KObj(m))
		return nil
	}

	name := client.ObjectKey{Namespace: m.Namespace, Name: chainName}
	result = append(result, ctrl.Request{NamespacedName: name})

	return result
}
