package conversion

import (
	"github.com/jinzhu/copier"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
)

// ConvertToV1CronJob converts a CronJob model to its v1 representation.
func ConvertToV1CronJob(cronJob *model.CronJobM) *nwv1.CronJob {
	var bizCronJob nwv1.CronJob
	_ = copier.Copy(&bizCronJob, cronJob) // Copy model data to the response.

	var job nwv1.Job
	_ = copier.Copy(&job, cronJob.JobTemplate)
	bizCronJob.JobTemplate = &job
	bizCronJob.CreatedAt = timestamppb.New(cronJob.CreatedAt)
	bizCronJob.UpdatedAt = timestamppb.New(cronJob.UpdatedAt)
	return &bizCronJob
}

// ConvertToV1Job converts a Job model to its v1 representation.
func ConvertToV1Job(job *model.JobM) *nwv1.Job {
	var bizJob nwv1.Job
	_ = copier.Copy(&bizJob, job) // Copy model data to the response.

	bizJob.CreatedAt = timestamppb.New(job.CreatedAt)
	bizJob.UpdatedAt = timestamppb.New(job.UpdatedAt)
	return &bizJob
}
