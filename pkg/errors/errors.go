// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package errors

// MinerStatusError defines errors states for Miner objects.
type MinerStatusError string

// Constants aren't automatically generated for unversioned packages.
// Instead share the same constant for all versioned packages.

const (
	// InvalidConfigurationMinerError represents that the combination
	// of configuration in the MinerSpec is not supported by this cluster.
	// This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	//
	// Example: the ProviderSpec specifies an instance type that doesn't exist,.
	InvalidConfigurationMinerError MinerStatusError = "InvalidConfiguration"

	// UnsupportedChangeMinerError indicates that the MinerSpec has been updated in a way that
	// is not supported for reconciliation on this cluster. The spec may be
	// completely valid from a configuration standpoint, but the controller
	// does not support changing the real world state to match the new
	// spec.
	//
	// Example: the responsible controller is not capable of changing the
	// container runtime from docker to rkt.
	UnsupportedChangeMinerError MinerStatusError = "UnsupportedChange"

	// InsufficientResourcesMinerError generally refers to exceeding one's quota in a cloud provider,
	// or running out of physical miners in an on-premise environment.
	InsufficientResourcesMinerError MinerStatusError = "InsufficientResources"

	// CreateMinerError indicates an error while trying to create a Node to match this
	// Miner. This may indicate a transient problem that will be fixed
	// automatically with time, such as a service outage, or a terminal
	// error during creation that doesn't match a more specific
	// MinerStatusError value.
	//
	// Example: timeout trying to connect to GCE.
	CreateMinerError MinerStatusError = "CreateError"

	// UpdateMinerError indicates an error while trying to update a Node that this
	// Miner represents. This may indicate a transient problem that will be
	// fixed automatically with time, such as a service outage,
	//
	// Example: error updating load balancers.
	UpdateMinerError MinerStatusError = "UpdateError"

	// DeleteMinerError indicates an error was encountered while trying to delete the Node that this
	// Miner represents. This could be a transient or terminal error, but
	// will only be observable if the provider's Miner controller has
	// added a finalizer to the object to more gracefully handle deletions.
	//
	// Example: cannot resolve EC2 IP address.
	DeleteMinerError MinerStatusError = "DeleteError"

	// JoinClusterTimeoutMinerError indicates that the miner did not join the cluster
	// as a new node within the expected timeframe after instance
	// creation at the provider succeeded
	//
	// Example use case: A controller that deletes Miners which do
	// not result in a Node joining the cluster within a given timeout
	// and that are managed by a MinerSet.
	JoinClusterTimeoutMinerError = "JoinClusterTimeoutError"
)

// MinerSetStatusError defines errors states for MinerSet objects.
type MinerSetStatusError string

const (
	// InvalidConfigurationMinerSetError represents
	// the combination of configuration in the MinerTemplateSpec
	// is not supported by this cluster. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	//
	// Example: the ProviderSpec specifies an instance type that doesn't exist.
	InvalidConfigurationMinerSetError MinerSetStatusError = "InvalidConfiguration"
)
