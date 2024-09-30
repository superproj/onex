package job

//go:generate mockgen -destination mock_job.go -package job github.com/superproj/onex/internal/nightwatch/biz/job JobBiz

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/store"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/store/where"
)

// JobBiz defines the interface for managing jobs.
type JobBiz interface {
	Create(ctx context.Context, rq *nwv1.CreateJobRequest) (*nwv1.CreateJobResponse, error)
	Update(ctx context.Context, rq *nwv1.UpdateJobRequest) (*nwv1.UpdateJobResponse, error)
	Delete(ctx context.Context, rq *nwv1.DeleteJobRequest) (*nwv1.DeleteJobResponse, error)
	Get(ctx context.Context, rq *nwv1.GetJobRequest) (*nwv1.GetJobResponse, error)
	List(ctx context.Context, rq *nwv1.ListJobRequest) (*nwv1.ListJobResponse, error)
}

// jobBiz is the concrete implementation of the JobBiz interface.
type jobBiz struct {
	ds store.IStore
}

// Ensure jobBiz implements the JobBiz interface.
var _ JobBiz = (*jobBiz)(nil)

// New creates a new instance of jobBiz with the provided data store.
func New(ds store.IStore) *jobBiz {
	return &jobBiz{ds: ds}
}

// Create adds a new job to the data store.
func (b *jobBiz) Create(ctx context.Context, rq *nwv1.CreateJobRequest) (*nwv1.CreateJobResponse, error) {
	var jobM model.JobM
	_ = copier.Copy(&jobM, rq.Job) // Copy request data to the model.

	if err := b.ds.Jobs().Create(ctx, &jobM); err != nil {
		return nil, err
	}

	return &nwv1.CreateJobResponse{JobID: jobM.JobID}, nil
}

// Update modifies an existing job in the data store.
func (b *jobBiz) Update(ctx context.Context, rq *nwv1.UpdateJobRequest) (*nwv1.UpdateJobResponse, error) {
	jobM, err := b.ds.Jobs().Get(ctx, where.F("job_id", rq.JobID))
	if err != nil {
		return nil, err
	}

	if rq.Name != nil {
		jobM.Name = *rq.Name
	}
	if rq.Description != nil {
		jobM.Description = *rq.Description
	}
	if rq.Schedule != nil {
		jobM.Schedule = *rq.Schedule
	}
	if rq.ConcurrencyPolicy != nil {
		jobM.ConcurrencyPolicy = *rq.ConcurrencyPolicy
	}
	if rq.Suspend != nil {
		jobM.Suspend = *rq.Suspend
	}
	if rq.SuccessHistoryLimit != nil {
		jobM.SuccessHistoryLimit = *rq.SuccessHistoryLimit
	}
	if rq.FailedHistoryLimit != nil {
		jobM.FailedHistoryLimit = *rq.FailedHistoryLimit
	}

	if err := b.ds.Jobs().Update(ctx, jobM); err != nil {
		return nil, err
	}

	return &nwv1.UpdateJobResponse{}, nil
}

// Delete removes one or more jobs by their IDs from the data store.
func (b *jobBiz) Delete(ctx context.Context, rq *nwv1.DeleteJobRequest) (*nwv1.DeleteJobResponse, error) {
	if err := b.ds.Jobs().Delete(ctx, where.F("job_id", rq.JobIDs)); err != nil {
		return nil, err
	}

	return &nwv1.DeleteJobResponse{}, nil
}

// Get retrieves a job by its ID from the data store.
func (b *jobBiz) Get(ctx context.Context, rq *nwv1.GetJobRequest) (*nwv1.GetJobResponse, error) {
	job, err := b.ds.Jobs().Get(ctx, where.F("job_id", rq.JobID))
	if err != nil {
		return nil, err
	}

	var resp nwv1.GetJobResponse
	_ = copier.Copy(&resp.Job, job) // Copy model data to the response.

	return &resp, nil
}

// List retrieves all jobs from the data store.
func (b *jobBiz) List(ctx context.Context, rq *nwv1.ListJobRequest) (*nwv1.ListJobResponse, error) {
	count, jobList, err := b.ds.Jobs().List(ctx, where.NewWhere(where.WithPage(rq.Offset, rq.Limit)))
	if err != nil {
		return nil, err
	}

	jobs := make([]*nwv1.Job, len(jobList))
	for i, item := range jobList {
		_ = copier.Copy(&jobs[i], item)
	}

	return &nwv1.ListJobResponse{TotalCount: &count, Jobs: jobs}, nil
}