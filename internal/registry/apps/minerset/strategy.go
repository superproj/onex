// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:gocritic
package minerset

import (
	"context"
	"fmt"
	"strconv"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
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

// minerSetStrategy implements behavior for MinerSet objects.
type minerSetStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating MinerSet
// objects via the REST API.
var Strategy = minerSetStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (minerSetStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for minersets.
func (minerSetStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (minerSetStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (minerSetStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	minerset := obj.(*apps.MinerSet)
	minerset.Status = apps.MinerSetStatus{}
	minerset.Generation = 1

	dropMinerSetDisabledFields(minerset, nil)

	// Be explicit that users cannot create pre-provisioned minersets.
	minerset.Status.Conditions = []apps.Condition{}
}

// Validate validates a new minerset.
func (minerSetStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	minerset := obj.(*apps.MinerSet)
	return validation.ValidateMinerSet(minerset)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (minerSetStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	newMinerSet := obj.(*apps.MinerSet)
	var warnings []string
	if msgs := utilvalidation.IsDNS1123Label(newMinerSet.Name); len(msgs) != 0 {
		warnings = append(warnings, fmt.Sprintf("metadata.name: this is used in Pod names and hostnames, which can result in surprising behavior;a DNS label is recommended: %v", msgs))
	}
	return warnings
}

// Canonicalize normalizes the object after validation.
func (minerSetStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for minersets.
func (minerSetStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (minerSetStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMinerSet := obj.(*apps.MinerSet)
	oldMinerSet := old.(*apps.MinerSet)
	// Update is not allowed to set status
	newMinerSet.Status = oldMinerSet.Status

	dropMinerSetDisabledFields(newMinerSet, oldMinerSet)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldMinerSet.Spec, newMinerSet.Spec) {
		newMinerSet.Generation = oldMinerSet.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (minerSetStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateMinerSetUpdate(obj.(*apps.MinerSet), old.(*apps.MinerSet))
}

// WarningsOnUpdate returns warnings for the given update.
func (minerSetStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (minerSetStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// Storage strategy for the Status subresource.
type minerSetStatusStrategy struct {
	minerSetStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = minerSetStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (minerSetStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"apps.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
			fieldpath.MakePathOrDie("metadata", "labels"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (minerSetStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMinerSet := obj.(*apps.MinerSet)
	oldMinerSet := old.(*apps.MinerSet)

	// Updating /status should not modify spec
	newMinerSet.Spec = oldMinerSet.Spec
	newMinerSet.Labels = oldMinerSet.Labels
	newMinerSet.DeletionTimestamp = nil

	// don't allow the minersets/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newMinerSet.OwnerReferences = oldMinerSet.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (minerSetStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateMinerSetStatusUpdate(obj.(*apps.MinerSet), old.(*apps.MinerSet))
}

// WarningsOnUpdate returns warnings for the given update.
func (minerSetStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (minerSetStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that represents the object.
func ToSelectableFields(obj *apps.MinerSet) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
	minerSetSpecificFieldsSet := fields.Set{
		"status.replicas": strconv.Itoa(int(obj.Status.Replicas)),
		// "spec.type":    obj.Spec.Type, TODO ?
		// "spec.address": obj.Spec.Address,
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, minerSetSpecificFieldsSet)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	ms, ok := obj.(*apps.MinerSet)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a minerset")
	}
	return labels.Set(ms.ObjectMeta.Labels), ToSelectableFields(ms), nil
}

// Matcher is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func Matcher(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// dropMinerSetDisabledFields drops fields that are not used if their associated feature gates
// are not enabled.
// The typical pattern is:
//
//	if !utilfeature.DefaultFeatureGate.Enabled(features.MyFeature) && !myFeatureInUse(oldSvc) {
//	    newSvc.Spec.MyFeature = nil
//	}
func dropMinerSetDisabledFields(newMinerSet *apps.MinerSet, oldMinerSet *apps.MinerSet) {
}
