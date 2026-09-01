// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/onexstack/onex/internal/controller/job/job/providers/registry"
	jobutil "github.com/onexstack/onex/internal/pkg/util/job"
	"github.com/onexstack/onex/internal/pkg/util/patch"
	"github.com/onexstack/onex/internal/pkg/util/predicates"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

const (
	controllerName = "job-controller"

	// providerPollInterval is how often a running job re-checks provider status.
	providerPollInterval = 10 * time.Second
)

var controllerKind = v1beta1.SchemeGroupVersion.WithKind("Job")

// Reconciler reconciles a Job object, driving its lifecycle
// (Pending -> Running -> Succeeded/Failed).
type Reconciler struct {
	client client.Client

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string

	recorder     record.EventRecorder
	jobControl   jobControlInterface
	provider     registry.Provider
	expectations *ControllerExpectations
	backoff      *backoffStore

	// clock returns the current time and provides fake-clock support for tests.
	clock clock.Clock
}

// SetupWithManager sets up the controller with the manager.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Job{}).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.All(
			ctrl.LoggerFrom(ctx),
			predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
		))

	r.client = mgr.GetClient()
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.jobControl = &realJobControl{client: mgr.GetClient()}
	r.provider = newRealProviderControl(mgr.GetClient())
	r.expectations = NewControllerExpectations()
	r.backoff = newBackoffStore()
	r.clock = clock.RealClock{}

	return builder.Complete(r)
}

// Reconcile reads the state of a Job and drives it towards its desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	job := &v1beta1.Job{}
	if err := r.client.Get(ctx, req.NamespacedName, job); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	helper, err := patch.NewHelper(job, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := helper.Patch(ctx, job); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	if !job.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.sync(ctx, job)
}

