// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

// ANCHOR: CommonConditions

// Common ConditionTypes used by Cluster API objects.
const (
	// ReadyCondition defines the Ready condition type that summarizes the operational state of a Cluster API object.
	ReadyCondition ConditionType = "Ready"
)

// Common ConditionReason used by Cluster API objects.
const (
	// DeletingReason (Severity=Info) documents a condition not in Status=True because the underlying object it is currently being deleted.
	DeletingReason = "Deleting"

	// DeletionFailedReason (Severity=Warning) documents a condition not in Status=True because the underlying object
	// encountered problems during deletion. This is a warning because the reconciler will retry deletion.
	DeletionFailedReason = "DeletionFailed"

	// DeletedReason (Severity=Info) documents a condition not in Status=True because the underlying object was deleted.
	DeletedReason = "Deleted"

	// IncorrectExternalRefReason (Severity=Error) documents a CAPI object with an incorrect external object reference.
	IncorrectExternalRefReason = "IncorrectExternalRef"
)

const (
	// InfrastructureReadyCondition reports a summary of current status of the infrastructure object defined for this cluster/miner/minerpool.
	// This condition is mirrored from the Ready condition in the infrastructure ref object, and
	// the absence of this condition might signal problems in the reconcile external loops or the fact that
	// the infrastructure provider does not implement the Ready condition yet.
	InfrastructureReadyCondition ConditionType = "InfrastructureReady"

	// WaitingForInfrastructureFallbackReason (Severity=Info) documents a cluster/miner/minerpool waiting for the underlying infrastructure
	// to be available.
	// NOTE: This reason is used only as a fallback when the infrastructure object is not reporting its own ready condition.
	WaitingForInfrastructureFallbackReason = "WaitingForInfrastructure"
)

// ANCHOR_END: CommonConditions

// Conditions and condition Reasons for the Cluster object.

const (
	// ControlPlaneInitializedCondition reports if the cluster's control plane has been initialized such that the
	// cluster's apiserver is reachable and at least one control plane Miner has a node reference. Once this
	// condition is marked true, its value is never changed. See the ControlPlaneReady condition for an indication of
	// the current readiness of the cluster's control plane.
	ControlPlaneInitializedCondition ConditionType = "ControlPlaneInitialized"

	// MissingPodRefReason (Severity=Info) documents a cluster waiting for at least one control plane Miner to have
	// its node reference populated.
	MissingPodRefReason = "MissingPodRef"

	// WaitingForControlPlaneProviderInitializedReason (Severity=Info) documents a cluster waiting for the control plane
	// provider to report successful control plane initialization.
	WaitingForControlPlaneProviderInitializedReason = "WaitingForControlPlaneProviderInitialized"

	// ControlPlaneReadyCondition reports the ready condition from the control plane object defined for this cluster.
	// This condition is mirrored from the Ready condition in the control plane ref object, and
	// the absence of this condition might signal problems in the reconcile external loops or the fact that
	// the control plane provider does not implement the Ready condition yet.
	ControlPlaneReadyCondition ConditionType = "ControlPlaneReady"

	// WaitingForControlPlaneFallbackReason (Severity=Info) documents a cluster waiting for the control plane
	// to be available.
	// NOTE: This reason is used only as a fallback when the control plane object is not reporting its own ready condition.
	WaitingForControlPlaneFallbackReason = "WaitingForControlPlane"

	// WaitingForControlPlaneAvailableReason (Severity=Info) documents a Cluster API object
	// waiting for the control plane miner to be available.
	//
	// NOTE: Having the control plane miner available is a pre-condition for joining additional control planes
	// or workers nodes.
	WaitingForControlPlaneAvailableReason = "WaitingForControlPlaneAvailable"
)

// Conditions and condition Reasons for the Miner object.

const (
	// BootstrapReadyCondition reports a summary of current status of the bootstrap object defined for this miner.
	// This condition is mirrored from the Ready condition in the bootstrap ref object, and
	// the absence of this condition might signal problems in the reconcile external loops or the fact that
	// the bootstrap provider does not implement the Ready condition yet.
	BootstrapReadyCondition ConditionType = "BootstrapReady"

	// WaitingForDataSecretFallbackReason (Severity=Info) documents a miner waiting for the bootstrap data secret
	// to be available.
	// NOTE: This reason is used only as a fallback when the bootstrap object is not reporting its own ready condition.
	WaitingForDataSecretFallbackReason = "WaitingForDataSecret"

	// DrainingSucceededCondition provide evidence of the status of the node drain operation which happens during the miner
	// deletion process.
	DrainingSucceededCondition ConditionType = "DrainingSucceeded"

	// DrainingReason (Severity=Info) documents a miner node being drained.
	DrainingReason = "Draining"

	// DrainingFailedReason (Severity=Warning) documents a miner node drain operation failed.
	DrainingFailedReason = "DrainingFailed"

	// PreDrainDeleteHookSucceededCondition reports a miner waiting for a PreDrainDeleteHook before being delete.
	PreDrainDeleteHookSucceededCondition ConditionType = "PreDrainDeleteHookSucceeded"

	// PreTerminateDeleteHookSucceededCondition reports a miner waiting for a PreDrainDeleteHook before being delete.
	PreTerminateDeleteHookSucceededCondition ConditionType = "PreTerminateDeleteHookSucceeded"

	// WaitingExternalHookReason (Severity=Info) provide evidence that we are waiting for an external hook to complete.
	WaitingExternalHookReason = "WaitingExternalHook"

	// VolumeDetachSucceededCondition reports a miner waiting for volumes to be detached.
	VolumeDetachSucceededCondition ConditionType = "VolumeDetachSucceeded"

	// WaitingForVolumeDetachReason (Severity=Info) provide evidence that a miner node waiting for volumes to be attached.
	WaitingForVolumeDetachReason = "WaitingForVolumeDetach"
)

