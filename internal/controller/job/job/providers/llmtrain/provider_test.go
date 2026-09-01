// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package llmtrain

import (
	"context"
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	fakeminio "github.com/onexstack/onex/internal/pkg/client/minio/fake"
	"github.com/onexstack/onex/internal/pkg/client/train"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

// instantTrainer completes tasks immediately, so training advances in a single
// Status call (the real TrainManager simulates a 20s processing delay).
type instantTrainer struct{}

func (instantTrainer) CreateTask(context.Context, string, string) (string, error) {
	return "task-instant", nil
}

func (instantTrainer) GetTaskStatus(context.Context, string) (string, error) {
	return train.StatusCompleted, nil
}

// failingMinio fails every Read to exercise the terminal-error path.
type failingMinio struct{}

func (failingMinio) Read(context.Context, string) ([]string, error) {
	return nil, errors.New("read failed")
}

func (failingMinio) Write(context.Context, string, []string) error {
	return nil
}

func newTestProvider() *Provider {
	m, _ := fakeminio.NewFakeMinioClient("test")
	return newProvider(m, instantTrainer{}, deterministicEmbedder{})
}

func newJob() *v1beta1.Job {
	return &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "llm", Namespace: "default"},
		Spec:       v1beta1.JobSpec{Type: Type},
	}
}

func mustStatus(t *testing.T, job *v1beta1.Job) *Status {
	t.Helper()
	st, err := decodeStatus(job)
	if err != nil {
		t.Fatalf("decode status: %v", err)
	}
	return st
}

func TestPipelinePhases(t *testing.T) {
	p := newTestProvider()
	job := newJob()
	ctx := context.Background()

	if err := p.Start(ctx, job); err != nil {
		t.Fatalf("start: %v", err)
	}
	if got := mustStatus(t, job).Phase; got != phasePending {
		t.Fatalf("expected Pending after start, got %q", got)
	}

	want := []string{
		phaseDownloading, phaseDownloaded, phaseEmbedding, phaseEmbedded,
		phaseTraining, phaseTrained, phaseSucceeded,
	}
	got := make([]string, 0, len(want))
	for i := 0; i < len(want); i++ {
		if _, err := p.Status(ctx, job); err != nil {
			t.Fatalf("status #%d: %v", i, err)
		}
		got = append(got, mustStatus(t, job).Phase)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: expected phase %q, got %q (all: %v)", i, want[i], got[i], got)
		}
	}
}

func TestPipelineSucceeds(t *testing.T) {
	p := newTestProvider()
	job := newJob()
	ctx := context.Background()

	if err := p.Start(ctx, job); err != nil {
		t.Fatalf("start: %v", err)
	}

	var finished bool
	for i := 0; i < 20; i++ {
		st, err := p.Status(ctx, job)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		if st.Finished {
			if !st.Succeeded {
				t.Fatalf("expected success, got %+v", st)
			}
			finished = true
			break
		}
	}
	if !finished {
		t.Fatal("pipeline did not finish")
	}

	s := mustStatus(t, job)
	if s.DataPath == nil || s.EmbeddedDataPath == nil || s.TaskID == nil || s.ResultPath == nil {
		t.Fatalf("expected pipeline artifacts, got %+v", s)
	}
}

func TestPipelineFailsOnError(t *testing.T) {
	p := newProvider(failingMinio{}, instantTrainer{}, deterministicEmbedder{})
	job := newJob()
	ctx := context.Background()

	if err := p.Start(ctx, job); err != nil {
		t.Fatalf("start: %v", err)
	}

	var finished bool
	for i := 0; i < 20; i++ {
		st, err := p.Status(ctx, job)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		if st.Finished {
			if st.Succeeded {
				t.Fatalf("expected failure, got %+v", st)
			}
			finished = true
			break
		}
	}
	if !finished {
		t.Fatal("pipeline did not finish")
	}
	if got := mustStatus(t, job).Phase; got != phaseFailed {
		t.Fatalf("expected phase Failed, got %q", got)
	}
}

func TestPipelineTimesOut(t *testing.T) {
	p := newTestProvider()
	job := newJob()
	job.Spec.ProviderSpec = &runtime.RawExtension{Raw: []byte(`{"jobTimeoutSec":1}`)}
	ctx := context.Background()

	if err := p.Start(ctx, job); err != nil {
		t.Fatalf("start: %v", err)
	}
	startedAt := metav1.NewTime(time.Now().Add(-time.Minute))
	job.Status.StartedAt = &startedAt

	st, err := p.Status(ctx, job)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !st.Finished || st.Succeeded {
		t.Fatalf("expected finished+failed (timeout), got %+v", st)
	}
	if got := mustStatus(t, job).Phase; got != phaseFailed {
		t.Fatalf("expected phase Failed, got %q", got)
	}
}
