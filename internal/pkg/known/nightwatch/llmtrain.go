package known

import (
	stringsutil "github.com/superproj/onex/pkg/util/strings"
)

// LLM train statuses represent the various phases of llm train.
// These constants are derived from the job status constants and include additional states
// relevant to the llm train process.
const (
	// LLMTrainSucceeded indicates that the llm train has successfully completed.
	LLMTrainSucceeded = JobSucceeded
	// LLMTrainFailed indicates that the llm train has failed.
	LLMTrainFailed = JobFailed

	// LLMTrainPending indicates that the llm train is pending.
	LLMTrainPending = JobPending
	// LLMTrainDownloading indicates that the llm train data is
	// being downloaded.
	LLMTrainDownloading = "Downloading"
	// LLMTrainDownloaded indicates that the llm train data has
	// been downloaded.
	LLMTrainDownloaded = "Downloaded"
	// LLMTrainEmbedding indicates that the llm train is in the
	// embedding phase.
	LLMTrainEmbedding = "Embedding"
	// LLMTrainEmbedded indicates that the embedding process has completed.
	LLMTrainEmbedded = "Embedded"
	// LLMTrainTraining indicates that the llm is currently
	// being trained.
	LLMTrainTraining = "Training"
	// LLMTrainTrained indicates that the train of the llm has completed.
	LLMTrainTrained = "Trained"
)

// LLMTrainTimeout defines the maximum duration (in seconds) allowed for training jobs.
// This constant is set to 14400 seconds, which equals 4 hours.
const LLMTrainTimeout = 14400

// Rate Limits for controlling the concurrency and frequency of job execution.
const (
	// LLMTrainMaxWorkers specify the maximum number of workers
	// allowed for llm train jobs.
	LLMTrainMaxWorkers = 5
	// LLMTrainEmbeddingQPS specify the maximum queries per second
	// for the embedding process during llm train.
	LLMTrainEmbeddingQPS = 10
	// LLMTrainEvaluateQPS specify the maximum queries per second
	// for the evaluation process during llm train.
	LLMTrainEvaluateQPS = 150
)

// Additional constants for llm train.
const (
	// MaxLLMTrainFeedbacks specify the maximum number of feedback entries
	// allowed for llm train.
	MaxLLMTrainFeedbacks = 3000
	// LLMTrainHitRatePrecision specify the precision for calculating hit rate
	// during llm train.
	LLMTrainHitRatePrecision = 0.0001
)

func StandardLLMTrainStatus(trainStatus string) string {
	if !stringsutil.StringIn(trainStatus, []string{LLMTrainFailed, LLMTrainSucceeded, LLMTrainPending}) {
		return JobRunning
	}

	return trainStatus
}
