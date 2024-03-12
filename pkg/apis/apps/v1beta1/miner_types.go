// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmerrors "github.com/superproj/onex/pkg/errors"
)

const (
	// MinerFinalizer is the finalizer used by the Miner controller to
	// clean up referenced template resources if necessary when a Miner is being deleted.
	MinerFinalizer = "miner.onex.io/finalizer"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Miner is the Schema for the miners API.
type Miner struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the miner.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec MinerSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the miner.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status MinerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MinerSpec defines the desired state of Miner.
type MinerSpec struct {
	// ObjectMeta will autopopulate the Pod created. Use this to
	// indicate what labels, annotations, name prefix, etc., should be used
	// when creating the Pod.
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// The display name of the miner.
	// +optional
	DisplayName string `json:"displayName,omitempty" protobuf:"bytes,2,opt,name=displayName"`

	// Miner machine configuration.
	// +optional
	MinerType string `json:"minerType,omitempty" protobuf:"bytes,3,opt,name=minerType"`

	// +optional
	ChainName string `json:"chainName,omitempty" protobuf:"bytes,5,opt,name=chainName"`

	// Restart policy for the miner.
	// One of Always, OnFailure, Never.
	// Default to Always.
	// +optional
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,6,opt,name=restartPolicy,casttype=RestartPolicy"`

	// PodDeletionTimeout defines how long the controller will attempt to delete the Pod that the Machine
	// hosts after the Machine is marked for deletion. A duration of 0 will retry deletion indefinitely.
	// Defaults to 10 seconds.
	// +optional
	PodDeletionTimeout *metav1.Duration `json:"podDeletionTimeout,omitempty" protobuf:"bytes,7,opt,name=podDeletionTimeout"`
}

// MinerStatus defines the observed state of Miner.
type MinerStatus struct {
	// PodRef will point to the corresponding Pod if it exists.
	// +optional
	PodRef *corev1.ObjectReference `json:"podRef,omitempty" protobuf:"bytes,1,opt,name=podRef,casttype=k8s.io/api/core/v1.ObjectReference"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty" protobuf:"bytes,2,opt,name=lastUpdated"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Miner and will contain a succinct value suitable
	// for miner interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Miner's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Miners
	// can be added as events to the Miner object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *cmerrors.MinerStatusError `json:"failureReason,omitempty" protobuf:"bytes,3,opt,name=failureReason"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Miner and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Miner's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Miners
	// can be added as events to the Miner object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty" protobuf:"bytes,4,opt,name=failureMessage"`

	// Addresses is a list of addresses assigned to the miner. Queried from kind cluster, if available.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses MinerAddresses `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,5,rep,name=addresses"`

	// Phase represents the current phase of miner actuation.
	// One of: Failed, Provisioning, Provisioned, Running, Deleting
	// This field is maintained by miner controller.
	// +optional
	Phase string `json:"phase,omitempty" protobuf:"bytes,6,opt,name=phase"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,7,opt,name=observedGeneration"`

	// Conditions defines the current state of the Miner
	// +optional
	Conditions Conditions `json:"conditions,omitempty" protobuf:"bytes,8,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinerList is a list of Miner objects.
type MinerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of schema objects.
	Items []Miner `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
type LocalObjectReference struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// TODO: Add other useful fields. apiVersion, kind, uid?
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// PodInfo is a set of ids/uuids to uniquely identify the pod.
type PodInfo struct {
	// The Operating System reported by the pod
	OperatingSystem string `json:"operatingSystem" protobuf:"bytes,9,opt,name=operatingSystem"`
	// The Architecture reported by the  pod
	Architecture string `json:"architecture" protobuf:"bytes,10,opt,name=architecture"`
}

// GetConditions returns the set of conditions for this object.
func (m *Miner) GetConditions() Conditions {
	return m.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (m *Miner) SetConditions(conditions Conditions) {
	m.Status.Conditions = conditions
}

// SetTypedPhase sets the Phase field to the string representation of MinerPhase.
func (m *MinerStatus) SetTypedPhase(p MinerPhase) {
	m.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed MinerPhase representation as described in `miner_phase_types.go`.
func (m *MinerStatus) GetTypedPhase() MinerPhase {
	switch phase := MinerPhase(m.Phase); phase {
	case
		MinerPhasePending,
		MinerPhaseProvisioning,
		MinerPhaseRunning,
		MinerPhaseDeleting,
		MinerPhaseFailed:
		return phase
	default:
		return MinerPhaseUnknown
	}
}
