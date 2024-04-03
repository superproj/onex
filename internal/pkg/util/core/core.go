// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package core implements core utilities.
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sversion "k8s.io/apimachinery/pkg/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/superproj/onex/internal/pkg/util/contract"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const (
	// CharSet defines the alphanumeric set for random string generation.
	CharSet = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

	// ErrNoChain is returned when the cluster
	// label could not be found on the object passed in.
	ErrNoChain = fmt.Errorf("no %q label present", v1beta1.ChainNameLabel)

	// ErrUnstructuredFieldNotFound determines that a field
	// in an unstructured object could not be found.
	ErrUnstructuredFieldNotFound = fmt.Errorf("field not found")
)

// RandomString returns a random alphanumeric string.
func RandomString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = CharSet[rnd.Intn(len(CharSet))]
	}
	return string(result)
}

// Ordinalize takes an int and returns the ordinalized version of it.
// Eg. 1 --> 1st, 103 --> 103rd.
func Ordinalize(n int) string {
	m := map[int]string{
		0: "th",
		1: "st",
		2: "nd",
		3: "rd",
		4: "th",
		5: "th",
		6: "th",
		7: "th",
		8: "th",
		9: "th",
	}

	an := int(math.Abs(float64(n)))
	if an < 10 {
		return fmt.Sprintf("%d%s", n, m[an])
	}
	return fmt.Sprintf("%d%s", n, m[an%10])
}

// IsExternalManagedControlPlane returns a bool indicating whether the control plane referenced
// in the passed Unstructured resource is an externally managed control plane such as AKS, EKS, GKE, etc.
func IsExternalManagedControlPlane(controlPlane *unstructured.Unstructured) bool {
	managed, found, err := unstructured.NestedBool(controlPlane.Object, "status", "externalManagedControlPlane")
	if err != nil || !found {
		return false
	}
	return managed
}

// GetMinerIfExists gets a miner from the API server if it exists.
func GetMinerIfExists(ctx context.Context, c client.Client, namespace, name string) (*v1beta1.Miner, error) {
	if c == nil {
		// Being called before k8s is setup as part of control plane VM creation
		return nil, nil
	}

	// Miners are identified by name
	miner := &v1beta1.Miner{}
	err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, miner)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return miner, nil
}

// IsNodeReady returns true if a node is ready.
func IsNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

// GetChainFromMetadata returns the Chain object (if present) using the object metadata.
func GetChainFromMetadata(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*v1beta1.Chain, error) {
	if obj.Labels[v1beta1.ChainNameLabel] == "" {
		return nil, errors.WithStack(ErrNoChain)
	}
	return GetChainByName(ctx, c, obj.Namespace, obj.Labels[v1beta1.ChainNameLabel])
}

