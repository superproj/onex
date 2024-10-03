// Package statesync is a watcher implement.
package statesync

import (
	"context"

	"github.com/gammazero/workerpool"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/store"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
	"github.com/superproj/onex/pkg/watch/registry"
)

var _ registry.Watcher = (*Watcher)(nil)

// watcher implement.
type Watcher struct {
	store      store.IStore
	maxWorkers int64
}

// Run runs the watcher.
func (w *Watcher) Run() {
	ctx := context.Background()

	// Query active cron jobs that are not suspended
	_, cronjobs, err := w.store.CronJobs().List(ctx, where.F("suspend", 0))
	if err != nil {
		return
	}

	wp := workerpool.New(int(w.maxWorkers))
	for _, cronjob := range cronjobs {
		wp.Submit(func() {
			ctx = log.WithContext(ctx, "cronjob_id", cronjob.CronJobID)
			_, jobs, err := w.store.Jobs().List(ctx, where.F("cronjob_id", cronjob.CronJobID))
			if err != nil || len(jobs) == 0 {
				return
			}

			active := make([]int64, 0)
			var lastSuccessJob *model.JobM
			var lastScheduleJob *model.JobM

			// Process each job related to the cron job
			for _, job := range jobs {
				if job.Status == known.JobRunning {
					active = append(active, job.ID)
				}

				// QueryJobM orders by ID in descending order, with the first being the most recent.
				if lastSuccessJob == nil && job.Status == known.JobSucceeded {
					lastSuccessJob = job
				}

				if lastScheduleJob == nil && !job.StartedAt.IsZero() {
					lastScheduleJob = job
				}
			}

			cronjob.Status = &model.CronJobStatus{Active: active, LastJobID: jobs[0].JobID}

			if lastSuccessJob != nil {
				cronjob.Status.LastSuccessfulTime = lastSuccessJob.EndedAt.Unix()
			}

			if lastSuccessJob != nil {
				cronjob.Status.LastScheduleTime = lastSuccessJob.StartedAt.Unix()
			}

			_ = w.store.CronJobs().Update(ctx, cronjob)
		})
	}

	wp.StopWait()
}

// Spec is parsed using the time zone of task Cron instance as the default.
func (w *Watcher) Spec() string {
	return "@every 1s"
}

// SetStore sets the persistence store for the Watcher.
func (w *Watcher) SetStore(store store.IStore) {
	w.store = store
}

func (w *Watcher) SetMaxWorkers(maxWorkers int64) {
	w.maxWorkers = maxWorkers
}

func init() {
	registry.Register("statesync", &Watcher{})
}
