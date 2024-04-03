// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package apps

// MinerPhase is a string representation of a Miner Phase.
//
// This type is a high-level indicator of the status of the Miner as it is provisioned,
// from the API user’s perspective.
//
// The value should not be interpreted by any software components as a reliable indication
// of the actual state of the Miner, and controllers should not use the Miner Phase field
// value when making decisions about what action to take.
//
// Controllers should always look at the actual state of the Miner’s fields to make those decisions.
type MinerPhase string

const (
	// MinerPhasePending is the first state a Miner is assigned by
	// Cluster API Miner controller after being created.
	MinerPhasePending = MinerPhase("Pending")

	// MinerPhaseProvisioning is the state when the
	// Miner infrastructure is being created.
	MinerPhaseProvisioning = MinerPhase("Provisioning")

	// MinerPhaseProvisioned is the state when its
	// infrastructure has been created and configured.
	MinerPhaseProvisioned = MinerPhase("Provisioned")

	// MinerPhaseRunning is the Miner state when it has
	// become a Kubernetes Node in a Ready state.
	MinerPhaseRunning = MinerPhase("Running")

	// MinerPhaseDeleting is the Miner state when a delete
	// request has been sent to the API Server,
	// but its infrastructure has not yet been fully deleted.
	MinerPhaseDeleting = MinerPhase("Deleting")

	// MinerPhaseDeleted is the Miner state when the object
	// and the related infrastructure is deleted and
	// ready to be garbage collected by the API Server.
	MinerPhaseDeleted = MinerPhase("Deleted")

	// MinerPhaseFailed is the Miner state when the system
	// might require user intervention.
	MinerPhaseFailed = MinerPhase("Failed")

	// MinerPhaseUnknown is returned if the Miner state cannot be determined.
	MinerPhaseUnknown = MinerPhase("Unknown")
)
