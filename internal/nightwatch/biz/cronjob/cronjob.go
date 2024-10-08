package cronjob

//go:generate mockgen -destination mock_cronjob.go -package cronjob github.com/superproj/onex/internal/nightwatch/biz/cronjob CronJobBiz

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/nightwatch/conversion"
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

	CronJobExpansion
}

// CronJobExpansion defines additional methods for cronjob operations.
type CronJobExpansion interface{}

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
	cronJobM, err := b.ds.CronJobs().Get(ctx, where.T(ctx).F("cronjob_id", rq.CronJobID))
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
		cronJobM.ConcurrencyPolicy = int32(*rq.ConcurrencyPolicy)
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
	if err := b.ds.CronJobs().Delete(ctx, where.T(ctx).F("cronjob_id", rq.CronJobIDs)); err != nil {
		return nil, err
	}

	return &nwv1.DeleteCronJobResponse{}, nil
}

// Get retrieves a cron job by its ID from the data store.
func (b *cronJobBiz) Get(ctx context.Context, rq *nwv1.GetCronJobRequest) (*nwv1.GetCronJobResponse, error) {
	cronJob, err := b.ds.CronJobs().Get(ctx, where.T(ctx).F("cronjob_id", rq.CronJobID))
	if err != nil {
		return nil, err
	}

	bizCronJob := conversion.ConvertToV1CronJob(cronJob)
	return &nwv1.GetCronJobResponse{CronJob: bizCronJob}, nil
}

// List retrieves all cron jobs from the data store.
func (b *cronJobBiz) List(ctx context.Context, rq *nwv1.ListCronJobRequest) (*nwv1.ListCronJobResponse, error) {
	count, cronJobList, err := b.ds.CronJobs().List(ctx, where.T(ctx).P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		return nil, err
	}

	cronJobs := make([]*nwv1.CronJob, len(cronJobList))
	for i, cronJob := range cronJobList {
		cronJobs[i] = conversion.ConvertToV1CronJob(cronJob)
	}

	return &nwv1.ListCronJobResponse{TotalCount: count, CronJobs: cronJobs}, nil
}