// sync reconciles a single Job, implementing the same write-then-read-consistency
// gate, suspend/active-deadline handling and condition management as the
// upstream Job controller.
func (r *Reconciler) sync(ctx context.Context, job *v1beta1.Job) (ctrl.Result, error) {
	key := jobKey(job)

	// Write-then-read consistency gate: don't act until the informer has observed
	// the controller's own writes.
	if !r.expectations.SatisfiedExpectations(key) {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// A suspended job is not dispatched and any running workload is stopped.
	if job.Spec.Suspend {
		setJobCondition(job, v1beta1.JobSuspended, corev1.ConditionTrue, "JobSuspended", "Job is suspended")
		job.Status.Phase = v1beta1.JobPhasePending
		job.Status.StartedAt = nil
		if err := r.provider.Stop(ctx, job); err != nil {
			r.recorder.Eventf(job, corev1.EventTypeWarning, "StopFailed", "failed to stop suspended job: %v", err)
			return ctrl.Result{}, err
		}
		r.backoff.reset(key)
		r.expectations.DeleteExpectations(key)
		return ctrl.Result{}, nil
	}
	removeJobCondition(job, v1beta1.JobSuspended)

	// Nothing to do once the job has reached a terminal state.
	if jobutil.IsJobFinished(job) {
		return ctrl.Result{}, nil
	}

	// Enforce the active deadline.
	if job.Spec.ActiveDeadlineSeconds != nil && job.Status.StartedAt != nil {
		deadline := job.Status.StartedAt.Add(time.Duration(*job.Spec.ActiveDeadlineSeconds) * time.Second)
		if r.clock.Now().After(deadline) {
			message := fmt.Sprintf("job exceeded its activeDeadlineSeconds of %d", *job.Spec.ActiveDeadlineSeconds)
			setJobCondition(job, v1beta1.JobFailed, corev1.ConditionTrue, v1beta1.JobReasonDeadlineExceeded, message)
			job.Status.Phase = v1beta1.JobPhaseFailed
			job.Status.ErrorMessage = ptr.To(message)
			now := metav1.Now()
			job.Status.EndedAt = &now
			r.recorder.Event(job, corev1.EventTypeWarning, "DeadlineExceeded", message)
			_ = r.provider.Stop(ctx, job)
			r.expectations.DeleteExpectations(key)
			return ctrl.Result{}, nil
		}
	}

	// Start the job if it has not started yet.
	if job.Status.StartedAt == nil {
		if remaining := r.backoff.getRemainingTime(key, r.clock.Now()); remaining > 0 {
			return ctrl.Result{RequeueAfter: remaining}, nil
		}
		r.expectations.SetExpectations(key, 1, 0)
		if err := r.provider.Start(ctx, job); err != nil {
			r.expectations.CreationObserved(key)
			r.backoff.recordFailure(key, r.clock.Now())
			r.recorder.Eventf(job, corev1.EventTypeWarning, "StartFailed", "failed to start job: %v", err)
			return ctrl.Result{}, err
		}
		r.expectations.CreationObserved(key)
		now := metav1.Now()
		job.Status.StartedAt = &now
		job.Status.Phase = v1beta1.JobPhaseRunning
		r.backoff.reset(key)
		r.recorder.Event(job, corev1.EventTypeNormal, "SuccessfulStart", "Started job")
		return ctrl.Result{RequeueAfter: providerPollInterval}, nil
	}

	// Running: keep provider-managed children claimed and poll completion.
	if err := r.reconcileChildren(ctx, job); err != nil {
		return ctrl.Result{}, err
	}

	status, err := r.provider.Status(ctx, job)
	if err != nil {
		return ctrl.Result{}, err
	}
	switch {
	case status.Succeeded:
		job.Status.Phase = v1beta1.JobPhaseSucceeded
		setJobCondition(job, v1beta1.JobComplete, corev1.ConditionTrue, "JobComplete", status.Message)
		now := metav1.Now()
		job.Status.EndedAt = &now
		r.recorder.Event(job, corev1.EventTypeNormal, "Completed", "Job completed")
		r.expectations.DeleteExpectations(key)
		return ctrl.Result{}, nil
	case status.Finished && !status.Succeeded:
		job.Status.Phase = v1beta1.JobPhaseFailed
		setJobCondition(job, v1beta1.JobFailed, corev1.ConditionTrue, "JobFailed", status.Message)
		job.Status.ErrorMessage = ptr.To(status.Message)
		now := metav1.Now()
		job.Status.EndedAt = &now
		r.recorder.Event(job, corev1.EventTypeWarning, "Failed", status.Message)
		r.expectations.DeleteExpectations(key)
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{RequeueAfter: providerPollInterval}, nil
	}
}

// reconcileChildren adopts or releases provider-managed child objects via their
// controller reference so ownership stays consistent with the Job.
func (r *Reconciler) reconcileChildren(ctx context.Context, job *v1beta1.Job) error {
	children, err := r.provider.Children(ctx, job)
	if err != nil {
		return err
	}
	refManager := NewControllerRefManager(r.client, job, controllerKind)
	for _, child := range children {
		if _, err := refManager.ClaimObject(ctx, child, func(metav1.Object) bool { return true }); err != nil {
			return err
		}
	}
	return nil
}

func jobKey(job *v1beta1.Job) string {
	return job.Namespace + "/" + job.Name
}

func findJobCondition(conditions []v1beta1.JobCondition, t v1beta1.JobConditionType) *v1beta1.JobCondition {
	for i := range conditions {
		if conditions[i].Type == t {
			return &conditions[i]
		}
	}
	return nil
}

func setJobCondition(job *v1beta1.Job, conditionType v1beta1.JobConditionType, status corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	existing := findJobCondition(job.Status.Conditions, conditionType)
	if existing == nil {
		job.Status.Conditions = append(job.Status.Conditions, v1beta1.JobCondition{
			Type:               conditionType,
			Status:             status,
			Reason:             reason,
			Message:            message,
			LastProbeTime:      now,
			LastTransitionTime: now,
		})
		return
	}
	if existing.Status != status {
		existing.LastTransitionTime = now
	}
	existing.Status = status
	existing.Reason = reason
	existing.Message = message
	existing.LastProbeTime = now
}

func removeJobCondition(job *v1beta1.Job, conditionType v1beta1.JobConditionType) {
	conditions := make([]v1beta1.JobCondition, 0, len(job.Status.Conditions))
	for _, c := range job.Status.Conditions {
		if c.Type != conditionType {
			conditions = append(conditions, c)
		}
	}
	job.Status.Conditions = conditions
}
