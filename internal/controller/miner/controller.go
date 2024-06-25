// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	builderruntime "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/superproj/onex/internal/controller/miner/apis/config"
	"github.com/superproj/onex/internal/pkg/feature"
	"github.com/superproj/onex/internal/pkg/util/annotations"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	logutil "github.com/superproj/onex/internal/pkg/util/log"
	minerutil "github.com/superproj/onex/internal/pkg/util/miner"
	"github.com/superproj/onex/internal/pkg/util/patch"
	"github.com/superproj/onex/internal/pkg/util/predicates"
	"github.com/superproj/onex/internal/pkg/util/ssa"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1/index"
	"github.com/superproj/onex/pkg/record"
)

const (
	// controllerName defines the controller used when creating clients.
	controllerName = "miner-controller"
)

// Reconciler reconciles a Miner object.
type Reconciler struct {
	client    client.Client
	APIReader client.Reader

	DryRun          bool
	ProviderClient  kubernetes.Interface
	RedisClient     *redis.Client
	ComponentConfig *config.MinerControllerConfiguration

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string

	// podDeletionRetryTimeout determines how long the controller will retry deleting a pod
	// during a single reconciliation.
	podDeletionRetryTimeout time.Duration
	ssaCache                ssa.Cache
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options, providerCluster cluster.Cluster) error {
	if r.podDeletionRetryTimeout.Nanoseconds() == 0 {
		r.podDeletionRetryTimeout = 10 * time.Second
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Miner{}).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))

	if !r.DryRun {
		builder = builder.WatchesRawSource(
			source.Kind(providerCluster.GetCache(), &corev1.Pod{}),
			handler.EnqueueRequestsFromMapFunc(r.PodToMiners),
			builderruntime.WithPredicates(
				predicates.All(ctrl.LoggerFrom(ctx),
					predicates.Any(ctrl.LoggerFrom(ctx), predicates.MinerSetUnpaused(ctrl.LoggerFrom(ctx))),
					predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
				),
			))
	}

	if _, err := builder.Build(r); err != nil {
		return fmt.Errorf("failed setting up with a controller manager: %w", err)
	}

	r.client = mgr.GetClient()
	r.ssaCache = ssa.NewCache()

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	if feature.DefaultFeatureGate.Enabled(feature.MachinePool) {
		log.Info("Enable feature gates", "featureGates", feature.MachinePool)
	}

	// 1. Fetch the Miner object
	m := &v1beta1.Miner{}
	if err := r.client.Get(ctx, rq.NamespacedName, m); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. AddOwners adds the owners of Machine as k/v pairs to the logger.
	ctx, log = logutil.AddOwners(ctx, m)
	log = log.WithValues("Chain", klog.KRef(m.ObjectMeta.Namespace, m.Spec.ChainName))
	ctx = ctrl.LoggerInto(ctx, log)

	// Return early if the object is paused.
	if annotations.IsPaused(m) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	helper, err := patch.NewHelper(m, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the Miner object and status after each reconciliation.
		r.reconcilePhase(ctx, m)

		// Always attempt to patch the object and status after each reconciliation.
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchMiner(ctx, helper, m, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(m, v1beta1.MinerFinalizer) {
		controllerutil.AddFinalizer(m, v1beta1.MinerFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle deletion reconciliation loop.
	if !m.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, m)
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, m)
}

// TODO ? 研究这个函数实现.
func patchMiner(ctx context.Context, helper *patch.Helper, miner *v1beta1.Miner, options ...patch.Option) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it
	// after provisioning - e.g. when a MHC condition exists - or during the deletion process).
	conditions.SetSummary(miner,
		conditions.WithConditions(
			// Infrastructure problems should take precedence over all the other conditions
			// v1beta1.InfrastructureReadyCondition,
			// Bootstrap comes after, but it is relevant only during initial miner provisioning.
			// v1beta1.BootstrapReadyCondition,
			// MHC reported condition should take precedence over the remediation progress
			v1beta1.MinerHealthCheckSucceededCondition,
			v1beta1.MinerOwnerRemediatedCondition,
		),
		// conditions.WithStepCounterIf(miner.ObjectMeta.DeletionTimestamp.IsZero() && miner.Spec.ProviderID == nil), // TODO? PodNotExists
		conditions.WithStepCounterIfOnly(
		// v1beta1.BootstrapReadyCondition,
		// v1beta1.InfrastructureReadyCondition,
		),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	// Also, if rquested, we are adding additional options like e.g. Patch ObservedGeneration when issuing the
	// patch at the end of the reconcile loop.
	options = append(options,
		patch.WithOwnedConditions{Conditions: []v1beta1.ConditionType{
			v1beta1.ReadyCondition,
			v1beta1.BootstrapReadyCondition,
			v1beta1.InfrastructureReadyCondition,
			v1beta1.DrainingSucceededCondition,
			v1beta1.MinerHealthCheckSucceededCondition,
			v1beta1.MinerOwnerRemediatedCondition,
		}},
	)

	return helper.Patch(ctx, miner, options...)
}

func (r *Reconciler) reconcile(ctx context.Context, m *v1beta1.Miner) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// TODO: support for owners across namespaces
	/*
		ch := &v1beta1.Chain{}
		key := client.ObjectKey{Namespace: metav1.NamespaceSystem, Name: m.Spec.ChainName}
		if err := r.client.Get(ctx, key, ch); err != nil {
			record.Warnf(m, "FailedCreate", "Failed to get chain %s: %v", key, err)
			return ctrl.Result{}, err
		}

		// If the miner is a stand-alone one, meaning not originated from a MinerDeployment, then set it as directly
		// owned by the Chain (if not already present).
		if r.shouldAdopt(m) {
			m.SetOwnerReferences(coreutil.EnsureOwnerRef(m.GetOwnerReferences(), metav1.OwnerReference{
				APIVersion: v1beta1.SchemeGroupVersion.String(),
				Kind:       v1beta1.SchemeGroupVersion.WithKind("Chain").Kind,
				Name:       ch.Name,
				UID:        ch.UID,
			}))
		}
	*/

	// Return early if the object is in failed phase.
	if minerIsFailed(m) {
		log.V(1).Info("Miner has gone `Failed` phase. It won't reconcile")
		return ctrl.Result{}, nil
	}

	phases := []func(context.Context, *v1beta1.Miner) (ctrl.Result, error){
		r.reconcileAnnotations,
		r.reconcileProviderPod,
		r.reconcileProviderService,
	}

	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, m)
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

