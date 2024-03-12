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
	// MinerSetFinalizer is the finalizer used by the MinerSet controller to
	// clean up referenced template resources if necessary when a MinerSet is being deleted.
	MinerSetFinalizer = "minerset.onex.io"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinerSet ensures that a specified number of miners replicas are running at any given time.
type MinerSet struct {
	metav1.TypeMeta `json:",inline"`

	// If the Labels of a MinerSet are empty, they are defaulted to
	// be the same as the Miner(s) that the MinerSet manages.
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the specification of the desired behavior of the MinerSet.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec MinerSetSpec `json:"spec,omitempty"`

	// Status is the most recently observed status of the MinerSet.
	// This data may be out of date by some window of time.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status MinerSetStatus `json:"status,omitempty"`
}

// MinerSetSpec defines the desired state of MinerSet.
type MinerSetSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Selector is a label query over miners that should match the replica count.
	// Label keys and values that must match in order to be controlled by this MinerSet.
	// It must match the miner template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// Template is the object that describes the miner that will be created if
	// insufficient replicas are detected.
	// +optional
	Template MinerTemplateSpec `json:"template,omitempty" protobuf:"bytes,1,opt,name=template"`

	// The display name of the minerset.
	DisplayName string `json:"displayName,omitempty"`

	// DeletePolicy defines the policy used to identify miners to delete when downscaling.
	// Defaults to "Random". Valid values are "Random, "Newest", "Oldest"
	// +kubebuilder:validation:Enum=Random;Newest;Oldest
	// +optional
	DeletePolicy string `json:"deletePolicy,omitempty"`

	// Minimum number of seconds for which a newly created miner should be ready
	// without any of its component crashing, for it to be considered available.
	// Defaults to 0 (miner will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,4,opt,name=minReadySeconds"`

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will
	// not be estimated during the time a deployment is paused. Defaults to 600s.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`
}

// MinerTemplateSpec describes the data needed to create a Miner from a template.
type MinerTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the miner.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec MinerSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// MinerSetStatus represents the current status of a MinerSet.
type MinerSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas" protobuf:"varint,1,opt,name=replicas"`

	// The number of miners that have labels matching the labels of the miner template of the minerset.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty" protobuf:"varint,2,opt,name=fullyLabeledReplicas"`

	// readyReplicas is the number of miners targeted by this MinerSet with a Ready Condition.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`

	// The number of available replicas (ready for at least minReadySeconds) for this minerset.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty" protobuf:"varint,4,opt,name=availableReplicas"`

	// ObservedGeneration reflects the generation of the most recently observed MinerSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,5,opt,name=observedGeneration"`

	// In the event that there is a terminal problem reconciling the
	// replicas, both FailureReason and FailureMessage will be set. FailureReason
	// will be populated with a succinct value suitable for miner
	// interpretation, while FailureMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the MinerTemplate's spec or the configuration of
	// the miner controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the miner controller, or the
	// responsible miner controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Miners
	// can be added as events to the MinerSet object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *cmerrors.MinerSetStatusError `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the MinerSet and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the MinerSet's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of MinerSets
	// can be added as events to the MinerSet object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Represents the latest available observations of a miner set's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,6,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinerSetList contains a list of MinerSet.
type MinerSetList struct {
	metav1.TypeMeta `             json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `             json:"metadata,omitempty"`

	// List of MinerSets.
	Items []MinerSet `json:"items"`
}

type MinerSetDeletePolicy string

const (
	// RandomMinerSetDeletePolicy prioritizes both Miners that have the annotation
	// "apps.onex.io/delete-miner=yes" and Miners that are unhealthy
	// (Status.ErrorReason or Status.ErrorMessage are set to a non-empty value).
	// Finally, it picks Miners at random to delete.
	RandomMinerSetDeletePolicy MinerSetDeletePolicy = "Random"
	// NewestMinerSetDeletePolicy prioritizes both Miners that have the annotation
	// "apps.onex.io/delete-miner=yes" and Miners that are unhealthy
	// (Status.ErrorReason or Status.ErrorMessage are set to a non-empty value).
	// It then prioritizes the newest Miners for deletion based on the Miner's CreationTimestamp.
	NewestMinerSetDeletePolicy MinerSetDeletePolicy = "Newest"
	// OldestMinerSetDeletePolicy prioritizes both Miners that have the annotation
	// "apps.onex.io/delete-miner=yes" and Miners that are unhealthy
	// (Status.ErrorReason or Status.ErrorMessage are set to a non-empty value).
	// It then prioritizes the oldest Miners for deletion based on the Miner's CreationTimestamp.
	OldestMinerSetDeletePolicy MinerSetDeletePolicy = "Oldest"
)
