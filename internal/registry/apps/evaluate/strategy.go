// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:gocritic
package evaluate

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

// evaluateStrategy implements behavior for Evaluate objects.
type evaluateStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Evaluate
// objects via the REST API.
var Strategy = evaluateStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (evaluateStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for evaluates.
func (evaluateStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (evaluateStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (evaluateStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	evaluate := obj.(*apps.Evaluate)
	evaluate.Status = apps.EvaluateStatus{}
	evaluate.Generation = 1
	evaluate.Status.StartedAt = ptr.To(metav1.Now())
	evaluate.Status.Phase = string(apps.EvaluatePhasePending)

	dropEvaluateDisabledFields(evaluate, nil)

	// Be explicit that users cannot create pre-provisioned evaluates.
	evaluate.Status.Conditions = []apps.Condition{}
}

// Validate validates a new evaluate.
func (evaluateStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	evaluate := obj.(*apps.Evaluate)
	return validation.ValidateEvaluate(evaluate)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (evaluateStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (evaluateStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for evaluates.
func (evaluateStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (evaluateStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newEvaluate := obj.(*apps.Evaluate)
	oldEvaluate := old.(*apps.Evaluate)
	// Update is not allowed to set status
	newEvaluate.Status = oldEvaluate.Status

	dropEvaluateDisabledFields(newEvaluate, oldEvaluate)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldEvaluate.Spec, newEvaluate.Spec) {
		newEvaluate.Generation = oldEvaluate.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (evaluateStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateEvaluateUpdate(obj.(*apps.Evaluate), old.(*apps.Evaluate))
}

// WarningsOnUpdate returns warnings for the given update.
func (evaluateStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (evaluateStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// Storage strategy for the Status subresource.
type evaluateStatusStrategy struct {
	evaluateStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = evaluateStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (evaluateStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (evaluateStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newEvaluate := obj.(*apps.Evaluate)
	oldEvaluate := old.(*apps.Evaluate)

	// Updating /status should not modify spec
	newEvaluate.Spec = oldEvaluate.Spec
	newEvaluate.DeletionTimestamp = nil

	// don't allow the evaluates/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newEvaluate.OwnerReferences = oldEvaluate.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (evaluateStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateEvaluateStatusUpdate(obj.(*apps.Evaluate), old.(*apps.Evaluate))
}

// WarningsOnUpdate returns warnings for the given update.
func (evaluateStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (evaluateStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that can be used for filter selection.
func ToSelectableFields(obj *apps.Evaluate) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	c, ok := obj.(*apps.Evaluate)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a evaluate")
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
	return obj.(*apps.Evaluate).ObjectMeta.Name
}

func dropEvaluateDisabledFields(evaluate *apps.Evaluate, oldEvaluate *apps.Evaluate) {
}
