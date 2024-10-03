package known

// These are the valid statuses of jobs.
// The constants defined here represent the possible states that a Job can be in
// during its lifecycle in eam-nightwatch.
const (
	// JobPending indicates job has been created but is not yet running.
	// It's pending for resources or scheduling.
	JobPending string = "Pending"
	// JobRunning indicates job is currently running, and its watcher are
	// executing the specified task.
	JobRunning string = "Running"
	// JobSucceeded indicates job has successfully completed all of its
	// tasks with a successful exit status.
	JobSucceeded string = "Succeeded"
	// JobFailed indicates job has failed and has reached a state where
	// it could not complete its tasks successfully.
	JobFailed string = "Failed"
)

// Job Scope defines the scope of the job for organizational purposes.
const (
	LLMJobScope = "llm"
)

// Job Watcher identifiers for monitoring specific job types.
const (
	// TrainWatcher identifier for llm train job watcher.
	LLMTrainWatcher = "llmtrain"
)

const (
	// JobStatusNonSuspended indicates that the job is currently active and not suspended.
	JobNonSuspended = 0
	// JobStatusSuspended indicates that the job is currently suspended and not active.
	JobSuspended = 1
)

// MaxJobsPerCronJob defines the maximum number of jobs that can be scheduled
// to run concurrently for a single cron job. This limit helps to prevent
// resource exhaustion and ensures that the system remains stable under load.
const (
	MaxJobsPerCronJob = 50
)

// Job Execution Idempotency indicates whether the job can be executed
// multiple times without changing the result.
const (
	// NonIdempotentExecution indicates that the execution is non-idempotent.
	NonIdempotentExecution int64 = 0
	// IdempotentExecution indicates that the execution is idempotent.
	IdempotentExecution int64 = 1
)