func (r *Reconciler) reconcileDelete(ctx context.Context, m *v1beta1.Miner) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	helper, err := patch.NewHelper(m, r.client)
	if err != nil {
		log.Error(err, "Failed to new patch helper")
		return ctrl.Result{}, err
	}
	conditions.MarkFalse(m, v1beta1.MinerPodHealthyCondition, v1beta1.DeletingReason, v1beta1.ConditionSeverityInfo, "")
	if err := patchMiner(ctx, helper, m); err != nil {
		conditions.MarkFalse(m, v1beta1.MinerPodHealthyCondition, v1beta1.DeletionFailedReason, v1beta1.ConditionSeverityInfo, "")
		log.Error(err, "Failed to patch Miner")
		return ctrl.Result{}, err
	}

	if minerutil.IsGenesisMiner(m) {
		if err := r.ProviderClient.CoreV1().Services(m.Namespace).Delete(
			ctx,
			minerutil.GetProviderServiceName(m),
			metav1.DeleteOptions{},
		); err != nil &&
			!apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	log.Info("Deleting pod", "pod", klog.KRef(m.Namespace, m.Name))
	var deletePodErr error
	waitErr := wait.PollUntilContextTimeout(ctx, 2*time.Second, r.podDeletionRetryTimeout, true, func(ctx context.Context) (bool, error) {
		if deletePodErr = r.ProviderClient.CoreV1().Pods(m.Namespace).Delete(
			ctx,
			m.Name,
			metav1.DeleteOptions{},
		); deletePodErr != nil &&
			!apierrors.IsNotFound(deletePodErr) {
			return false, nil
		}

		return true, nil
	})

	if waitErr != nil {
		log.Error(deletePodErr, "Timed out deleting pod", "Pod", klog.KRef(m.Status.PodRef.Namespace, m.Status.PodRef.Name))
		conditions.MarkFalse(m, v1beta1.MinerPodHealthyCondition, v1beta1.DeletionFailedReason, v1beta1.ConditionSeverityWarning, "")
		record.Warnf(m, "FailedDeletePod", "error deleting Miner's pod: %v", deletePodErr)
		// If the pod deletion timeout is not expired yet, rqueue the Miner for reconciliation.
		if m.Spec.PodDeletionTimeout == nil || m.Spec.PodDeletionTimeout.Nanoseconds() == 0 ||
			m.DeletionTimestamp.Add(m.Spec.PodDeletionTimeout.Duration).After(time.Now()) {
			return ctrl.Result{}, deletePodErr
		}
		log.Info("Pod deletion timeout expired, continuing without Pod deletion")
	}

	controllerutil.RemoveFinalizer(m, v1beta1.MinerFinalizer)
	return ctrl.Result{}, nil
}

