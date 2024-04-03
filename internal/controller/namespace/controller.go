// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package controllermanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/metadata"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	v1clientset "github.com/superproj/onex/pkg/generated/clientset/versioned/typed/core/v1"
)

const controllerName = "controller-manager.namespace"

// NamespacedResourcesDeleter is a reconciler used to delete a namespace with all resources in it.
type NamespacedResourcesDeleter struct {
	client.Client
	// Client to manipulate the namespace.
	nsClient v1clientset.NamespaceInterface
	// Dynamic client to list and delete all namespaced resources.
	metadataClient metadata.Interface
	// Cache of what operations are not supported on each group version resource.
	opCache             *operationNotSupportedCache
	DiscoverResourcesFn func() ([]*metav1.APIResourceList, error)
	// The finalizer token that should be removed from the namespace
	// when all resources in that namespace have been deleted.
	finalizerToken v1.FinalizerName
}

// newReconciler returns a new reconcile.Reconciler.
func NewNamespacedResourcesDeleter(mgr manager.Manager, client clientset.Interface, metadataClient metadata.Interface) *NamespacedResourcesDeleter {
	return &NamespacedResourcesDeleter{
		Client:         mgr.GetClient(),
		nsClient:       client.CoreV1().Namespaces(),
		metadataClient: metadataClient,
		opCache: &operationNotSupportedCache{
			m: make(map[operationKey]bool),
		},
		finalizerToken:      v1.FinalizerKubernetes,
		DiscoverResourcesFn: client.Discovery().ServerPreferredNamespacedResources,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespacedResourcesDeleter) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		WithOptions(options).
		Named(controllerName).
		Complete(r)
}

// reconcile reads that state of the cluster for a RollingUpgrade object and makes changes based on the state read
// and the details in the RollingUpgrade.Spec.
func (r *NamespacedResourcesDeleter) Reconcile(ctx context.Context, rq ctrl.Request) (ctrl.Result, error) {
	namespace := &v1.Namespace{}
	err := r.Get(ctx, rq.NamespacedName, namespace)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if namespace.DeletionTimestamp == nil {
		return ctrl.Result{}, nil
	}

	klog.V(4).InfoS("Reconciling namespace", "namespace", klog.KObj(namespace), "finalizerToken", r.finalizerToken)

	// ensure that the status is up to date on the namespace
	// if we get a not found error, we assume the namespace is truly gone
	namespace, err = r.retryOnConflictError(namespace, r.updateNamespaceStatusFunc)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if namespace.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// return if it is already finalized.
	if finalized(namespace) {
		return ctrl.Result{}, nil
	}

	// there may still be content for us to remove
	estimate, err := r.deleteAllContent(namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if estimate > 0 {
		return ctrl.Result{}, &ResourcesRemainingError{estimate}
	}

	// we have removed content, so mark it finalized by us
	_, err = r.retryOnConflictError(namespace, r.finalizeNamespace)
	if err != nil {
		// in normal practice, this should not be possible, but if a deployment is running
		// two controllers to do namespace deletion that share a common finalizer token it's
		// possible that a not found could occur since the other controller would have finished the delete.
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NamespacedResourcesDeleter) initOpCache() {
	// pre-fill opCache with the discovery info
	//
	// TODO(sttts): get rid of opCache and http 405 logic around it and trust discovery info
	resources, err := r.DiscoverResourcesFn()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get all supported resources from server: %w", err))
	}
	if len(resources) == 0 {
		klog.Fatalf("Unable to get any supported resources from server: %v", err)
	}

	for _, rl := range resources {
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			klog.ErrorS(err, fmt.Sprintf("Failed to parse GroupVersion %q", rl.GroupVersion))
			continue
		}

		for _, rs := range rl.APIResources {
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: rs.Name}
			verbs := sets.NewString([]string(rs.Verbs)...)

			if !verbs.Has("delete") {
				klog.V(6).Infof("Skipping resource %v because it cannot be deleted.", gvr)
			}

			for _, op := range []operation{operationList, operationDeleteCollection} {
				if !verbs.Has(string(op)) {
					r.opCache.setNotSupported(operationKey{operation: op, gvr: gvr})
				}
			}
		}
	}
}

// ResourcesRemainingError is used to inform the caller that all resources are not yet fully removed from the namespace.
type ResourcesRemainingError struct {
	Estimate int64
}

