package llmtrain

import (
	"github.com/looplab/fsm"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	known "github.com/superproj/onex/internal/pkg/known/nightwatch"
	fsmutil "github.com/superproj/onex/internal/pkg/util/fsm"
)

// StateMachine represents a finite state machine for managing daily estimation jobs.
type StateMachine struct {
	Watcher *Watcher
	Job     *model.JobM
	FSM     *fsm.FSM
}

// NewStateMachine initializes a new StateMachine with the given initial state, watcher, and job.
// It configures the FSM with defined events and their corresponding state transitions,
// as well as callbacks for entering specific states.
func NewStateMachine(initial string, watcher *Watcher, job *model.JobM) *StateMachine {
	sm := &StateMachine{Watcher: watcher, Job: job}

	sm.FSM = fsm.NewFSM(
		initial,
		fsm.Events{
			// Define state transitions for the daily estimation process.
			{Name: known.LLMTrainPending, Src: []string{known.LLMTrainPending}, Dst: known.LLMTrainDownloading},
			{Name: known.LLMTrainDownloading, Src: []string{known.LLMTrainDownloading}, Dst: known.LLMTrainDownloaded},
			{Name: known.LLMTrainDownloaded, Src: []string{known.LLMTrainDownloaded}, Dst: known.LLMTrainEmbedding},
			{Name: known.LLMTrainEmbedding, Src: []string{known.LLMTrainEmbedding}, Dst: known.LLMTrainEmbedded},
			{Name: known.LLMTrainEmbedded, Src: []string{known.LLMTrainEmbedded}, Dst: known.LLMTrainTraining},
			{Name: known.LLMTrainTraining, Src: []string{known.LLMTrainTraining}, Dst: known.LLMTrainTrained},
			{Name: known.LLMTrainTrained, Src: []string{known.LLMTrainTrained}, Dst: known.LLMTrainSucceeded},
		},
		fsm.Callbacks{
			// enter_state 先于 enter_xxx 执行。此时：event=Pending, current=Downloading
			"enter_state": fsmutil.WrapEvent(sm.EnterState),
			// 此时 event = Downloading，current = Downloaded
			"enter_" + known.LLMTrainDownloaded: fsmutil.WrapEvent(sm.Download),
			"enter_" + known.LLMTrainEmbedded:   fsmutil.WrapEvent(sm.Embedding),
			"enter_" + known.LLMTrainTrained:    fsmutil.WrapEvent(sm.Train),
		},
	)

	return sm
}