// GetOwnerChain returns the Chain object owning the current resource.
func GetOwnerChain(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*v1beta1.Chain, error) {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Kind != "Chain" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == v1beta1.SchemeGroupVersion.Group {
			return GetChainByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetChainByName finds and return a Chain object using the specified params.
func GetChainByName(ctx context.Context, c client.Client, namespace, name string) (*v1beta1.Chain, error) {
	chain := &v1beta1.Chain{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, chain); err != nil {
		return nil, errors.Wrapf(err, "failed to get Chain/%s", name)
	}

	return chain, nil
}

// ObjectKey returns client.ObjectKey for the object.
func ObjectKey(object metav1.Object) client.ObjectKey {
	return client.ObjectKey{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
}

// GetOwnerMiner returns the Miner object owning the current resource.
func GetOwnerMiner(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*v1beta1.Miner, error) {
	for _, ref := range obj.GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, err
		}
		if ref.Kind == "Miner" && gv.Group == v1beta1.SchemeGroupVersion.Group {
			return GetMinerByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetMinerByName finds and return a Miner object using the specified params.
func GetMinerByName(ctx context.Context, c client.Client, namespace, name string) (*v1beta1.Miner, error) {
	m := &v1beta1.Miner{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(ctx, key, m); err != nil {
		return nil, err
	}
	return m, nil
}

// HasOwnerRef returns true if the OwnerReference is already in the slice. It matches based on Group, Kind and Name.
func HasOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) bool {
	return indexOwnerRef(ownerReferences, ref) > -1
}

// EnsureOwnerRef makes sure the slice contains the OwnerReference.
// Note: EnsureOwnerRef will update the version of the OwnerReference fi it exists with a different version. It will also update the UID.
func EnsureOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) []metav1.OwnerReference {
	idx := indexOwnerRef(ownerReferences, ref)
	if idx == -1 {
		return append(ownerReferences, ref)
	}
	ownerReferences[idx] = ref
	return ownerReferences
}

// ReplaceOwnerRef re-parents an object from one OwnerReference to another
// It compares strictly based on UID to avoid reparenting across an intentional deletion: if an object is deleted
// and re-created with the same name and namespace, the only way to tell there was an in-progress deletion
// is by comparing the UIDs.
func ReplaceOwnerRef(ownerReferences []metav1.OwnerReference, source metav1.Object, target metav1.OwnerReference) []metav1.OwnerReference {
	fi := -1
	for index, r := range ownerReferences {
		if r.UID == source.GetUID() {
			fi = index
			ownerReferences[index] = target
			break
		}
	}
	if fi < 0 {
		ownerReferences = append(ownerReferences, target)
	}
	return ownerReferences
}

// RemoveOwnerRef returns the slice of owner references after removing the supplied owner ref.
// Note: RemoveOwnerRef ignores apiVersion and UID. It will remove the passed ownerReference where it matches Name, Group and Kind.
func RemoveOwnerRef(ownerReferences []metav1.OwnerReference, inputRef metav1.OwnerReference) []metav1.OwnerReference {
	if index := indexOwnerRef(ownerReferences, inputRef); index != -1 {
		return append(ownerReferences[:index], ownerReferences[index+1:]...)
	}
	return ownerReferences
}

// indexOwnerRef returns the index of the owner reference in the slice if found, or -1.
func indexOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) int {
	for index, r := range ownerReferences {
		if referSameObject(r, ref) {
			return index
		}
	}
	return -1
}

// IsOwnedByObject returns true if any of the owner references point to the given target.
// It matches the object based on the Group, Kind and Name.
func IsOwnedByObject(obj metav1.Object, target client.Object) bool {
	for _, ref := range obj.GetOwnerReferences() {
		ref := ref
		if refersTo(&ref, target) {
			return true
		}
	}
	return false
}

// IsControlledBy differs from metav1.IsControlledBy. This function matches on Group, Kind and Name. The metav1.IsControlledBy function matches on UID only.
func IsControlledBy(obj metav1.Object, owner client.Object) bool {
	controllerRef := metav1.GetControllerOfNoCopy(obj)
	if controllerRef == nil {
		return false
	}
	return refersTo(controllerRef, owner)
}

// Returns true if a and b point to the same object based on Group, Kind and Name.
func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV.Group == bGV.Group && a.Kind == b.Kind && a.Name == b.Name
}

// Returns true if ref refers to obj based on Group, Kind and Name.
func refersTo(ref *metav1.OwnerReference, obj client.Object) bool {
	refGv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	return refGv.Group == gvk.Group && ref.Kind == gvk.Kind && ref.Name == obj.GetName()
}

// UnstructuredUnmarshalField is a wrapper around json and unstructured objects to decode and copy a specific field
// value into an object.
func UnstructuredUnmarshalField(obj *unstructured.Unstructured, v any, fields ...string) error {
	if obj == nil || obj.Object == nil {
		return errors.Errorf("failed to unmarshal unstructured object: object is nil")
	}

	value, found, err := unstructured.NestedFieldNoCopy(obj.Object, fields...)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve field %q from %q", strings.Join(fields, "."), obj.GroupVersionKind())
	}
	if !found || value == nil {
		return ErrUnstructuredFieldNotFound
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "failed to json-encode field %q value from %q", strings.Join(fields, "."), obj.GroupVersionKind())
	}
	if err := json.Unmarshal(valueBytes, v); err != nil {
		return errors.Wrapf(err, "failed to json-decode field %q value from %q", strings.Join(fields, "."), obj.GroupVersionKind())
	}
	return nil
}

// HasOwner checks if any of the references in the passed list match the given group from apiVersion and one of the given kinds.
func HasOwner(refList []metav1.OwnerReference, apiVersion string, kinds []string) bool {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return false
	}

	kindMap := make(map[string]bool)
	for _, kind := range kinds {
		kindMap[kind] = true
	}

	for _, mr := range refList {
		mrGroupVersion, err := schema.ParseGroupVersion(mr.APIVersion)
		if err != nil {
			return false
		}

		if mrGroupVersion.Group == gv.Group && kindMap[mr.Kind] {
			return true
		}
	}

	return false
}

// GetGVKMetadata retrieves a CustomResourceDefinition metadata from the API server using partial object metadata.
//
// This function is greatly more efficient than GetCRDWithContract and should be preferred in most cases.
func GetGVKMetadata(ctx context.Context, c client.Client, gvk schema.GroupVersionKind) (*metav1.PartialObjectMetadata, error) {
	meta := &metav1.PartialObjectMetadata{}
	meta.SetName(contract.CalculateCRDName(gvk.Group, gvk.Kind))
	meta.SetGroupVersionKind(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
	if err := c.Get(ctx, client.ObjectKeyFromObject(meta), meta); err != nil {
		return meta, errors.Wrap(err, "failed to retrieve metadata from GVK resource")
	}
	return meta, nil
}

// KubeAwareAPIVersions is a sortable slice of kube-like version strings.
//
// Kube-like version strings are starting with a v, followed by a major version,
// optional "alpha" or "beta" strings followed by a minor version (e.g. v1, v2beta1).
// Versions will be sorted based on GA/alpha/beta first and then major and minor
// versions. e.g. v2, v1, v1beta2, v1beta1, v1alpha1.
type KubeAwareAPIVersions []string

func (k KubeAwareAPIVersions) Len() int      { return len(k) }
func (k KubeAwareAPIVersions) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k KubeAwareAPIVersions) Less(i, j int) bool {
	return k8sversion.CompareKubeAwareVersionStrings(k[i], k[j]) < 0
}

