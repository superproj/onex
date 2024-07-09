// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:gocritic
package chain

import (
	"context"
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	"github.com/superproj/onex/pkg/apis/apps"
	"github.com/superproj/onex/pkg/apis/apps/validation"
)

// chainStrategy implements behavior for Chain objects.
type chainStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Chain
// objects via the REST API.
var Strategy = chainStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (chainStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for chains.
func (chainStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (chainStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (chainStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	chain := obj.(*apps.Chain)
	chain.Status = apps.ChainStatus{}
	chain.Generation = 1

	dropChainDisabledFields(chain, nil)

	// Be explicit that users cannot create pre-provisioned chains.
	chain.Status.Conditions = []apps.Condition{}
}

// Validate validates a new chain.
func (chainStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	chain := obj.(*apps.Chain)
	return validation.ValidateChain(chain)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (chainStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string { return nil }

// Canonicalize normalizes the object after validation.
func (chainStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for chains.
func (chainStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (chainStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newChain := obj.(*apps.Chain)
	oldChain := old.(*apps.Chain)
	// Update is not allowed to set status
	newChain.Status = oldChain.Status

	dropChainDisabledFields(newChain, oldChain)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldChain.Spec, newChain.Spec) {
		newChain.Generation = oldChain.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (chainStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateChainUpdate(obj.(*apps.Chain), old.(*apps.Chain))
}

// WarningsOnUpdate returns warnings for the given update.
func (chainStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (chainStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// Storage strategy for the Status subresource.
type chainStatusStrategy struct {
	chainStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = chainStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (chainStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (chainStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newChain := obj.(*apps.Chain)
	oldChain := old.(*apps.Chain)

	// Updating /status should not modify spec
	newChain.Spec = oldChain.Spec
	newChain.DeletionTimestamp = nil

	// don't allow the chains/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newChain.OwnerReferences = oldChain.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (chainStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateChainStatusUpdate(obj.(*apps.Chain), old.(*apps.Chain))
}

// WarningsOnUpdate returns warnings for the given update.
func (chainStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (chainStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that can be used for filter selection.
func ToSelectableFields(obj *apps.Chain) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	c, ok := obj.(*apps.Chain)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a chain")
	}
	return labels.Set(c.Labels), ToSelectableFields(c), nil
}

// Matcher is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func Matcher(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{"metadata.name"},
	}
}

// NameTriggerFunc returns value metadata.namespace of given object.
func NameTriggerFunc(obj runtime.Object) string {
	return obj.(*apps.Chain).ObjectMeta.Name
}

func dropChainDisabledFields(chain *apps.Chain, oldChain *apps.Chain) {
}
