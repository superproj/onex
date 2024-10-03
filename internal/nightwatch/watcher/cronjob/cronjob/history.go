// Package cronjob is a watcher implement.
package cronjob

import (
	"context"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/store"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
	"github.com/superproj/onex/pkg/watch/registry"
)

var _ registry.Watcher = (*Watcher)(nil)

// watcher implement.
type History struct {
	store store.IStore
}

// Run runs the watcher.
func (h *History) Run() {
	ctx := context.Background()
	_, cronjobs, err := h.store.CronJobs().List(ctx, where.F("suspend", known.JobNonSuspended))
	if err != nil {
		return
	}

	for _, cronjob := range cronjobs {
		h.retainRecords(ctx, known.JobSucceeded, cronjob.SuccessHistoryLimit)
		h.retainRecords(ctx, known.JobFailed, cronjob.FailedHistoryLimit)
	}
}

// Spec is parsed using the time zone of task Cron instance as the default.
func (h *History) Spec() string {
	return "@every 1s"
}

// SetStore sets the persistence store for the Watcher.
func (h *History) SetStore(store store.IStore) {
	h.store = store
}

func (h *History) retainRecords(ctx context.Context, status string, maxRecords int32) {
	_, jobs, err := h.store.Jobs().List(ctx, where.F("status", status))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list jobs")
		return
	}
	removedIDs := retainMaxElements(jobs, maxRecords)
	if err := h.store.Jobs().Delete(ctx, where.F("job_id", removedIDs)); err != nil {
		log.C(ctx).Errorw(err, "Failed to delete jobs")
	}
}

func retainMaxElements(jobs []*model.JobM, maxRecords int32) []string {
	all := make([]string, len(jobs))
	for i, job := range jobs {
		all[i] = job.JobID
	}

	if len(all) <= int(maxRecords) {
		return []string{}
	}

	return all[maxRecords:]
}

func init() {
	registry.Register("history", &Watcher{})
}
