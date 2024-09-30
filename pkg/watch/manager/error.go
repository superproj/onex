package manager

import "fmt"

// JobExistsError represents an error when a job already exists.
type JobExistsError struct {
	JobName string
}

func (e *JobExistsError) Error() string {
	return fmt.Sprintf("job %s already exists", e.JobName)
}

// JobNotFoundError represents an error when a job does not exist.
type JobNotFoundError struct {
	JobName string
}

// Error implements the error interface for JobNotFoundError.
func (e *JobNotFoundError) Error() string {
	return fmt.Sprintf("job %s not found", e.JobName)
}