const (
	// MinerHealthCheckSucceededCondition is set on miners that have passed a healthcheck by the MinerHealthCheck controller.
	// In the event that the health check fails it will be set to False.
	MinerHealthCheckSucceededCondition ConditionType = "HealthCheckSucceeded"

	// MinerHealthCheckSuccededCondition is set on miners that have passed a healthcheck by the MinerHealthCheck controller.
	// In the event that the health check fails it will be set to False.
	// Deprecated: This const is going to be removed in a next release. Use MinerHealthCheckSucceededCondition instead.
	MinerHealthCheckSuccededCondition ConditionType = "HealthCheckSucceeded"

	// MinerHasFailureReason is the reason used when a miner has either a FailureReason or a FailureMessage set on its status.
	MinerHasFailureReason = "MinerHasFailure"

	// PodStartupTimeoutReason is the reason used when a miner's node does not appear within the specified timeout.
	PodStartupTimeoutReason = "PodStartupTimeout"

	// UnhealthyPodConditionReason is the reason used when a miner's node has one of the MinerHealthCheck's unhealthy conditions.
	UnhealthyPodConditionReason = "UnhealthyPod"
)

const (
	// MinerOwnerRemediatedCondition is set on miners that have failed a healthcheck by the MinerHealthCheck controller.
	// MinerOwnerRemediatedCondition is set to False after a health check fails, but should be changed to True by the owning controller after remediation succeeds.
	MinerOwnerRemediatedCondition ConditionType = "OwnerRemediated"

	// WaitingForRemediationReason is the reason used when a miner fails a health check and remediation is needed.
	WaitingForRemediationReason = "WaitingForRemediation"

	// RemediationFailedReason is the reason used when a remediation owner fails to remediate an unhealthy miner.
	RemediationFailedReason = "RemediationFailed"

	// RemediationInProgressReason is the reason used when an unhealthy miner is being remediated by the remediation owner.
	RemediationInProgressReason = "RemediationInProgress"

	// ExternalRemediationTemplateAvailable is set on minerhealthchecks when MinerHealthCheck controller uses external remediation.
	// ExternalRemediationTemplateAvailable is set to false if external remediation template is not found.
	ExternalRemediationTemplateAvailable ConditionType = "ExternalRemediationTemplateAvailable"

	// ExternalRemediationTemplateNotFound is the reason used when a miner health check fails to find external remediation template.
	ExternalRemediationTemplateNotFound = "ExternalRemediationTemplateNotFound"

	// ExternalRemediationRequestAvailable is set on minerhealthchecks when MinerHealthCheck controller uses external remediation.
	// ExternalRemediationRequestAvailable is set to false if creating external remediation request fails.
	ExternalRemediationRequestAvailable ConditionType = "ExternalRemediationRequestAvailable"

	// ExternalRemediationRequestCreationFailed is the reason used when a miner health check fails to create external remediation request.
	ExternalRemediationRequestCreationFailed = "ExternalRemediationRequestCreationFailed"
)

// Conditions and condition Reasons for the Miner's Pod object.
const (
	// MinerPodHealthyCondition provides info about the operational state of the Kubernetes node hosted on the miner by summarizing  node conditions.
	// If the conditions defined in a Kubernetes node (i.e., PodReady, PodMemoryPressure, PodDiskPressure, PodPIDPressure, and PodNetworkUnavailable) are in a healthy state, it will be set to True.
	MinerPodHealthyCondition ConditionType = "PodHealthy"

	// WaitingForPodRefReason (Severity=Info) documents a miner.spec.providerId is not assigned yet.
	WaitingForPodRefReason = "WaitingForPodRef"

	// PodProvisioningReason (Severity=Info) documents miner in the process of provisioning a node.
	// NB. provisioning --> PodRef == "".
	PodProvisioningReason = "PodProvisioning"

	// PodNotFoundReason (Severity=Error) documents a miner's node has previously been observed but is now gone.
	// NB. provisioned --> PodRef != "".
	PodNotFoundReason = "PodNotFound"

	// PodConditionsFailedReason (Severity=Warning) documents a node is not in a healthy state due to the failed state of at least 1 Kubelet condition.
	PodConditionsFailedReason = "PodConditionsFailed"
)

