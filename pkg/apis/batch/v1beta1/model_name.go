// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package v1beta1

// OpenAPIModelName returns the OpenAPI model name for this type.
func (j *Job) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/batch/v1beta1.Job"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (j *JobList) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/batch/v1beta1.JobList"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (c *CronJob) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/batch/v1beta1.CronJob"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (c *CronJobList) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/batch/v1beta1.CronJobList"
}
