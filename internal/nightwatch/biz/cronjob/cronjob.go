package cronjob

//go:generate mockgen -destination mock_cronjob.go -package cronjob github.com/superproj/onex/internal/nightwatch/biz/cronjob CronJobBiz

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/store"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/store/where"
)

// CronJobBiz defines the interface for managing cron jobs.
type CronJobBiz interface {
	Create(ctx context.Context, rq *nwv1.CreateCronJobRequest) (*nwv1.CreateCronJobResponse, error)
	Update(ctx context.Context, rq *nwv1.UpdateCronJobRequest) (*nwv1.UpdateCronJobResponse, error)
	Delete(ctx context.Context, rq *nwv1.DeleteCronJobRequest) (*nwv1.DeleteCronJobResponse, error)
	Get(ctx context.Context, rq *nwv1.GetCronJobRequest) (*nwv1.GetCronJobResponse, error)
	List(ctx context.Context, rq *nwv1.ListCronJobRequest) (*nwv1.ListCronJobResponse, error)
}

// cronJobBiz is the concrete implementation of the CronJobBiz interface.
type cronJobBiz struct {
	ds store.IStore
}

// Ensure cronJobBiz implements the CronJobBiz interface.
var _ CronJobBiz = (*cronJobBiz)(nil)

// New creates a new instance of cronJobBiz with the provided data store.
func New(ds store.IStore) *cronJobBiz {
	return &cronJobBiz{ds: ds}
}

// Create adds a new cron job to the data store.
func (b *cronJobBiz) Create(ctx context.Context, rq *nwv1.CreateCronJobRequest) (*nwv1.CreateCronJobResponse, error) {
	var cronJobM model.CronJobM
	_ = copier.Copy(&cronJobM, rq.CronJob) // Copy request data to the model.

	if err := b.ds.CronJobs().Create(ctx, &cronJobM); err != nil {
		return nil, err
	}

	return &nwv1.CreateCronJobResponse{CronJobID: cronJobM.CronJobID}, nil
}

// Update modifies an existing cron job in the data store.
func (b *cronJobBiz) Update(ctx context.Context, rq *nwv1.UpdateCronJobRequest) (*nwv1.UpdateCronJobResponse, error) {
	cronJobM, err := b.ds.CronJobs().Get(ctx, where.F("cronjob_id", rq.CronJobID))
	if err != nil {
		return nil, err
	}

	if rq.Name != nil {
		cronJobM.Name = *rq.Name
	}
	if rq.Description != nil {
		cronJobM.Description = *rq.Description
	}
	if rq.Schedule != nil {
		cronJobM.Schedule = *rq.Schedule
	}
	if rq.ConcurrencyPolicy != nil {
		cronJobM.ConcurrencyPolicy = *rq.ConcurrencyPolicy
	}
	if rq.Suspend != nil {
		cronJobM.Suspend = *rq.Suspend
	}
	if rq.SuccessHistoryLimit != nil {
		cronJobM.SuccessHistoryLimit = *rq.SuccessHistoryLimit
	}
	if rq.FailedHistoryLimit != nil {
		cronJobM.FailedHistoryLimit = *rq.FailedHistoryLimit
	}

	if err := b.ds.CronJobs().Update(ctx, cronJobM); err != nil {
		return nil, err
	}

	return &nwv1.UpdateCronJobResponse{}, nil
}

// Delete removes one or more cron jobs by their IDs from the data store.
func (b *cronJobBiz) Delete(ctx context.Context, rq *nwv1.DeleteCronJobRequest) (*nwv1.DeleteCronJobResponse, error) {
	if err := b.ds.CronJobs().Delete(ctx, where.F("cronjob_id", rq.CronJobIDs)); err != nil {
		return nil, err
	}

	return &nwv1.DeleteCronJobResponse{}, nil
}

// Get retrieves a cron job by its ID from the data store.
func (b *cronJobBiz) Get(ctx context.Context, rq *nwv1.GetCronJobRequest) (*nwv1.GetCronJobResponse, error) {
	cronJob, err := b.ds.CronJobs().Get(ctx, where.F("cronjob_id", rq.CronJobID))
	if err != nil {
		return nil, err
	}

	var resp nwv1.GetCronJobResponse
	_ = copier.Copy(&resp.CronJob, cronJob) // Copy model data to the response.

	return &resp, nil
}

// List retrieves all cron jobs from the data store.
func (b *cronJobBiz) List(ctx context.Context, rq *nwv1.ListCronJobRequest) (*nwv1.ListCronJobResponse, error) {
	count, cronJobList, err := b.ds.CronJobs().List(ctx, where.NewWhere(where.WithPage(rq.Offset, rq.Limit)))
	if err != nil {
		return nil, err
	}

	cronJobs := make([]*nwv1.CronJob, len(cronJobList))
	for i, item := range cronJobList {
		_ = copier.Copy(&cronJobs[i], item)
	}

	return &nwv1.ListCronJobResponse{TotalCount: &count, CronJobs: cronJobs}, nil
}
