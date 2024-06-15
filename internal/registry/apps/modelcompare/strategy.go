// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:gocritic
package modelcompare

import (
	"context"
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	"github.com/superproj/onex/pkg/apis/apps"
	"github.com/superproj/onex/pkg/apis/apps/validation"
)

// modelcompareStrategy implements behavior for ModelCompare objects.
type modelcompareStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ModelCompare
// objects via the REST API.
var Strategy = modelcompareStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (modelcompareStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for modelcompares.
func (modelcompareStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (modelcompareStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (modelcompareStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	modelcompare := obj.(*apps.ModelCompare)
	modelcompare.Status = apps.ModelCompareStatus{}
	modelcompare.Generation = 1
	modelcompare.Status.StartedAt = ptr.To(metav1.Now())
	modelcompare.Status.Phase = string(apps.ModelComparePhasePending)

	dropModelCompareDisabledFields(modelcompare, nil)

	// Be explicit that users cannot create pre-provisioned modelcompares.
	modelcompare.Status.Conditions = []apps.Condition{}
}

// Validate validates a new modelcompare.
func (modelcompareStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	modelcompare := obj.(*apps.ModelCompare)
	return validation.ValidateModelCompare(modelcompare)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (modelcompareStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (modelcompareStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for modelcompares.
func (modelcompareStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (modelcompareStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newModelCompare := obj.(*apps.ModelCompare)
	oldModelCompare := old.(*apps.ModelCompare)
	// Update is not allowed to set status
	newModelCompare.Status = oldModelCompare.Status

	dropModelCompareDisabledFields(newModelCompare, oldModelCompare)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldModelCompare.Spec, newModelCompare.Spec) {
		newModelCompare.Generation = oldModelCompare.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (modelcompareStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateModelCompareUpdate(obj.(*apps.ModelCompare), old.(*apps.ModelCompare))
}

// WarningsOnUpdate returns warnings for the given update.
func (modelcompareStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (modelcompareStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// Storage strategy for the Status subresource.
type modelcompareStatusStrategy struct {
	modelcompareStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = modelcompareStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (modelcompareStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (modelcompareStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newModelCompare := obj.(*apps.ModelCompare)
	oldModelCompare := old.(*apps.ModelCompare)

	// Updating /status should not modify spec
	newModelCompare.Spec = oldModelCompare.Spec
	newModelCompare.DeletionTimestamp = nil

	// don't allow the modelcompares/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newModelCompare.OwnerReferences = oldModelCompare.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (modelcompareStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateModelCompareStatusUpdate(obj.(*apps.ModelCompare), old.(*apps.ModelCompare))
}

// WarningsOnUpdate returns warnings for the given update.
func (modelcompareStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (modelcompareStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that can be used for filter selection.
func ToSelectableFields(obj *apps.ModelCompare) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	c, ok := obj.(*apps.ModelCompare)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a modelcompare")
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
	return obj.(*apps.ModelCompare).ObjectMeta.Name
}

func dropModelCompareDisabledFields(modelcompare *apps.ModelCompare, oldModelCompare *apps.ModelCompare) {
}
