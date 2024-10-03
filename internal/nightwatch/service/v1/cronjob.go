package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/superproj/onex/internal/pkg/core"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/api/zerrors"
)

// CreateCronJob handles the creation of a new CronJob.
func (s *NightWatchService) CreateCronJob(c *gin.Context) {
	var rq nwv1.CreateCronJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	if err := s.valid.ValidateCreateCronJobRequest(c, &rq); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	resp, err := s.biz.CronJobs().Create(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// UpdateCronJob handles the update of an existing CronJob.
func (s *NightWatchService) UpdateCronJob(c *gin.Context) {
	var rq nwv1.UpdateCronJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}
	rq.CronJobID = c.Param("cronJobID")

	resp, err := s.biz.CronJobs().Update(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// DeleteCronJob handles the deletion of a specified CronJob.
func (s *NightWatchService) DeleteCronJob(c *gin.Context) {
	rq := nwv1.DeleteCronJobRequest{
		CronJobIDs: []string{c.Param("cronJobID")},
	}
	resp, err := s.biz.CronJobs().Delete(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// GetCronJob retrieves a specified CronJob.
func (s *NightWatchService) GetCronJob(c *gin.Context) {
	rq := nwv1.GetCronJobRequest{
		CronJobID: c.Param("cronJobID"),
	}
	cronJob, err := s.biz.CronJobs().Get(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, cronJob)
}

// ListCronJob retrieves all CronJobs.
func (s *NightWatchService) ListCronJob(c *gin.Context) {
	var rq nwv1.ListCronJobRequest
	if err := c.ShouldBindQuery(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := s.biz.CronJobs().List(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}