func (e *ResourcesRemainingError) Error() string {
	return fmt.Sprintf("some content remains in the namespace, estimate %d seconds before it is removed", e.Estimate)
}

// operation is used for caching if an operation is supported on a dynamic client.
type operation string

const (
	operationDeleteCollection operation = "deletecollection"
	operationList             operation = "list"
	// assume a default estimate for finalizers to complete when found on items pending deletion.
	finalizerEstimateSeconds int64 = int64(15)
)

// operationKey is an entry in a cache.
type operationKey struct {
	operation operation
	gvr       schema.GroupVersionResource
}

// operationNotSupportedCache is a simple cache to remember if an operation is not supported for a resource.
// if the operationKey maps to true, it means the operation is not supported.
type operationNotSupportedCache struct {
	lock sync.RWMutex
	m    map[operationKey]bool
}

// isSupported returns true if the operation is supported.
func (o *operationNotSupportedCache) isSupported(key operationKey) bool {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return !o.m[key]
}

func (o *operationNotSupportedCache) setNotSupported(key operationKey) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.m[key] = true
}

// updateNamespaceFunc is a function that makes an update to a namespace.
type updateNamespaceFunc func(namespace *v1.Namespace) (*v1.Namespace, error)

// retryOnConflictError retries the specified fn if there was a conflict error
// it will return an error if the UID for an object changes across retry operations.
// TODO RetryOnConflict should be a generic concept in client code.
func (r *NamespacedResourcesDeleter) retryOnConflictError(namespace *v1.Namespace, fn updateNamespaceFunc) (result *v1.Namespace, err error) {
	latestNamespace := namespace
	for {
		result, err = fn(latestNamespace)
		if err == nil {
			return result, nil
		}
		if !apierrors.IsConflict(err) {
			return nil, err
		}
		prevNamespace := latestNamespace

		latestNamespace = &v1.Namespace{}
		// if err := r.Client.Get(context.TODO(), prevNamespace.Name, latestNamespace); err != nil {
		if err := r.Client.Get(context.TODO(), types.NamespacedName{}, latestNamespace); err != nil {
			return nil, err
		}

		if prevNamespace.UID != latestNamespace.UID {
			return nil, fmt.Errorf("namespace uid has changed across retries")
		}
	}
}

// updateNamespaceStatusFunc will verify that the status of the namespace is correct.
func (r *NamespacedResourcesDeleter) updateNamespaceStatusFunc(namespace *v1.Namespace) (*v1.Namespace, error) {
	if namespace.DeletionTimestamp.IsZero() || namespace.Status.Phase == v1.NamespaceTerminating {
		return namespace, nil
	}
	newNamespace := namespace.DeepCopy()
	newNamespace.Status.Phase = v1.NamespaceTerminating
	return newNamespace, r.Client.Status().Update(context.TODO(), newNamespace)
}

// finalized returns true if the namespace.Spec.Finalizers is an empty list.
func finalized(namespace *v1.Namespace) bool {
	return len(namespace.Spec.Finalizers) == 0
}

// finalizeNamespace removes the specified finalizerToken and finalizes the namespace.
func (r *NamespacedResourcesDeleter) finalizeNamespace(namespace *v1.Namespace) (*v1.Namespace, error) {
	namespaceFinalize := v1.Namespace{}
	namespaceFinalize.ObjectMeta = namespace.ObjectMeta
	namespaceFinalize.Spec = namespace.Spec
	finalizerSet := sets.NewString()
	for i := range namespace.Spec.Finalizers {
		if namespace.Spec.Finalizers[i] != r.finalizerToken {
			finalizerSet.Insert(string(namespace.Spec.Finalizers[i]))
		}
	}
	namespaceFinalize.Spec.Finalizers = make([]v1.FinalizerName, 0, len(finalizerSet))
	for _, value := range finalizerSet.List() {
		namespaceFinalize.Spec.Finalizers = append(namespaceFinalize.Spec.Finalizers, v1.FinalizerName(value))
	}
	namespace, err := r.nsClient.Finalize(context.Background(), &namespaceFinalize, metav1.UpdateOptions{})
	if err != nil {
		// it was removed already, so life is good
		if apierrors.IsNotFound(err) {
			return namespace, nil
		}
	}
	return namespace, err
}

