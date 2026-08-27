// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

//nolint:gocritic
package job

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
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"

	"github.com/onexstack/onex/pkg/apis/batch"
	"github.com/onexstack/onex/pkg/apis/batch/validation"
)

// jobStrategy implements behavior for Job objects.
type jobStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Job
// objects via the REST API.
var Strategy = jobStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var (
	// Make sure we correctly implement the interface.
	_ = rest.GarbageCollectionDeleteStrategy(Strategy)
	// Strategy should implement rest.RESTCreateStrategy.
	_ rest.RESTCreateStrategy = Strategy
	// Strategy should implement rest.RESTUpdateStrategy.
	_ rest.RESTUpdateStrategy = Strategy
)

// DefaultGarbageCollectionPolicy returns DeleteDependents for all currently served versions.
func (jobStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.DeleteDependents
}

// NamespaceScoped is true for jobs.
func (jobStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (jobStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"batch.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (jobStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	job := obj.(*batch.Job)
	job.Status = batch.JobStatus{}
	job.Generation = 1

	dropJobDisabledFields(job, nil)

	// Be explicit that users cannot create pre-provisioned jobs.
	// job.Status.Conditions = []batch.Condition{}
}

// Validate validates a new job.
func (jobStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	job := obj.(*batch.Job)
	return validation.ValidateJob(job)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (jobStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string { return nil }

// Canonicalize normalizes the object after validation.
func (jobStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for jobs.
func (jobStrategy) AllowCreateOnUpdate(ctx context.Context) bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (jobStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newJob := obj.(*batch.Job)
	oldJob := old.(*batch.Job)
	// Update is not allowed to set status
	newJob.Status = oldJob.Status

	dropJobDisabledFields(newJob, oldJob)

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object.
	// See metav1.ObjectMeta description for more information on Generation.
	if !apiequality.Semantic.DeepEqual(oldJob.Spec, newJob.Spec) {
		newJob.Generation = oldJob.Generation + 1
	}
}

// ValidateUpdate is the default update validation for an end user.
func (jobStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateJobUpdate(obj.(*batch.Job), old.(*batch.Job))
}

// WarningsOnUpdate returns warnings for the given update.
func (jobStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (jobStrategy) AllowUnconditionalUpdate(ctx context.Context) bool {
	return true
}

// Storage strategy for the Status subresource.
type jobStatusStrategy struct {
	jobStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = jobStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (jobStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"batch.onex.io/v1beta1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
			fieldpath.MakePathOrDie("status", "conditions"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status.
func (jobStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newJob := obj.(*batch.Job)
	oldJob := old.(*batch.Job)

	// Updating /status should not modify spec
	newJob.Spec = oldJob.Spec
	newJob.DeletionTimestamp = nil

	// don't allow the jobs/status endpoint to touch owner references since old kubelets corrupt them in a way
	// that breaks garbage collection
	newJob.OwnerReferences = oldJob.OwnerReferences
}

// ValidateUpdate is the default update validation for an end user updating status.
func (jobStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateJobStatusUpdate(obj.(*batch.Job), old.(*batch.Job))
}

// WarningsOnUpdate returns warnings for the given update.
func (jobStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (jobStatusStrategy) Canonicalize(obj runtime.Object) {
}

// ToSelectableFields returns a field set that can be used for filter selection.
func ToSelectableFields(obj *batch.Job) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
	jobSpecificFieldsSet := fields.Set{
		//"spec.jobType": obj.Spec.JobType,
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, jobSpecificFieldsSet)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	m, ok := obj.(*batch.Job)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a job")
	}
	return labels.Set(m.Labels), ToSelectableFields(m), nil
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

// dropJobDisabledFields drops fields that are not used if their associated feature gates
// are not enabled.
// The typical pattern is:
//
//	if !utilfeature.DefaultFeatureGate.Enabled(features.MyFeature) && !myFeatureInUse(oldSvc) {
//	    newSvc.Spec.MyFeature = nil
//	}
func dropJobDisabledFields(newJob *batch.Job, oldJob *batch.Job) {
}
