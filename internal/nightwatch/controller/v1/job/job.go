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

// JobController handles requests related to Jobs.
type JobController struct {
	biz biz.IBiz // Business logic interface
}

// New creates a new instance of JobController.
func New(biz biz.IBiz) *JobController {
	return &JobController{biz}
}

// Create handles the creation of a new Job.
func (ctrl *JobController) Create(c *gin.Context) {
	var rq nwv1.CreateJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.Jobs().Create(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// Update handles the update of an existing Job.
func (ctrl *JobController) Update(c *gin.Context) {
	var rq nwv1.UpdateJobRequest
	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.Jobs().Update(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete handles the deletion of a specified Job.
func (ctrl *JobController) Delete(c *gin.Context) {
	rq := nwv1.DeleteJobRequest{
		JobIDs: []string{c.Param("jobID")},
	}
	if err := ctrl.biz.Jobs().Delete(c, &rq); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}

// Get retrieves a specified Job.
func (ctrl *JobController) Get(c *gin.Context) {
	rq := nwv1.GetJobRequest{
		JobID: c.Param("jobID"),
	}
	job, err := ctrl.biz.Jobs().Get(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, job)
}

// List retrieves all Jobs.
func (ctrl *JobController) List(c *gin.Context) {
	var rq nwv1.ListJobRequest
	if err := c.ShouldBindQuery(&rq); err != nil {
		core.WriteResponse(c, zerrors.ErrorInvalidParameter(err.Error()), nil)
		return
	}

	resp, err := ctrl.biz.Jobs().List(c, &rq)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}
