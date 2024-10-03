// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/superproj/onex/internal/nightwatch/store"
	"github.com/superproj/onex/internal/pkg/known"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/api/zerrors"
)

// ProviderSet is validator providers.
var ProviderSet = wire.NewSet(New, wire.Bind(new(any), new(*Validator)))

// Validator struct implements the custom validator interface.
type Validator struct {
	ds store.IStore
}

// New creates and initializes a custom validator.
// It receives an instance of store.IStore interface as parameter ds
// and returns a new *Validator and an error.
func New(ds store.IStore) *Validator {
	return &Validator{ds: ds}
}

func (valid *Validator) ValidateCreateCronJobRequest(ctx *gin.Context, rq *nwv1.CreateCronJobRequest) error {
	if rq.CronJob == nil {
		return zerrors.ErrorInvalidParameter("cronJob cannot be empty")
	}

	if err := validateJob(ctx, rq.CronJob.JobTemplate); err != nil {
		return err
	}

	rq.CronJob.UserID = ctx.GetString(known.XUserID)
	return nil
}

func (valid *Validator) ValidateUpdateCronJobRequest(ctx *gin.Context, rq *nwv1.UpdateCronJobRequest) error {
	return nil
}

func (valid *Validator) ValidateDeleteCronJobRequest(ctx *gin.Context, rq *nwv1.DeleteCronJobRequest) error {
	return nil
}

func (valid *Validator) ValidateGetCronJobRequest(ctx *gin.Context, rq *nwv1.GetCronJobRequest) error {
	return nil
}

func (valid *Validator) ValidateListCronJobRequest(ctx *gin.Context, rq *nwv1.ListCronJobRequest) error {
	return nil
}

func (valid *Validator) ValidateCreateJobRequest(ctx *gin.Context, rq *nwv1.CreateJobRequest) error {
	if rq.Job == nil {
		return zerrors.ErrorInvalidParameter("job cannot be empty")
	}

	if err := validateJob(ctx, rq.Job); err != nil {
		return err
	}

	rq.Job.UserID = ctx.Value(known.XUserID).(string)
	return nil
}

func (valid *Validator) ValidateUpdateJobRequest(ctx *gin.Context, rq *nwv1.UpdateJobRequest) error {
	return nil
}

func (valid *Validator) ValidateDeleteJobRequest(ctx *gin.Context, rq *nwv1.DeleteJobRequest) error {
	return nil
}

func (valid *Validator) ValidateGetJobRequest(ctx *gin.Context, rq *nwv1.GetJobRequest) error {
	return nil
}

func (valid *Validator) ValidateListJobRequest(ctx *gin.Context, rq *nwv1.ListJobRequest) error {
	return nil
}
