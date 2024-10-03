package llmtrain

import (
	"context"

	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"go.uber.org/ratelimit"

	"github.com/superproj/onex/internal/nightwatch/store"
	"github.com/superproj/onex/internal/nightwatch/watcher"
	"github.com/superproj/onex/internal/pkg/client/minio"
	"github.com/superproj/onex/internal/pkg/client/train"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
	"github.com/superproj/onex/pkg/watch/registry"
)

// Ensure Watcher implements the registry.Watcher interface.
var _ registry.Watcher = (*Watcher)(nil)

// Limiter holds rate limiters for different operations.
type Limiter struct {
	Embedding ratelimit.Limiter
	Train     ratelimit.Limiter
}

// Watcher monitors and processes daily estimation jobs.
type Watcher struct {
	//Metric          metrics.Metric
	Train *train.TrainManager
	Minio minio.IMinio
	Store store.IStore

	// Maximum number of concurrent workers.
	MaxWorkers int64
	// Rate limiters for operations.
	Limiter Limiter
}

// Run executes the watcher logic to process jobs.
func (w *Watcher) Run() {
	// Define the phases that the watcher can handle.
	runablePhase := []string{
		known.LLMTrainPending,
		known.LLMTrainDownloading,
		known.LLMTrainDownloaded,
		known.LLMTrainEmbedding,
		known.LLMTrainEmbedded,
		known.LLMTrainTraining,
		known.LLMTrainTrained,
	}

	_, jobs, err := w.Store.Jobs().List(context.Background(), where.F(
		"scope", known.LLMJobScope,
		"watcher", known.LLMTrainWatcher,
		"status", runablePhase,
		"suspend", known.JobNonSuspended,
	))
	if err != nil {
		log.Errorw(err, "Failed to get runnable jobs")
		return
	}

	wp := workerpool.New(int(w.MaxWorkers))
	for _, job := range jobs {
		ctx := log.WithContext(context.Background(), "run_id", uuid.New().String(), "watcher", job.Watcher, "job_id", job.JobID)
		log.C(ctx).Infow("Start to train llm model")

		wp.Submit(func() {
			sm := NewStateMachine(job.Status, w, job)
			if err := sm.FSM.Event(ctx, job.Status); err != nil {
				return
			}
		})
	}

	wp.StopWait()
}

// Spec returns the cron job specification for scheduling.
func (w *Watcher) Spec() string {
	return "@every 1s"
}

// SetAggregateConfig configures the watcher with the provided aggregate configuration.
func (w *Watcher) SetAggregateConfig(config *watcher.AggregateConfig) {
	w.Train = train.NewTrainManager()
	w.Minio = config.Minio
	w.Store = config.Store
	w.Limiter = Limiter{
		Embedding: ratelimit.New(known.LLMTrainEmbeddingQPS),
		Train:     ratelimit.New(known.LLMTrainEvaluateQPS),
	}
}

// SetMaxWorkers sets the maximum number of concurrent workers for the watcher.
func (w *Watcher) SetMaxWorkers(maxWorkers int64) {
	// Since the daily accuracy evaluation task needs to call the embedding model, a custom
	// maxWorkers setting is used here to reduce the pressure on the embedding model.
	w.MaxWorkers = known.LLMTrainMaxWorkers
}

func init() {
	registry.Register(known.LLMTrainWatcher, &Watcher{})
}
