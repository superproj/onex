// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

const (
	// ModelCompareNameLabel is the label set on evaluates linked to a modelcompare.
	ModelCompareNameLabel = "apps.onex.io/modelcompare-name"

	// ChainNameLabel is the label set on miners linked to a chain.
	ChainNameLabel = "apps.onex.io/chain-name"

	// MinerSetNameLabel is the label set on miners linked to a minerset.
	MinerSetNameLabel = "apps.onex.io/minerset-name"

	// MinerDeploymentNameLabel is the label set on miners if they're controlled by MinerDeployment.
	MinerDeploymentNameLabel = "apps.onex.io/deployment-name"

	// MinerNamespaceAnnotation is the annotation set on pods identifying the namespace of the miner the pod belongs to.
	MinerNamespaceAnnotation = "apps.onex.io/miner-namespace"

	// MinerAnnotation is the annotation set on pods identifying the miner the pod belongs to.
	MinerAnnotation = "apps.onex.io/miner"

	// OwnerKindAnnotation is the annotation set on pods identifying the owner kind.
	OwnerKindAnnotation = "apps.onex.io/owner-kind"

	// OwnerNameAnnotation is the annotation set on pods identifying the owner name.
	OwnerNameAnnotation = "apps.onex.io/owner-name"

	// DisableMinerCreate is an annotation that can be used to signal a MinerSet to stop creating new miners.
	// It is utilized in the OnDelete MinerSetStrategy to allow the MinerSet controller to scale down
	// older MinerSets when Miners are deleted and add the new replicas to the latest MinerSet.
	DisableMinerCreateAnnotation = "apps.onex.io/disable-miner-create"

	// DeleteMinerAnnotation marks control plane and worker nodes that will be given priority for deletion
	// when KCP or a minerset scales down. This annotation is given top priority on all delete policies.
	DeleteMinerAnnotation = "apps.onex.io/delete-miner"

	// WatchLabel is a label othat can be applied to any OneX API object.
	//
	// Controllers which allow for selective reconciliation may check this label and proceed
	// with reconciliation of the object only if this label and a configured value is present.
	WatchLabel = "apps.onex.io/watch-filter"

	// PausedAnnotation is an annotation that can be applied to any OneX API
	// object to prevent a controller from processing a resource.
	//
	// Controllers working with OneX objects must check the existence of this annotation
	// on the reconciled object.
	PausedAnnotation = "apps.onex.io/paused"

	// MinerSkipRemediationAnnotation is the annotation used to mark the miners
	// that should not be considered for remediation by MinerHealthCheck reconciler.
	MinerSkipRemediationAnnotation = "apps.onex.io/skip-remediation"
)

// MinerAddressType describes a valid MinerAddress type.
type MinerAddressType string

// Define the MinerAddressType constants.
const (
	MinerHostName    MinerAddressType = "Hostname"
	MinerExternalIP  MinerAddressType = "ExternalIP"
	MinerInternalIP  MinerAddressType = "InternalIP"
	MinerExternalDNS MinerAddressType = "ExternalDNS"
	MinerInternalDNS MinerAddressType = "InternalDNS"
)

// MinerAddress contains information for the miner's address.
type MinerAddress struct {
	// Miner address type, one of Hostname, ExternalIP or InternalIP.
	Type MinerAddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=MinerAddressType"`

	// The machine address.
	Address string `json:"address" protobuf:"bytes,2,opt,name=address"`
}

// MinerAddresses is a slice of MinerAddress items to be used by infrastructure providers.
type MinerAddresses []MinerAddress

// ObjectMeta is metadata that all persisted resources must have, which includes all objects
// users must create. This is a copy of customizable fields from metav1.ObjectMeta.
//
// ObjectMeta is embedded in `Miner.Spec` and `MinerSet.Template`,
// which are not top-level Kubernetes objects. Given that metav1.ObjectMeta has lots of special cases
// and read-only fields which end up in the generated CRD validation, having it as a subset simplifies
// the API and some issues that can impact user experience.
//
// During the [upgrade to controller-tools@v2](https://github.com/kubernetes-sigs/cluster-api/pull/1054)
// for v1alpha2, we noticed a failure would occur running Cluster API test suite against the new CRDs,
// specifically `spec.metadata.creationTimestamp in body must be of type string: "null"`.
// The investigation showed that `controller-tools@v2` behaves differently than its previous version
// when handling types from [metav1](k8s.io/apimachinery/pkg/apis/meta/v1) package.
//
// In more details, we found that embedded (non-top level) types that embedded `metav1.ObjectMeta`
// had validation properties, including for `creationTimestamp` (metav1.Time).
// The `metav1.Time` type specifies a custom json marshaller that, when IsZero() is true, returns `null`
// which breaks validation because the field isn't marked as nullable.
//
// In future versions, controller-tools@v2 might allow overriding the type and validation for embedded
// types. When that happens, this hack should be revisited.
type ObjectMeta struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,1,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,2,rep,name=annotations"`
}
