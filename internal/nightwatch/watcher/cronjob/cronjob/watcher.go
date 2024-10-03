// Package cronjob is a watcher implement.
package cronjob

import (
	"context"
	"fmt"
	"strings"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/store"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
	"github.com/superproj/onex/pkg/watch/manager"
	"github.com/superproj/onex/pkg/watch/registry"
)

var _ registry.Watcher = (*Watcher)(nil)

const CronJobMPrefix = "nightwatch_cronjob/"

// watcher implement.
type Watcher struct {
	store store.IStore
	jm    *manager.JobManager
}

type saveJob struct {
	ctx     context.Context
	store   store.IStore
	cronJob *model.CronJobM
}

func (j saveJob) Run() {
	count, _, err := j.store.Jobs().List(j.ctx, where.F("cronjob_id", j.cronJob.CronJobID))
	if err != nil {
		return
	}
	if count >= known.MaxJobsPerCronJob {
		return
	}

	// To prevent primary key conflicts.
	job := j.cronJob.JobTemplate
	job.ID = 0
	job.CronJobID = &j.cronJob.CronJobID
	job.UserID = j.cronJob.UserID
	job.Scope = j.cronJob.Scope
	job.Name = fmt.Sprintf("job-for-%s", j.cronJob.Name)
	if err := j.store.Jobs().Create(j.ctx, job); err != nil {
		log.Errorw(err, "Failed to create job")
		return
	}
}

// Run runs the watcher.
func (w *Watcher) Run() {
	ctx := context.Background()
	_, cronjobs, err := w.store.CronJobs().List(ctx, where.F("suspend", known.JobNonSuspended))
	if err != nil {
		return
	}

	w.RemoveNonExistentCronJobs(cronjobs)

	for _, cronjob := range cronjobs {
		jobName := cronJobName(cronjob.CronJobID)
		ctx = log.WithContext(ctx, "cronjob_id", cronjob.CronJobID)

		if cronjob.JobTemplate == nil {
			continue
		}

		if w.jm.JobExists(jobName) {
			continue
		}

		w.jm.AddJob(jobName, cronjob.Schedule, saveJob{store: w.store, ctx: ctx, cronJob: cronjob})
	}
}

// RemoveNonExistentCronJobs removes Cron jobs from the scheduler that no longer exist.
func (w *Watcher) RemoveNonExistentCronJobs(cronjobs []*model.CronJobM) {
	validCronJobIDs := make(map[string]struct{}, len(cronjobs))
	for _, cronjob := range cronjobs {
		validCronJobIDs[cronJobName(cronjob.CronJobID)] = struct{}{}
	}

	for jobName := range w.jm.GetJobs() {
		// Note: Do not delete CronJobs that are not walle_cronjobs here.
		if !isCronJobMName(jobName) {
			continue
		}

		if _, exists := validCronJobIDs[jobName]; exists {
			continue
		}
		_ = w.jm.RemoveJob(jobName)
	}
}

// Spec is parsed using the time zone of task Cron instance as the default.
func (w *Watcher) Spec() string {
	return "@every 1s"
}

// SetStore sets the persistence store for the Watcher.
func (w *Watcher) SetStore(store store.IStore) {
	w.store = store
}

// SetJobManager sets the JobManager for the Watcher.
func (w *Watcher) SetJobManager(jm *manager.JobManager) {
	w.jm = jm
}

func cronJobName(cronJobID string) string {
	return fmt.Sprintf("%s%s", CronJobMPrefix, cronJobID)
}

func isCronJobMName(jobName string) bool {
	return strings.HasPrefix(jobName, CronJobMPrefix)
}

func init() {
	registry.Register("cronjob", &Watcher{})
}
