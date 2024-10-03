package model

import (
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/zid"
)

// AfterCreate runs after creating a CronJobM database record and updates the JobID field.
func (cj *CronJobM) AfterCreate(tx *gorm.DB) (err error) {
	cj.CronJobID = zid.CronJob.New(uint64(cj.ID)) // Generate and set a new cronjob ID.

	return tx.Save(cj).Error // Save the updated cronjob record to the database.
}

// AfterCreate runs after creating a JobM database record and updates the JobID field.
func (j *JobM) AfterCreate(tx *gorm.DB) (err error) {
	j.JobID = zid.Job.New(uint64(j.ID)) // Generate and set a new job ID.

	return tx.Save(j).Error // Save the updated job record to the database.
}
