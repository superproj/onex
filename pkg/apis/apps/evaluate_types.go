// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package apps

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmerrors "github.com/superproj/onex/pkg/errors"
)

const (
	// EvaluateFinalizer is the finalizer used by the Evaluate controller to
	// clean up referenced template resources if necessary when a Evaluate is being deleted.
	EvaluateFinalizer = "evaluate.onex.io/finalizer"
)

// of the actual state of the Miner, and controllers should not use the Miner Phase field
// value when making decisions about what action to take.
//
// Controllers should always look at the actual state of the Miner’s fields to make those decisions.
type EvaluatePhase string

const (
	// EvaluatePhasePending is the first state a Miner is assigned by
	// Cluster API Miner controller after being created.
	EvaluatePhasePending = EvaluatePhase("Pending")

	// EvaluatePhaseProvisioning is the state when the
	// Miner infrastructure is being created.
	EvaluatePhasePrepared = EvaluatePhase("Prepared")

	// EvaluatePhaseProvisioned is the state when its
	// infrastructure has been created and configured.
	EvaluatePhaseEvaluating = EvaluatePhase("Evaluating")

	// EvaluatePhaseRunning is the Miner state when it has
	// become a Kubernetes Node in a Ready state.
	EvaluatePhaseFailed = EvaluatePhase("Failed")

	// EvaluatePhaseDeleting is the Miner state when a delete
	// request has been sent to the API Server,
	// but its infrastructure has not yet been fully deleted.
	EvaluatePhaseSucceeded = EvaluatePhase("Succeeded")

	EvaluatePhaseUnknown = EvaluatePhase("Unknown")
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Evaluate is the Schema for the evaluates API.
type Evaluate struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the evaluate.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec EvaluateSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the evaluate.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status EvaluateStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// EvaluateSpec defines the desired state of Evaluate.
type EvaluateSpec struct {
	// The display name of the evaluate.
	// +optional
	DisplayName string `json:"displayName,omitempty" protobuf:"bytes,1,opt,name=displayName"`

	// Evaluate machine configuration.
	// +optional
	ModelID int64 `json:"modelID,omitempty" protobuf:"bytes,2,opt,name=modelID"`

	// Evaluate machine configuration.
	// +optional
	Provider string `json:"provider,omitempty" protobuf:"bytes,3,opt,name=provider"`

	// +optional
	SampleID int64 `json:"sampleID,omitempty" protobuf:"bytes,4,opt,name=sampleID"`
}

// EvaluateAddresses XXXX
type EvaluateAddresses struct {
	HDFSRoot          string `json:"hdfsRoot,omitempty" protobuf:"bytes,1,opt,name=hdfsRoot"`
	HDFSPtPath        string `json:"hdfsPTPath,omitempty" protobuf:"bytes,2,opt,name=hdfsPTPath"`
	TOSTrainDataPath  string `json:"tosTrainDataPath,omitempty" protobuf:"bytes,3,opt,name=tosTrainDataPath"`
	TOSTestDataPath   string `json:"tosTestDataPath,omitempty" protobuf:"bytes,4,opt,name=tosTestDataPath"`
	TOSTrainDataCount int64  `json:"tosTrainDataCount,omitempty" protobuf:"bytes,5,opt,name=tosTrainDataCount"`
	TOSTestDataConut  int64  `json:"tosTestDataCount,omitempty" protobuf:"bytes,6,opt,name=tosTestDataCount"`
}

// EvaluateStatus defines the observed state of Evaluate.
type EvaluateStatus struct {
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Evaluate's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Evaluates
	// can be added as events to the Evaluate object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *cmerrors.MinerStatusError `json:"failureReason,omitempty" protobuf:"bytes,1,opt,name=failureReason"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Evaluate and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Evaluate's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Evaluates
	// can be added as events to the Evaluate object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty" protobuf:"bytes,2,opt,name=failureMessage"`

	// Addresses is a list of addresses assigned to the evaluate. Queried from kind cluster, if available.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses EvaluateAddresses `json:"addresses,omitempty" protobuf:"bytes,3,opt,name=addresses"`

	// Addresses is a list of addresses assigned to the modelcompare. Queried from kind cluster, if available.
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,4,opt,name=startedAt"`

	// +optional
	EndedAt *metav1.Time `json:"endedAt,omitempty" protobuf:"bytes,5,opt,name=endedAt"`

	// +optional
	ArthurID *string `json:"arthurID,omitempty" protobuf:"bytes,4,opt,name=arthurID"`

	// Phase represents the current phase of evaluate actuation.
	// One of: Failed, Provisioning, Provisioned, Running, Deleting
	// This field is maintained by evaluate controller.
	// +optional
	Phase string `json:"phase,omitempty" protobuf:"bytes,5,opt,name=phase"`

	// Conditions defines the current state of the Evaluate
	// +optional
	Conditions Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EvaluateList is a list of Evaluate objects.
type EvaluateList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of schema objects.
	Items []Evaluate `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// SetTypedPhase sets the Phase field to the string representation of EvaluatePhase.
func (m *EvaluateStatus) SetTypedPhase(p EvaluatePhase) {
	m.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed EvaluatePhase representation as described in `evaluate_phase_types.go`.
func (m *EvaluateStatus) GetTypedPhase() EvaluatePhase {
	switch phase := EvaluatePhase(m.Phase); phase {
	case
		EvaluatePhasePending,
		EvaluatePhasePrepared,
		EvaluatePhaseEvaluating,
		EvaluatePhaseFailed,
		EvaluatePhaseSucceeded:
		return phase
	default:
		return EvaluatePhaseUnknown
	}
}