// ChainToTypedObjectsMapper returns a mapper function that gets a cluster and lists all objects for the object passed in
// and returns a list of requests.
// Note: This function uses the passed in typed ObjectList and thus with the default client configuration all list calls
// will be cached.
// NB: The objects are required to have `v1beta1.ChainNameLabel` applied.
func ChainToTypedObjectsMapper(c client.Client, ro client.ObjectList, scheme *runtime.Scheme) (handler.MapFunc, error) {
	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return nil, err
	}

	// Note: we create the typed ObjectList once here, so we don't have to use
	// reflection in every execution of the actual event handler.
	obj, err := scheme.New(gvk)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct object of type %s", gvk)
	}
	objectList, ok := obj.(client.ObjectList)
	if !ok {
		return nil, errors.Errorf("expected objject to be a client.ObjectList, is actually %T", obj)
	}

	isNamespaced, err := isAPINamespaced(gvk, c.RESTMapper())
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, o client.Object) []ctrl.Request {
		cluster, ok := o.(*v1beta1.Chain)
		if !ok {
			return nil
		}

		listOpts := []client.ListOption{
			client.MatchingLabels{
				v1beta1.ChainNameLabel: cluster.Name,
			},
		}

		if isNamespaced {
			listOpts = append(listOpts, client.InNamespace(cluster.Namespace))
		}

		objectList = objectList.DeepCopyObject().(client.ObjectList)
		if err := c.List(ctx, objectList, listOpts...); err != nil {
			return nil
		}

		objects, err := meta.ExtractList(objectList)
		if err != nil {
			return nil
		}

		results := []ctrl.Request{}
		for _, obj := range objects {
			// Note: We don't check if the type cast succeeds as all items in an client.ObjectList
			// are client.Objects.
			o := obj.(client.Object)
			results = append(results, ctrl.Request{
				NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()},
			})
		}
		return results
	}, nil
}

// isAPINamespaced detects if a GroupVersionKind is namespaced.
func isAPINamespaced(gk schema.GroupVersionKind, restmapper meta.RESTMapper) (bool, error) {
	restMapping, err := restmapper.RESTMapping(schema.GroupKind{Group: gk.Group, Kind: gk.Kind})
	if err != nil {
		return false, fmt.Errorf("failed to get restmapping: %w", err)
	}

	switch restMapping.Scope.Name() {
	case "":
		return false, errors.New("Scope cannot be identified. Empty scope returned")
	case meta.RESTScopeNameRoot:
		return false, nil
	default:
		return true, nil
	}
}

// ObjectReferenceToUnstructured converts an object reference to an unstructured object.
func ObjectReferenceToUnstructured(in corev1.ObjectReference) *unstructured.Unstructured {
	out := &unstructured.Unstructured{}
	out.SetKind(in.Kind)
	out.SetAPIVersion(in.APIVersion)
	out.SetNamespace(in.Namespace)
	out.SetName(in.Name)
	return out
}

// IsSupportedVersionSkew will return true if a and b are no more than one minor version off from each other.
func IsSupportedVersionSkew(a, b semver.Version) bool {
	if a.Major != b.Major {
		return false
	}
	if a.Minor > b.Minor {
		return a.Minor-b.Minor == 1
	}
	return b.Minor-a.Minor <= 1
}

// LowestNonZeroResult compares two reconciliation results
// and returns the one with lowest requeue time.
func LowestNonZeroResult(i, j ctrl.Result) ctrl.Result {
	switch {
	case i.IsZero():
		return j
	case j.IsZero():
		return i
	case i.Requeue:
		return i
	case j.Requeue:
		return j
	case i.RequeueAfter < j.RequeueAfter:
		return i
	default:
		return j
	}
}

// LowestNonZeroInt32 returns the lowest non-zero value of the two provided values.
func LowestNonZeroInt32(i, j int32) int32 {
	if i == 0 {
		return j
	}
	if j == 0 {
		return i
	}
	if i < j {
		return i
	}
	return j
}

// IsNil returns an error if the passed interface is equal to nil or if it has an interface value of nil.
func IsNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Interface, reflect.UnsafePointer, reflect.Func:
		return reflect.ValueOf(i).IsValid() && reflect.ValueOf(i).IsNil()
	}
	return false
}

// MergeMap merges maps.
// NOTE: In case a key exists in multiple maps, the value of the first map is preserved.
func MergeMap(maps ...map[string]string) map[string]string {
	m := make(map[string]string)
	for i := len(maps) - 1; i >= 0; i-- {
		for k, v := range maps[i] {
			m[k] = v
		}
	}

	// Nil the result if the map is empty, thus avoiding triggering infinite reconcile
	// given that at json level label: {} or annotation: {} is different from no field, which is the
	// corresponding value stored in etcd given that those fields are defined as omitempty.
	if len(m) == 0 {
		return nil
	}
	return m
}
