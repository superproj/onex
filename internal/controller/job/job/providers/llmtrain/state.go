// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package llmtrain implements the "LLMTrain" Job provider, which runs a Job as
// an in-process LLM training pipeline (download -> embed -> train). It is a
// self-contained port of the reference opsassist llmtrain state machine,
// adapted to the controller-runtime provider contract.
package llmtrain

import "github.com/onexstack/onex/pkg/apis/batch/v1beta1"

const (
	// Type is the Job type for the LLM training provider.
	Type v1beta1.JobType = "LLMTrain"
)

// LLM training phases, mirroring the reference opsassist llmtrain state machine.
const (
	phasePending     = "Pending"
	phaseDownloading = "Downloading"
	phaseDownloaded  = "Downloaded"
	phaseEmbedding   = "Embedding"
	phaseEmbedded    = "Embedded"
	phaseTraining    = "Training"
	phaseTrained     = "Trained"
	phaseSucceeded   = "Succeeded"
	phaseFailed      = "Failed"
)

const (
	// defaultJobTimeout is the default maximum training duration (seconds),
	// matching the reference implementation's 4-hour timeout.
	defaultJobTimeout int64 = 14400
)

// Spec is the LLMTrain provider configuration, decoded from
// job.Spec.ProviderSpec.Raw.
type Spec struct {
	// SourceObject is the object-storage key holding the raw training data.
	SourceObject string `json:"sourceObject"`
	// JobTimeoutSec overrides the maximum training duration in seconds.
	JobTimeoutSec int64 `json:"jobTimeoutSec"`
	// IdempotentExecution skips already-completed phases when true.
	IdempotentExecution bool `json:"idempotentExecution"`
}

// Status is the mutable LLMTrain pipeline state, serialized into
// job.Status.ProviderStatus.Raw so it survives controller restarts.
type Status struct {
	// Phase is the current pipeline phase.
	Phase string `json:"phase"`
	// DataPath is the object-storage key of the downloaded raw data.
	DataPath *string `json:"dataPath,omitempty"`
	// EmbeddedDataPath is the object-storage key of the embedded data.
	EmbeddedDataPath *string `json:"embeddedDataPath,omitempty"`
	// TaskID is the training task identifier, set once on first submit.
	TaskID *string `json:"taskID,omitempty"`
	// ResultPath is the object-storage key where the trained model is written.
	ResultPath *string `json:"resultPath,omitempty"`
	// Message carries a terminal failure reason.
	Message string `json:"message,omitempty"`
}