// deleteCollection is a helper function that will delete the collection of resources
// it returns true if the operation was supported on the server.
// it returns an error if the operation was supported on the server but was unable to complete.
func (r *NamespacedResourcesDeleter) deleteCollection(gvr schema.GroupVersionResource, namespace string) (bool, error) {
	klog.V(5).Infof("namespace controller - deleteCollection - namespace: %s, gvr: %v", namespace, gvr)

	key := operationKey{operation: operationDeleteCollection, gvr: gvr}
	if !r.opCache.isSupported(key) {
		klog.V(5).Infof("namespace controller - deleteCollection ignored since not supported - namespace: %s, gvr: %v", namespace, gvr)
		return false, nil
	}

	// namespace controller does not want the garbage collector to insert the orphan finalizer since it calls
	// resource deletions generically.  it will ensure all resources in the namespace are purged prior to releasing
	// namespace itself.
	background := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{PropagationPolicy: &background}
	//nolint:kubelistcheck
	err := r.metadataClient.Resource(gvr).Namespace(namespace).DeleteCollection(context.TODO(), opts, metav1.ListOptions{})

	if err == nil {
		return true, nil
	}

	// this is strange, but we need to special case for both MethodNotSupported and NotFound errors
	// TODO: https://github.com/kubernetes/kubernetes/issues/22413
	// we have a resource returned in the discovery API that supports no top-level verbs:
	//  /apis/extensions/v1beta1/namespaces/default/replicationcontrollers
	// when working with this resource type, we will get a literal not found error rather than expected method not supported
	if errors.IsMethodNotSupported(err) || apierrors.IsNotFound(err) {
		klog.V(5).Infof("namespace controller - deleteCollection not supported - namespace: %s, gvr: %v", namespace, gvr)
		return false, nil
	}

	klog.V(5).Infof("namespace controller - deleteCollection unexpected error - namespace: %s, gvr: %v, error: %v", namespace, gvr, err)
	return true, err
}

// listCollection will list the items in the specified namespace
// it returns the following:
//
//	the list of items in the collection (if found)
//	a boolean if the operation is supported
//	an error if the operation is supported but could not be completed.
func (r *NamespacedResourcesDeleter) listCollection(gvr schema.GroupVersionResource, namespace string) (*metav1.PartialObjectMetadataList, bool, error) {
	klog.V(5).Infof("namespace controller - listCollection - namespace: %s, gvr: %v", namespace, gvr)

	key := operationKey{operation: operationList, gvr: gvr}
	if !r.opCache.isSupported(key) {
		klog.V(5).Infof("namespace controller - listCollection ignored since not supported - namespace: %s, gvr: %v", namespace, gvr)
		return nil, false, nil
	}

	//nolint:kubelistcheck
	partialList, err := r.metadataClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		return partialList, true, nil
	}

	// this is strange, but we need to special case for both MethodNotSupported and NotFound errors
	// TODO: https://github.com/kubernetes/kubernetes/issues/22413
	// we have a resource returned in the discovery API that supports no top-level verbs:
	//  /apis/extensions/v1beta1/namespaces/default/replicationcontrollers
	// when working with this resource type, we will get a literal not found error rather than expected method not supported
	if errors.IsMethodNotSupported(err) || apierrors.IsNotFound(err) {
		klog.V(5).Infof("namespace controller - listCollection not supported - namespace: %s, gvr: %v", namespace, gvr)
		return nil, false, nil
	}

	return nil, true, err
}

// deleteEachItem is a helper function that will list the collection of resources and delete each item 1 by 1.
func (r *NamespacedResourcesDeleter) deleteEachItem(gvr schema.GroupVersionResource, namespace string) error {
	klog.V(5).Infof("namespace controller - deleteEachItem - namespace: %s, gvr: %v", namespace, gvr)

	unstructuredList, listSupported, err := r.listCollection(gvr, namespace)
	if err != nil {
		return err
	}
	if !listSupported {
		return nil
	}
	for _, item := range unstructuredList.Items {
		background := metav1.DeletePropagationBackground
		opts := metav1.DeleteOptions{PropagationPolicy: &background}
		if err = r.metadataClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), item.GetName(), opts); err != nil && !apierrors.IsNotFound(err) && !errors.IsMethodNotSupported(err) {
			return err
		}
	}
	return nil
}

type gvrDeletionMetadata struct {
	// finalizerEstimateSeconds is an estimate of how much longer to wait.  zero means that no estimate has made and does not
	// mean that all content has been removed.
	finalizerEstimateSeconds int64
	// numRemaining is how many instances of the gvr remain
	numRemaining int
	// finalizersToNumRemaining maps finalizers to how many resources are stuck on them
	finalizersToNumRemaining map[string]int
}