// Conditions and condition Reasons for the MinerHealthCheck object.

const (
	// RemediationAllowedCondition is set on MinerHealthChecks to show the status of whether the MinerHealthCheck is
	// allowed to remediate any Miners or whether it is blocked from remediating any further.
	RemediationAllowedCondition ConditionType = "RemediationAllowed"

	// TooManyUnhealthyReason is the reason used when too many Miners are unhealthy and the MinerHealthCheck is blocked
	// from making any further remediations.
	TooManyUnhealthyReason = "TooManyUnhealthy"
)

// Conditions and condition Reasons for  MinerDeployments.

const (
	// MinerDeploymentAvailableCondition means the MinerDeployment is available, that is, at least the minimum available
	// miners required (i.e. Spec.Replicas-MaxUnavailable when MinerDeploymentStrategyType = RollingUpdate) are up and running for at least minReadySeconds.
	MinerDeploymentAvailableCondition ConditionType = "Available"

	// WaitingForAvailableMinersReason (Severity=Warning) reflects the fact that the required minimum number of miners for a minerdeployment are not available.
	WaitingForAvailableMinersReason = "WaitingForAvailableMiners"
)

// Conditions and condition Reasons for  MinerSets.

const (
	// MinersCreatedCondition documents that the miners controlled by the MinerSet are created.
	// When this condition is false, it indicates that there was an error when cloning the infrastructure/bootstrap template or
	// when generating the miner object.
	MinersCreatedCondition ConditionType = "MinersCreated"

	// MinersReadyCondition reports an aggregate of current status of the miners controlled by the MinerSet.
	MinersReadyCondition ConditionType = "MinersReady"

	// BootstrapTemplateCloningFailedReason (Severity=Error) documents a MinerSet failing to
	// clone the bootstrap template.
	BootstrapTemplateCloningFailedReason = "BootstrapTemplateCloningFailed"

	// InfrastructureTemplateCloningFailedReason (Severity=Error) documents a MinerSet failing to
	// clone the infrastructure template.
	InfrastructureTemplateCloningFailedReason = "InfrastructureTemplateCloningFailed"

	// MinerCreationFailedReason (Severity=Error) documents a MinerSet failing to
	// generate a miner object.
	MinerCreationFailedReason = "MinerCreationFailed"

	// ResizedCondition documents a MinerSet is resizing the set of controlled miners.
	ResizedCondition ConditionType = "Resized"

	// ScalingUpReason (Severity=Info) documents a MinerSet is increasing the number of replicas.
	ScalingUpReason = "ScalingUp"

	// ScalingDownReason (Severity=Info) documents a MinerSet is decreasing the number of replicas.
	ScalingDownReason = "ScalingDown"

	ConfigMapsCreatedCondition ConditionType = "ConfigMapsCreated"

	ConfigMapCreationFailedReason = "ConfigMapCreationFailed"
)

// Conditions and condition reasons for Clusters with a managed Topology.
const (
	// TopologyReconciledCondition provides evidence about the reconciliation of a Cluster topology into
	// the managed objects of the Cluster.
	// Status false means that for any reason, the values defined in Cluster.spec.topology are not yet applied to
	// managed objects on the Cluster; status true means that Cluster.spec.topology have been applied to
	// the objects in the Cluster (but this does not imply those objects are already reconciled to the spec provided).
	TopologyReconciledCondition ConditionType = "TopologyReconciled"

	// TopologyReconcileFailedReason (Severity=Error) documents the reconciliation of a Cluster topology
	// failing due to an error.
	TopologyReconcileFailedReason = "TopologyReconcileFailed"

	// TopologyReconciledControlPlaneUpgradePendingReason (Severity=Info) documents reconciliation of a Cluster topology
	// not yet completed because Control Plane is not yet updated to match the desired topology spec.
	TopologyReconciledControlPlaneUpgradePendingReason = "ControlPlaneUpgradePending"

	// TopologyReconciledMinerDeploymentsUpgradePendingReason (Severity=Info) documents reconciliation of a Cluster topology
	// not yet completed because at least one of the MinerDeployments is not yet updated to match the desired topology spec.
	TopologyReconciledMinerDeploymentsUpgradePendingReason = "MinerDeploymentsUpgradePending"

	// TopologyReconciledHookBlockingReason (Severity=Info) documents reconciliation of a Cluster topology
	// not yet completed because at least one of the lifecycle hooks is blocking.
	TopologyReconciledHookBlockingReason = "LifecycleHookBlocking"
)

const (
	// Approved indicates the charge request was approved.
	ChargeApproved ConditionType = "Approved"
)
