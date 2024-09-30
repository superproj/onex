package block

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/nightwatch/biz"
	"github.com/superproj/onex/internal/pkg/core"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/api/zerrors"
)

// ProviderSet is the set of controller providers.
var ProviderSet = wire.NewSet(New)

// CronJobController handles requests related to CronJobs.
type CronJobController struct {
	biz biz.IBiz // Business logic interface
}

// New creates a new instance of CronJobController.
func New(biz biz.IBiz) *CronJobController {
	return &CronJobController{biz}
}

// Create handles the creation of a new CronJob.
func (ctrl *CronJobController) Create(c *gin.Context) {
	var rq nwv1.CreateCronJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.CronJobs().Create(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// Update handles the update of an existing CronJob.
func (ctrl *CronJobController) Update(c *gin.Context) {
	var rq nwv1.UpdateCronJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.CronJobs().Update(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete handles the deletion of a specified CronJob.
func (ctrl *CronJobController) Delete(c *gin.Context) {
	rq := nwv1.DeleteCronJobRequest{
		CronJobIDs: []string{c.Param("cronJobID")},
	}
	if err := ctrl.biz.CronJobs().Delete(c, &rq); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}

// Get retrieves a specified CronJob.
func (ctrl *CronJobController) Get(c *gin.Context) {
	rq := nwv1.GetCronJobRequest{
		CronJobID: c.Param("cronJobID"),
	}
	cronJob, err := ctrl.biz.CronJobs().Get(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, cronJob)
}

// List retrieves all CronJobs.
func (ctrl *CronJobController) List(c *gin.Context) {
	var rq nwv1.ListCronJobRequest
	if err := c.ShouldBindQuery(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.CronJobs().List(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}