// deleteAllContentForGroupVersionResource will use the dynamic client to delete each resource identified in gvr.
// It returns an estimate of the time remaining before the remaining resources are deleted.
// If estimate > 0, not all resources are guaranteed to be gone.
func (r *NamespacedResourcesDeleter) deleteAllContentForGroupVersionResource(
	gvr schema.GroupVersionResource, namespace string,
	namespaceDeletedAt metav1.Time,
) (gvrDeletionMetadata, error) {
	klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - namespace: %s, gvr: %v", namespace, gvr)

	// estimate how long it will take for the resource to be deleted (needed for objects that support graceful delete)
	estimate, err := r.estimateGracefulTermination(gvr, namespace, namespaceDeletedAt)
	if err != nil {
		klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - unable to estimate - namespace: %s, gvr: %v, err: %v", namespace, gvr, err)
		return gvrDeletionMetadata{}, err
	}
	klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - estimate - namespace: %s, gvr: %v, estimate: %v", namespace, gvr, estimate)

	// first try to delete the entire collection
	deleteCollectionSupported, err := r.deleteCollection(gvr, namespace)
	if err != nil {
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
	}

	// delete collection was not supported, so we list and delete each item...
	if !deleteCollectionSupported {
		err = r.deleteEachItem(gvr, namespace)
		if err != nil {
			return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
		}
	}

	// verify there are no more remaining items
	// it is not an error condition for there to be remaining items if local estimate is non-zero
	klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - checking for no more items in namespace: %s, gvr: %v", namespace, gvr)
	unstructuredList, listSupported, err := r.listCollection(gvr, namespace)
	if err != nil {
		klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - error verifying no items in namespace: %s, gvr: %v, err: %v", namespace, gvr, err)
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
	}
	if !listSupported {
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, nil
	}
	klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - items remaining - namespace: %s, gvr: %v, items: %v", namespace, gvr, len(unstructuredList.Items))
	if len(unstructuredList.Items) == 0 {
		// we're done
		return gvrDeletionMetadata{finalizerEstimateSeconds: 0, numRemaining: 0}, nil
	}

	// use the list to find the finalizers
	finalizersToNumRemaining := map[string]int{}
	for _, item := range unstructuredList.Items {
		for _, finalizer := range item.GetFinalizers() {
			finalizersToNumRemaining[finalizer]++
		}
	}

	if estimate != int64(0) {
		klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - estimate is present - namespace: %s, gvr: %v, finalizers: %v", namespace, gvr, finalizersToNumRemaining)
		return gvrDeletionMetadata{
			finalizerEstimateSeconds: estimate,
			numRemaining:             len(unstructuredList.Items),
			finalizersToNumRemaining: finalizersToNumRemaining,
		}, nil
	}

	// if any item has a finalizer, we treat that as a normal condition, and use a default estimation to allow for GC to complete.
	if len(finalizersToNumRemaining) > 0 {
		klog.V(5).Infof("namespace controller - deleteAllContentForGroupVersionResource - items remaining with finalizers - namespace: %s, gvr: %v, finalizers: %v", namespace, gvr, finalizersToNumRemaining)
		return gvrDeletionMetadata{
			finalizerEstimateSeconds: finalizerEstimateSeconds,
			numRemaining:             len(unstructuredList.Items),
			finalizersToNumRemaining: finalizersToNumRemaining,
		}, nil
	}

	// nothing reported a finalizer, so something was unexpected as it should have been deleted.
	return gvrDeletionMetadata{
		finalizerEstimateSeconds: estimate,
		numRemaining:             len(unstructuredList.Items),
	}, fmt.Errorf("unexpected items still remain in namespace: %s for gvr: %v", namespace, gvr)
}

type allGVRDeletionMetadata struct {
	// gvrToNumRemaining is how many instances of the gvr remain
	gvrToNumRemaining map[schema.GroupVersionResource]int
	// finalizersToNumRemaining maps finalizers to how many resources are stuck on them
	finalizersToNumRemaining map[string]int
}

