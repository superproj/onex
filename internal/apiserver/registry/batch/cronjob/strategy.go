// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

//nolint:gocritic
package cronjob

import (
	"context"
	"fmt"

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
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"

	"github.com/onexstack/onex/pkg/apis/batch"
	"github.com/onexstack/onex/pkg/apis/batch/validation"
)

// cronJobStrategy implements behavior for CronJob objects.
type cronJobStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating CronJob
// objects via the REST API.
var Strategy = cronJobStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (cronJobStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for cronjobs.
func (cronJobStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (cronJobStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"batch.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (cronJobStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	cronjob := obj.(*batch.CronJob)
	cronjob.Status = batch.CronJobStatus{}
	cronjob.Generation = 1

	dropCronJobDisabledFields(cronjob, nil)

	// Be explicit that users cannot create pre-provisioned cronjobs.
	// cronjob.Status.Conditions = []batch.Condition{}
}

// Validate validates a new cronjob.
func (cronJobStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	cronjob := obj.(*batch.CronJob)
	return validation.ValidateCronJobCreate(cronjob)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (cronJobStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	newCronJob := obj.(*batch.CronJob)
	var warnings []string
	if msgs := utilvalidation.IsDNS1123Label(newCronJob.Name); len(msgs) != 0 {
		warnings = append(warnings, fmt.Sprintf("metadata.name: this is used in Pod names and hostnames, which can result in surprising behavior;a DNS label is recommended: %v", msgs))
	}
	return warnings
}

// Canonicalize normalizes the object after validation.
func (cronJobStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for cronjobs.
func (cronJobStrategy) AllowCreateOnUpdate(ctx context.Context) bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (cronJobStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newCronJob := obj.(*batch.CronJob)
	oldCronJob := old.(*batch.CronJob)
	// Update is not allowed to set status
	newCronJob.Status = oldCronJob.Status

	dropCronJobDisabledFields(newCronJob, oldCronJob)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldCronJob.Spec, newCronJob.Spec) {
		newCronJob.Generation = oldCronJob.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (cronJobStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateCronJobUpdate(obj.(*batch.CronJob), old.(*batch.CronJob))
}

// WarningsOnUpdate returns warnings for the given update.
func (cronJobStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (cronJobStrategy) AllowUnconditionalUpdate(ctx context.Context) bool {
	return true
}

// Storage strategy for the Status subresource.
type cronJobStatusStrategy struct {
	cronJobStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = cronJobStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (cronJobStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"batch.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
			fieldpath.MakePathOrDie("metadata", "labels"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (cronJobStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newCronJob := obj.(*batch.CronJob)
	oldCronJob := old.(*batch.CronJob)

	// Updating /status should not modify spec
	newCronJob.Spec = oldCronJob.Spec
	newCronJob.Labels = oldCronJob.Labels
	newCronJob.DeletionTimestamp = nil

	// don't allow the cronjobs/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newCronJob.OwnerReferences = oldCronJob.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (cronJobStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

// WarningsOnUpdate returns warnings for the given update.
func (cronJobStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (cronJobStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that represents the object.
func ToSelectableFields(obj *batch.CronJob) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
	cronJobSpecificFieldsSet := fields.Set{
		// "spec.type":    obj.Spec.Type, TODO ?
		// "spec.address": obj.Spec.Address,
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, cronJobSpecificFieldsSet)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	cj, ok := obj.(*batch.CronJob)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a cronjob")
	}
	return labels.Set(cj.ObjectMeta.Labels), ToSelectableFields(cj), nil
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

// dropCronJobDisabledFields drops fields that are not used if their associated feature gates
// are not enabled.
// The typical pattern is:
//
//	if !utilfeature.DefaultFeatureGate.Enabled(features.MyFeature) && !myFeatureInUse(oldSvc) {
//	    newSvc.Spec.MyFeature = nil
//	}
func dropCronJobDisabledFields(newCronJob *batch.CronJob, oldCronJob *batch.CronJob) {
}
