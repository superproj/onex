package manager

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
)

// JobManager manages cron jobs.
type JobManager struct {
	mu            sync.Mutex              // Mutex for synchronizing access to jobs
	cronScheduler *cron.Cron              // The cron scheduler instance
	jobs          map[string]cron.EntryID // Map to store job names and their Entry IDs
}

// Option defines a function type that configures JobManager options.
type Option func(jm *JobManager)

// WithCron is an Option that allows a custom cron scheduler to be set.
func WithCron(c *cron.Cron) Option {
	return func(jm *JobManager) {
		jm.cronScheduler = c
	}
}

// NewJobManager creates a new instance of JobManager.
func NewJobManager(opts ...Option) *JobManager {
	jm := &JobManager{
		cronScheduler: cron.New(),
		jobs:          make(map[string]cron.EntryID),
	}

	// Set with custom options
	for _, opt := range opts {
		opt(jm) // Invoke each option function with the job manager instance
	}

	return jm
}

// AddJob adds a new cron job to the manager.
func (jm *JobManager) AddJob(jobName string, schedule string, cmd cron.Job) (cron.EntryID, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Check if the job already exists
	if _, exists := jm.jobs[jobName]; exists {
		return 0, &JobExistsError{JobName: jobName}
	}

	// Add the job to the cron scheduler
	entryID, err := jm.cronScheduler.AddJob(schedule, cmd)
	if err != nil {
		return 0, err // Return error if adding the job fails
	}

	// Store the job in the map
	jm.jobs[jobName] = entryID
	return entryID, nil
}

// RemoveJob removes a specified cron job from the manager.
func (jm *JobManager) RemoveJob(jobName string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	entryID, exists := jm.jobs[jobName]
	if !exists {
		return nil
	}

	// Remove the job from the cron scheduler and delete it from the map
	jm.cronScheduler.Remove(entryID)
	delete(jm.jobs, jobName)
	return nil

}

// UpdateJob updates a specified cron job with a new schedule and function.
func (jm *JobManager) UpdateJob(jobName string, schedule string, cmd cron.Job) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Check if the job exists before attempting to remove it
	if _, exists := jm.jobs[jobName]; !exists {
		return &JobNotFoundError{JobName: jobName}
	}

	// Remove the existing job and add it again with new parameters
	err := jm.RemoveJob(jobName)
	if err != nil {
		return err
	}

	_, err = jm.AddJob(jobName, schedule, cmd)
	return err
}

// GetJobs returns a map of all the current cron jobs.
func (jm *JobManager) GetJobs() map[string]cron.EntryID {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	return jm.jobs
}

// JobExists checks if a specific job exists in the manager.
func (jm *JobManager) JobExists(jobName string) bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	_, exists := jm.jobs[jobName]
	return exists // Return true if the job exists, false otherwise
}

// Start starts the cron scheduler to begin executing jobs.
func (jm *JobManager) Start() {
	jm.cronScheduler.Start()
}

// Stop stops the cron scheduler.
func (jm *JobManager) Stop() context.Context {
	return jm.cronScheduler.Stop()
}
