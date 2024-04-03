// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ChargeRequestFinalizer is the finalizer used by the ChargeRequest controller to
	// clean up referenced template resources if necessary when a ChargeRequest is being deleted.
	ChargeRequestFinalizer = "chargerequest.onex.io/finalizer"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChargeRequest is the Schema for the chargerequests API.
type ChargeRequest struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the chargerequest.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec ChargeRequestSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the most recently observed status of the ChargeRequest.
	// This data may be out of date by some window of time.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status ChargeRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ChargeRequestSpec defines the desired state of ChargeRequest.
type ChargeRequestSpec struct {
	// +optional
	From     string `json:"from,omitempty" protobuf:"bytes,1,opt,name=from"`
	Password string `json:"password,omitempty" protobuf:"bytes,2,opt,name=password"`
}

// ChargeRequestStatus defines the observed state of ChargeRequest.
type ChargeRequestStatus struct {
	// +optional
	Conditions Conditions `json:"conditions,omitempty" protobuf:"bytes,1,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChargeRequestList is a list of ChargeRequest objects.
type ChargeRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of schema objects.
	Items []ChargeRequest `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GetConditions returns the set of conditions for this object.
func (cr *ChargeRequest) GetConditions() Conditions {
	return cr.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (cr *ChargeRequest) SetConditions(conditions Conditions) {
	cr.Status.Conditions = conditions
}
