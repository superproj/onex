// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	testingclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"

	"github.com/onexstack/onex/internal/controller/job/job/providers/registry"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

func newTestReconciler(provider registry.Provider, now time.Time) *Reconciler {
	return &Reconciler{
		provider:     provider,
		recorder:     record.NewFakeRecorder(100),
		expectations: NewControllerExpectations(),
		backoff:      newBackoffStore(),
		clock:        testingclock.NewFakeClock(now),
	}
}

func testJob(name string) *v1beta1.Job {
	return &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
	}
}

func TestSyncStartsJob(t *testing.T) {
	now := time.Now()
	r := newTestReconciler(&fakeProviderControl{}, now)
	job := testJob("job1")

	if _, err := r.sync(context.Background(), job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status.StartedAt == nil {
		t.Fatal("expected StartedAt to be set")
	}
	if job.Status.Phase != v1beta1.JobPhaseRunning {
		t.Fatalf("expected phase Running, got %q", job.Status.Phase)
	}
}

func TestSyncSuspendsJob(t *testing.T) {
	now := time.Now()
	r := newTestReconciler(&fakeProviderControl{}, now)
	job := testJob("job1")
	job.Spec.Suspend = true
	startedAt := metav1.NewTime(now.Add(-time.Hour))
	job.Status.StartedAt = &startedAt

	if _, err := r.sync(context.Background(), job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status.StartedAt != nil {
		t.Fatal("expected StartedAt to be cleared while suspended")
	}
	if !hasCondition(job, v1beta1.JobSuspended, corev1.ConditionTrue) {
		t.Fatal("expected Suspended condition True")
	}
}

func TestSyncActiveDeadline(t *testing.T) {
	now := time.Now()
	r := newTestReconciler(&fakeProviderControl{}, now)
	job := testJob("job1")
	startedAt := metav1.NewTime(now.Add(-2 * time.Hour))
	job.Status.StartedAt = &startedAt
	job.Spec.ActiveDeadlineSeconds = ptr.To(int64(60))

	if _, err := r.sync(context.Background(), job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasCondition(job, v1beta1.JobFailed, corev1.ConditionTrue) {
		t.Fatal("expected Failed condition True (deadline exceeded)")
	}
	if job.Status.Phase != v1beta1.JobPhaseFailed {
		t.Fatalf("expected phase Failed, got %q", job.Status.Phase)
	}
}

func TestSyncCompletesJob(t *testing.T) {
	now := time.Now()
	provider := &fakeProviderControl{StatusResp: registry.Status{Finished: true, Succeeded: true}}
	r := newTestReconciler(provider, now)
	job := testJob("job1")
	startedAt := metav1.NewTime(now.Add(-time.Minute))
	job.Status.StartedAt = &startedAt

	if _, err := r.sync(context.Background(), job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasCondition(job, v1beta1.JobComplete, corev1.ConditionTrue) {
		t.Fatal("expected Complete condition True")
	}
	if job.Status.Phase != v1beta1.JobPhaseSucceeded {
		t.Fatalf("expected phase Succeeded, got %q", job.Status.Phase)
	}
}

func TestSyncStartFailureAppliesBackoff(t *testing.T) {
	now := time.Now()
	provider := &fakeProviderControl{StartErr: fmt.Errorf("boom")}
	r := newTestReconciler(provider, now)
	job := testJob("job1")

	if _, err := r.sync(context.Background(), job); err == nil {
		t.Fatal("expected error from failed start")
	}
	if job.Status.StartedAt != nil {
		t.Fatal("expected StartedAt to remain nil on failed start")
	}
	if r.backoff.getRemainingTime(jobKey(job), now) <= 0 {
		t.Fatal("expected a positive backoff after a failed start")
	}
}

func TestExpectationsGate(t *testing.T) {
	e := NewControllerExpectations()
	key := "default/job1"

	if !e.SatisfiedExpectations(key) {
		t.Fatal("empty expectations should be satisfied")
	}

	e.SetExpectations(key, 1, 0)
	if e.SatisfiedExpectations(key) {
		t.Fatal("pending creation should block reconciliation")
	}

	e.CreationObserved(key)
	if !e.SatisfiedExpectations(key) {
		t.Fatal("observed creation should satisfy expectations")
	}
}

func hasCondition(job *v1beta1.Job, conditionType v1beta1.JobConditionType, status corev1.ConditionStatus) bool {
	c := findJobCondition(job.Status.Conditions, conditionType)
	return c != nil && c.Status == status
}
