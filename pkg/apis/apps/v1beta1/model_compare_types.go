// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// ModelCompareFinalizer is the finalizer used by the ModelCompare controller to
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmerrors "github.com/superproj/onex/pkg/errors"
)

const (
	// ModelCompareFinalizer is the finalizer used by the ModelCompare controller to
	// clean up referenced template resources if necessary when a ModelCompare is being deleted.
	ModelCompareFinalizer = "modelcompare.onex.io/finalizer"
)

// of the actual state of the Miner, and controllers should not use the Miner Phase field
// value when making decisions about what action to take.
//
// Controllers should always look at the actual state of the Miner’s fields to make those decisions.
type ModelComparePhase string

const (
	// ModelComparePhasePending is the first state a Miner is assigned by
	// Cluster API Miner controller after being created.
	ModelComparePhasePending = ModelComparePhase("Pending")

	// ModelComparePhaseProvisioning is the state when the
	// Miner infrastructure is being created.
	ModelComparePhasePrepared = ModelComparePhase("Prepared")

	// ModelComparePhaseProvisioned is the state when its
	// infrastructure has been created and configured.
	ModelComparePhaseEvaluating = ModelComparePhase("Evaluating")

	// ModelComparePhaseRunning is the Miner state when it has
	// become a Kubernetes Node in a Ready state.
	ModelComparePhaseFailed = ModelComparePhase("Failed")

	// ModelComparePhaseDeleting is the Miner state when a delete
	// request has been sent to the API Server,
	// but its infrastructure has not yet been fully deleted.
	ModelComparePhaseSucceeded = ModelComparePhase("Succeeded")

	ModelComparePhaseUnknown = ModelComparePhase("Unknown")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ModelCompare is the Schema for the modelcompares API.
type ModelCompare struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the modelcompare.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec ModelCompareSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the modelcompare.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status ModelCompareStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ModelCompareSpec defines the desired state of ModelCompare.
type ModelCompareSpec struct {
	// Selector is a label query over miners that should match the replica count.
	// Label keys and values that must match in order to be controlled by this MinerSet.
	// It must match the miner template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`

	// Template is the object that describes the miner that will be created if
	// insufficient replicas are detected.
	// +optional
	Template EvaluateTemplateSpec `json:"template,omitempty" protobuf:"bytes,2,opt,name=template"`

	// The display name of the modelcompare.
	// +optional
	DisplayName string `json:"displayName,omitempty" protobuf:"bytes,3,opt,name=displayName"`

	// ModelCompare machine configuration.
	// +optional
	ModelIDs []int64 `json:"modelIDs,omitempty" protobuf:"bytes,4,opt,name=modelIDs"`
}

// EvaluateTemplateSpec describes the data needed to create a Evaluate from a template.
type EvaluateTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the miner.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec EvaluateSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ModelCompareStatus defines the observed state of ModelCompare.
type ModelCompareStatus struct {
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the ModelCompare's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of ModelCompares
	// can be added as events to the ModelCompare object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *cmerrors.MinerStatusError `json:"failureReason,omitempty" protobuf:"bytes,1,opt,name=failureReason"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the ModelCompare and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the ModelCompare's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of ModelCompares
	// can be added as events to the ModelCompare object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty" protobuf:"bytes,2,opt,name=failureMessage"`

	// Addresses is a list of addresses assigned to the modelcompare. Queried from kind cluster, if available.
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,3,opt,name=startedAt"`

	// +optional
	EndedAt *metav1.Time `json:"endedAt,omitempty" protobuf:"bytes,4,opt,name=endedAt"`

	// Phase represents the current phase of modelcompare actuation.
	// One of: Failed, Provisioning, Provisioned, Running, Deleting
	// This field is maintained by modelcompare controller.
	// +optional
	Phase string `json:"phase,omitempty" protobuf:"bytes,5,opt,name=phase"`

	// Conditions defines the current state of the Evaluate
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,6,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ModelCompareList is a list of ModelCompare objects.
type ModelCompareList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of schema objects.
	Items []ModelCompare `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GetConditions returns the set of conditions for the MinerSet.
func (mc *ModelCompare) GetConditions() Conditions {
	return mc.Status.Conditions
}

// SetConditions updates the set of conditions on the MinerSet.
func (mc *ModelCompare) SetConditions(conditions Conditions) {
	mc.Status.Conditions = conditions
}

// SetTypedPhase sets the Phase field to the string representation of ModelComparePhase.
func (m *ModelCompareStatus) SetTypedPhase(p ModelComparePhase) {
	m.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed ModelComparePhase representation as described in `modelcompare_phase_types.go`.
func (m *ModelCompareStatus) GetTypedPhase() ModelComparePhase {
	switch phase := ModelComparePhase(m.Phase); phase {
	case
		ModelComparePhasePending,
		ModelComparePhasePrepared,
		ModelComparePhaseEvaluating,
		ModelComparePhaseFailed,
		ModelComparePhaseSucceeded:
		return phase
	default:
		return ModelComparePhaseUnknown
	}
}