// shouldAdopt returns true if the Miner should be adopted as a stand-alone Miner directly owned by the Chain.
func (r *Reconciler) shouldAdopt(m *v1beta1.Miner) bool {
	// if the miner is controlled by something (MS or KCP), or if it is a stand-alone miner directly owned by the Chain, then no-op.
	if metav1.GetControllerOf(m) != nil || coreutil.HasOwner(m.GetOwnerReferences(), v1beta1.SchemeGroupVersion.String(), []string{"MinerSet"}) {
		return false
	}

	// Note: following checks are required because after restore from a backup both the Miner controller and the
	// MinerSet/ControlPlane controller are racing to adopt Miners.

	// If the Miner is originated by a MinerSet, it should not be adopted directly by the Chain as a stand-alone Miner.
	if _, ok := m.Labels[v1beta1.MinerSetNameLabel]; ok {
		return false
	}

	return true
}

func (r *Reconciler) PodToMiners(ctx context.Context, o client.Object) []ctrl.Request {
	pod, ok := o.(*corev1.Pod)
	if !ok {
		panic(fmt.Sprintf("Expected a Pod but got a %T", o))
	}

	var filters []client.ListOption
	// Match by minerName when the node has the annotation.
	if minerName, ok := pod.GetAnnotations()[v1beta1.MinerAnnotation]; ok {
		filters = append(filters, client.MatchingFields{"metadata.name": minerName})
	}

	// Match by podName and status.podRef.name.
	minerList := &v1beta1.MinerList{}
	if err := r.client.List(
		context.TODO(),
		minerList,
		append(filters, client.MatchingFields{index.MinerPodNameField: pod.Name})...); err != nil {
		return nil
	}

	// There should be exactly 1 Miner for the miner.
	if len(minerList.Items) == 1 {
		return []ctrl.Request{{NamespacedName: coreutil.ObjectKey(&minerList.Items[0])}}
	}

	return nil
}

func minerIsFailed(m *v1beta1.Miner) bool {
	return m.Status.Phase == string(v1beta1.MinerPhaseFailed)
}

// writer implements io.Writer interface as a pass-through for klog.
type writer struct {
	logFunc func(msg string, keysAndValues ...any)
}

// Write passes string(p) into writer's logFunc and always returns len(p).
func (w writer) Write(p []byte) (n int, err error) {
	w.logFunc(string(p))
	return len(p), nil
}