// deleteAllContent will use the dynamic client to delete each resource identified in groupVersionResources.
// It returns an estimate of the time remaining before the remaining resources are deleted.
// If estimate > 0, not all resources are guaranteed to be gone.
func (r *NamespacedResourcesDeleter) deleteAllContent(ns *v1.Namespace) (int64, error) {
	namespace := ns.Name
	namespaceDeletedAt := *ns.DeletionTimestamp
	var errs []error
	conditionUpdater := namespaceConditionUpdater{}
	estimate := int64(0)
	klog.V(4).InfoS("namespace controller - deleteAllContent", "namespace", klog.KObj(ns))

	resources, err := r.DiscoverResourcesFn()
	if err != nil {
		// discovery errors are not fatal.  We often have some set of resources we can operate against even if we don't have a complete list
		errs = append(errs, err)
		conditionUpdater.ProcessDiscoverResourcesErr(err)
	}
	// TODO(sttts): get rid of opCache and pass the verbs (especially "deletecollection") down into the deleter
	deletableResources := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"delete"}}, resources)
	groupVersionResources, err := discovery.GroupVersionResources(deletableResources)
	if err != nil {
		// discovery errors are not fatal.  We often have some set of resources we can operate against even if we don't have a complete list
		errs = append(errs, err)
		conditionUpdater.ProcessGroupVersionErr(err)
	}

	numRemainingTotals := allGVRDeletionMetadata{
		gvrToNumRemaining:        map[schema.GroupVersionResource]int{},
		finalizersToNumRemaining: map[string]int{},
	}
	for gvr := range groupVersionResources {
		gvrDeletionMetadata, err := r.deleteAllContentForGroupVersionResource(gvr, namespace, namespaceDeletedAt)
		if err != nil {
			// If there is an error, hold on to it but proceed with all the remaining
			// groupVersionResources.
			errs = append(errs, err)
			conditionUpdater.ProcessDeleteContentErr(err)
		}
		if gvrDeletionMetadata.finalizerEstimateSeconds > estimate {
			estimate = gvrDeletionMetadata.finalizerEstimateSeconds
		}
		if gvrDeletionMetadata.numRemaining > 0 {
			numRemainingTotals.gvrToNumRemaining[gvr] = gvrDeletionMetadata.numRemaining
			for finalizer, numRemaining := range gvrDeletionMetadata.finalizersToNumRemaining {
				if numRemaining == 0 {
					continue
				}
				numRemainingTotals.finalizersToNumRemaining[finalizer] += numRemaining
			}
		}
	}
	conditionUpdater.ProcessContentTotals(numRemainingTotals)

	// we always want to update the conditions because if we have set a condition to "it worked" after it was previously, "it didn't work",
	// we need to reflect that information.  Recall that additional finalizers can be set on namespaces, so this finalizer may clear itself and
	// NOT remove the resource instance.
	if hasChanged := conditionUpdater.Update(ns); hasChanged {
		if err = r.Client.Status().Update(context.TODO(), ns); err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't update status condition for namespace %q: %w", namespace, err))
		}
	}

	// if len(errs)==0, NewAggregate returns nil.
	klog.V(4).Infof("namespace controller - deleteAllContent - namespace: %s, estimate: %v, errors: %v", namespace, estimate, utilerrors.NewAggregate(errs))
	return estimate, utilerrors.NewAggregate(errs)
}

// estimateGracefulTermination will estimate the graceful termination rquired for the specific entity in the namespace.
func (r *NamespacedResourcesDeleter) estimateGracefulTermination(gvr schema.GroupVersionResource, ns string, namespaceDeletedAt metav1.Time) (int64, error) {
	groupResource := gvr.GroupResource()
	klog.V(5).Infof("namespace controller - estimateGracefulTermination - group %s, resource: %s", groupResource.Group, groupResource.Resource)
	estimate := int64(0)
	var err error
	//nolint:gocritic
	switch groupResource {
	case schema.GroupResource{Group: "", Resource: "pods"}:
		estimate, err = r.estimateGracefulTerminationForPods(ns)
	}
	if err != nil {
		return 0, err
	}
	// determine if the estimate is greater than the deletion timestamp
	duration := time.Since(namespaceDeletedAt.Time)
	allowedEstimate := time.Duration(estimate) * time.Second
	if duration >= allowedEstimate {
		estimate = int64(0)
	}
	return estimate, nil
}

// estimateGracefulTerminationForPods determines the graceful termination period for pods in the namespace.
func (r *NamespacedResourcesDeleter) estimateGracefulTerminationForPods(ns string) (int64, error) {
	return 0, nil
}
