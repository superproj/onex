package llmtrain

import (
	"context"
	"fmt"
	"time"

	"github.com/looplab/fsm"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"k8s.io/utils/ptr"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	"github.com/superproj/onex/internal/nightwatch/path"
	"github.com/superproj/onex/internal/pkg/client/train"
	// onexembedder "github.com/superproj/onex/internal/pkg/embedding/embedder/onex"
	// "github.com/superproj/onex/internal/pkg/embedding/embedder/onex/text"
	// "github.com/superproj/onex/internal/pkg/embedding/embedder/onex/image"
	fakeminio "github.com/superproj/onex/internal/pkg/client/minio/fake"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	jobconditionsutil "github.com/superproj/onex/internal/pkg/util/jobconditions"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
	"github.com/superproj/onex/pkg/log"
)

// Download retrieves feedback data from VOC and saves it to TOS.
func (sm *StateMachine) Download(ctx context.Context, event *fsm.Event) error {
	// Set default job params.
	SetDefaultJobParams(sm.Job)

	// Skip the download if the operation has already been performed (idempotency check)
	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	time.Sleep(2 * time.Second)

	// Initialize job results if they are not already set
	if sm.Job.Results == nil || sm.Job.Results.Train == nil {
		sm.Job.Results = &model.JobResults{Train: &nwv1.TrainResults{}}
	}

	data, err := sm.Watcher.Minio.Read(ctx, fakeminio.FakeObjectName)
	if err != nil {
		return err
	}
	dataPath := path.Job.Path(sm.Job.JobID, path.JobDataName)
	if err := sm.Watcher.Minio.Write(ctx, dataPath, data); err != nil {
		return err
	}
	sm.Job.Results.Train.DataPath = &dataPath

	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))
	return nil
}

// Embedding embedding daily estimation data.
func (sm *StateMachine) Embedding(ctx context.Context, event *fsm.Event) error {
	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	results := sm.Job.Results.Train

	// Retrieve the downloaded feedback data from TOS
	docs, err := sm.Watcher.Minio.Read(ctx, *results.DataPath)
	if err != nil {
		return err
	}

	llm, err := ollama.New(ollama.WithModel("llama3"))
	if err != nil {
		log.Errorw(err, "Failed to new ollama client")
		return err
	}

	// Example of calling a handwritten embedder package.
	/*
		var typedEmbedder onexembedder.Embedder
		switch EmbedderType {
		case onexembedder.TextEmbeddingType:
			typedEmbedder = text.NewEmbedder(llm)
		case onexembedder.ImageEmbeddingType:
			typedEmbedder = image.NewEmbedder(llm)
		default:
		}
		// Create a new embedder with rate limiting
		embedder := onexembedder.NewEmbedder(typedEmbedder, onexembedder.WithRateLimiter(sm.Watcher.Limiter.Embedding))
		inputs := buildEmbedderInputs(EmbedderType, params)
		embs, err := embedder.Embedding(ctx, inputs)
		if err != nil {
			return err
		}
	*/

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Errorw(err, "Failed to NewEmbedder")
		return err
	}

	embs, err := embedder.EmbedDocuments(ctx, docs)
	if err != nil {
		log.Errorw(err, "Failed to EmbedDocuments")
		return err
	}
	embedingsStr := make([]string, len(embs))
	for i, emb := range embs {
		embedingsStr[i] = fmt.Sprintf("%v", emb)
	}

	// Update results and write the embedded data to TOS
	results.EmbeddedDataPath = ptr.To(path.Job.Path(sm.Job.JobID, path.JobEmbeddedDataName))
	if err := sm.Watcher.Minio.Write(ctx, *results.EmbeddedDataPath, embedingsStr); err != nil {
		return err
	}

	results.TaskID = nil
	jobconditionsutil.Delete(sm.Job.Conditions, known.LLMTrainTrained)

	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))
	return nil
}

func (sm *StateMachine) Train(ctx context.Context, event *fsm.Event) error {
	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	results := sm.Job.Results.Train
	_ = sm.Watcher.Limiter.Train.Take() // Rate limiting

	// Function to create the Arthur training task
	createTrainTaskFunc := func() error {
		resultPath := path.Job.Path(sm.Job.JobID, path.JobResultName)
		taskID, err := sm.Watcher.Train.CreateTask(ctx, *results.EmbeddedDataPath, resultPath)
		if err != nil {
			return err
		}
		results.TaskID = &taskID
		results.ResultPath = &resultPath
		return nil
	}

	// Create task if it hasn't been created yet
	if results.TaskID == nil {
		if err := createTrainTaskFunc(); err != nil {
			return err
		}
	}

	status, err := sm.Watcher.Train.GetTaskStatus(ctx, *results.TaskID)
	if err != nil {
		log.Errorw(err, "Failed to GetTask")
		return err
	}

	if status != train.StatusCompleted {
		log.Infow("Train task has not been completed", "status", status)
		event.FSM.SetState(event.Event)
		return nil
	}

	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))
	return nil
}

// EnterState handles the state transition of the state machine
// and updates the Job's status and conditions based on the event.
func (sm *StateMachine) EnterState(ctx context.Context, event *fsm.Event) error {
	sm.Job.Status = event.FSM.Current()

	// Record the start time of the job
	if sm.Job.Status == known.LLMTrainDownloading {
		sm.Job.StartedAt = time.Now()
	}

	// Unified handling logic for Job failure
	if event.Err != nil || isJobTimeout(sm.Job) {
		sm.Job.Status = known.LLMTrainFailed
		sm.Job.EndedAt = time.Now()

		var cond *nwv1.JobCondition
		if isJobTimeout(sm.Job) {
			log.Infow("LLM train task timeout")
			cond = jobconditionsutil.FalseCondition(event.FSM.Current(), fmt.Sprintf("LLM train task exceeded timeout seconds"))
		} else {
			cond = jobconditionsutil.FalseCondition(event.FSM.Current(), event.Err.Error())
		}

		sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, cond)
	}

	if err := sm.Watcher.Store.Jobs().Update(ctx, sm.Job); err != nil {
		return err
	}

	//sm.MustMetrics(ctx, event)

	return nil
}

/*
func (sm *StateMachine) MustMetrics(ctx context.Context, event *fsm.Event) {
	// Record metrics only on success or failure.
	if !stringsutil.StringIn(sm.Job.Status, []string{known.DailyEstimationSucceeded, known.DailyEstimationFailed}) {
		return
	}

	tags := []metrics.T{
		{Name: "env", Value: env.Env()},
		{Name: "tenant", Value: sm.Job.Tenant},
		{Name: "job_id", Value: strconv.FormatInt(sm.Job.ID, 10)},
		{Name: "model_id", Value: strconv.FormatInt(*sm.Job.Params.DailyEstimation.ModelID, 10)},
		{Name: "max_feedback_nums", Value: strconv.FormatInt(gptr.Indirect(sm.Job.Params.DailyEstimation.MaxFeedbackNums), 10)},
		{Name: "status", Value: sm.Job.Status},
		{Name: "cost", Value: strconv.Itoa(int(sm.Job.EndedAt.Sub(sm.Job.StartedAt).Seconds()))},
	}

	if err := sm.Watcher.Metric.WithTags(tags...).Emit(metrics.IncrCounter(1)); err != nil {
		log.Errorw(err, "Failed to emit metrics")
	}
}
*/
