// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package cronjob

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	coreutil "github.com/onexstack/onex/internal/pkg/util/core"
	jobutil "github.com/onexstack/onex/internal/pkg/util/job"
	"github.com/onexstack/onex/internal/pkg/util/patch"
	"github.com/onexstack/onex/internal/pkg/util/predicates"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"

	//"github.com/onexstack/onex/pkg/record"
	"github.com/onexstack/onex/third_party/protobuf/k8s.io/apimachinery/pkg/api/errors"
)

const controllerName = "cronjob-controller"

var (
	// controllerKind contains the schema.GroupVersionKind for the CronJob type.
	controllerKind = v1beta1.SchemeGroupVersion.WithKind("CronJob")

	// stateConfirmationTimeout is the amount of time allowed to wait for desired state.
	stateConfirmationTimeout = 10 * time.Second

	// stateConfirmationInterval is the amount of time between polling for the desired state.
	// The polling is against a local memory cache.
	stateConfirmationInterval = 100 * time.Millisecond

	nextScheduleDelta = 100 * time.Millisecond
)

// Reconciler reconciles a CronJob object.
type Reconciler struct {
	client client.Client

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
	recorder         record.EventRecorder

	jobControl     jobControlInterface
	cronJobControl cronJobControlInterface

	// clock returns the current time and provides fake-clock support for tests.
	clock clock.Clock
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.CronJob{}).
		Owns(&v1beta1.Job{}).
		Watches(
			&v1beta1.Job{},
			handler.EnqueueRequestsFromMapFunc(r.JobToCronJobs),
		).
		WithOptions(options).
		Named(controllerName).
		WithEventFilter(predicates.All(
			ctrl.LoggerFrom(ctx),
			predicates.CronJobNotSuspend(ctrl.LoggerFrom(ctx)),
			predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
		))

	r.client = mgr.GetClient()
	r.recorder = mgr.GetEventRecorderFor("cronjob-controller")
	r.cronJobControl = &realCJControl{client: mgr.GetClient()}
	r.jobControl = &realJobControl{client: mgr.GetClient()}

	r.clock = clock.RealClock{}

	return builder.Complete(r)
}

// Reconcile reads that state of the OneX for a CronJob object and makes changes based on the state read
// and what is in the CronJob.Spec.
func (r *Reconciler) Reconcile(ctx context.Context, rq ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)

	// 1. Fetch the CronJob object
	cronJob := &v1beta1.CronJob{}
	if err := r.client.Get(ctx, rq.NamespacedName, cronJob); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(4).Info("Reconcile cronjob")

	// Initialize the patch helper
	helper, err := patch.NewHelper(cronJob, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchCronJob(ctx, helper, cronJob); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	result, err := r.reconcile(ctx, cronJob)
	if err != nil {
		log.Error(err, "Failed to reconcile CronJob")
		// r.recorder.Warnf(cronJob, "ReconcileError", "%v", err)
	}
	return result, err
}

func (r *Reconciler) reconcile(ctx context.Context, cronJob *v1beta1.CronJob) (ctrl.Result, error) {
	// log := ctrl.LoggerFrom(ctx)

	childJobs, err := r.getJobsToBeReconciled(ctx, cronJob)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 清理历史Job，可结合你之前的cleanupFinishedJobs逻辑
	r.cleanupFinishedJobs(ctx, cronJob, childJobs)

	// 其它核心同步逻辑
	res, err := r.syncCronJob(ctx, cronJob, childJobs)
	if err != nil {
		return ctrl.Result{}, err
	}

	return coreutil.LowestNonZeroResult(res, ctrl.Result{}), nil
}

func patchCronJob(ctx context.Context, helper *patch.Helper, cronJob *v1beta1.CronJob, options ...patch.Option) error {
	return helper.Patch(ctx, cronJob, options...)
}

func (r *Reconciler) cleanupFinishedJobs(ctx context.Context, cronJob *v1beta1.CronJob, jobs []*v1beta1.Job) {
	// If neither limits are active, there is no need to do anything.
	if cronJob.Spec.FailedJobsHistoryLimit == nil && cronJob.Spec.SuccessfulJobsHistoryLimit == nil {
		return
	}

	failedJobs := []*v1beta1.Job{}
	successfulJobs := []*v1beta1.Job{}

	for _, job := range jobs {
		isFinished, finishedStatus := r.getFinishedStatus(job)
		if isFinished && finishedStatus == v1beta1.JobComplete {
			successfulJobs = append(successfulJobs, job)
		} else if isFinished && finishedStatus == v1beta1.JobFailed {
			failedJobs = append(failedJobs, job)
		}
	}

	if cronJob.Spec.SuccessfulJobsHistoryLimit != nil {
		r.removeOldestJobs(ctx, cronJob, successfulJobs, *cronJob.Spec.SuccessfulJobsHistoryLimit)
	}

	if cronJob.Spec.FailedJobsHistoryLimit != nil {
		r.removeOldestJobs(ctx, cronJob, failedJobs, *cronJob.Spec.FailedJobsHistoryLimit)
	}

	return
}

func (r *Reconciler) syncCronJob(ctx context.Context, cronJob *v1beta1.CronJob, jobs []*v1beta1.Job) (ctrl.Result, error) {
	now := r.clock.Now()

	childrenJobs := make(map[types.UID]bool)
	for _, job := range jobs {
		childrenJobs[job.ObjectMeta.UID] = true
		found := inActiveList(cronJob, job.ObjectMeta.UID)
		if !found && !jobutil.IsJobFinished(job) {
			cjCopy, err := r.cronJobControl.GetCronJob(ctx, cronJob.Namespace, cronJob.Name)
			if err != nil {
				return ctrl.Result{}, err
			}
			if inActiveList(cjCopy, job.ObjectMeta.UID) {
				cronJob = cjCopy
				continue
			}
			r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "UnexpectedJob", "Saw a job that the controller did not create or forgot: %s", job.Name)
			// We found an unfinished job that has us as the parent, but it is not in our Active list.
			// This could happen if we crashed right after creating the Job and before updating the status,
			// or if our jobs list is newer than our cj status after a relist, or if someone intentionally created
			// a job that they wanted us to adopt.
		} else if jobutil.IsJobFinished(job) {
			if found {
				_, condition := jobutil.FinishedCondition(job)
				deleteFromActiveList(cronJob, job.ObjectMeta.UID)
				r.recorder.Eventf(cronJob, corev1.EventTypeNormal, "SawCompletedJob", "Saw completed job: %s, condition: %v", job.Name, condition)
			}
			if jobutil.IsJobSucceeded(job) {
				// a job does not have to be in active list, as long as it has completed successfully, we will process the timestamp
				if cronJob.Status.LastSuccessfulTime == nil {
					cronJob.Status.LastSuccessfulTime = job.Status.EndedAt
				}
				if job.Status.EndedAt != nil && job.Status.EndedAt.After(cronJob.Status.LastSuccessfulTime.Time) {
					cronJob.Status.LastSuccessfulTime = job.Status.EndedAt
				}
			}

		}
	}

	// Remove any job reference from the active list if the corresponding job does not exist any more.
	// Otherwise, the cronjob may be stuck in active mode forever even though there is no matching
	// job running.
	for _, job := range cronJob.Status.Active {
		_, found := childrenJobs[job.UID]
		if found {
			continue
		}
		// Explicitly try to get the job from api-server to avoid a slow watch not able to update
		// the job lister on time, giving an unwanted miss
		_, err := r.jobControl.GetJob(ctx, job.Namespace, job.Name)
		switch {
		case errors.IsNotFound(err):
			// The job is actually missing, delete from active list and schedule a new one if within
			// deadline
			r.recorder.Eventf(cronJob, corev1.EventTypeNormal, "MissingJob", "Active job went missing: %v", job.Name)
			deleteFromActiveList(cronJob, job.UID)
		case err != nil:
			return ctrl.Result{}, err
		}
		// the job is missing in the lister but found in api-server
	}

	// The CronJob is being deleted.
	// Don't do anything other than updating status.
	if !cronJob.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	log := ctrl.LoggerFrom(ctx)
	if cronJob.Spec.TimeZone != nil {
		timeZone := ptr.Deref(cronJob.Spec.TimeZone, "")
		if _, err := time.LoadLocation(timeZone); err != nil {
			log.V(4).Info("Not starting job because timeZone is invalid", "cronjob", klog.KObj(cronJob), "timeZone", timeZone, "err", err)
			r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "UnknownTimeZone", "invalid timeZone: %q: %s", timeZone, err)
			return ctrl.Result{}, nil
		}
	}

	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		log.V(4).Info("Not starting job because the cron is suspended", "cronjob", klog.KObj(cronJob))
		return ctrl.Result{}, nil
	}

	sched, err := cron.ParseStandard(formatSchedule(cronJob, r.recorder))
	if err != nil {
		// this is likely a user error in defining the spec value
		// we should log the error and not reconcile this cronjob until an update to spec
		log.V(2).Info("Unparseable schedule", "cronjob", klog.KObj(cronJob), "schedule", cronJob.Spec.Schedule, "err", err)
		r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "UnparseableSchedule", "unparseable schedule: %q : %s", cronJob.Spec.Schedule, err)
		return ctrl.Result{}, nil
	}

	scheduledTime, err := nextScheduleTime(log, cronJob, now, sched, r.recorder)
	if err != nil {
		// this is likely a user error in defining the spec value
		// we should log the error and not reconcile this cronjob until an update to spec
		log.V(2).Info("Invalid schedule", "cronjob", klog.KObj(cronJob), "schedule", cronJob.Spec.Schedule, "err", err)
		r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "InvalidSchedule", "invalid schedule: %s : %s", cronJob.Spec.Schedule, err)
		return ctrl.Result{}, nil
	}
	if scheduledTime == nil {
		// no unmet start time, return cj,.
		// The only time this should happen is if queue is filled after restart.
		// Otherwise, the queue is always suppose to trigger sync function at the time of
		// the scheduled time, that will give atleast 1 unmet time schedule
		log.V(4).Info("No unmet start times", "cronjob", klog.KObj(cronJob))
		t := nextScheduleTimeDuration(cronJob, now, sched)
		return ctrl.Result{RequeueAfter: *t}, nil
	}

	tooLate := false
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = scheduledTime.Add(time.Second * time.Duration(*cronJob.Spec.StartingDeadlineSeconds)).Before(now)
	}
	if tooLate {
		log.V(4).Info("Missed starting window", "cronjob", klog.KObj(cronJob))
		r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "MissSchedule", "Missed scheduled time to start a job: %s", scheduledTime.UTC().Format(time.RFC1123Z))

		// TODO: Since we don't set LastScheduleTime when not scheduling, we are going to keep noticing
		// the miss every cycle.  In order to avoid sending multiple events, and to avoid processing
		// the cj again and again, we could set a Status.LastMissedTime when we notice a miss.
		// Then, when we call getRecentUnmetScheduleTimes, we can take max(creationTimestamp,
		// Status.LastScheduleTime, Status.LastMissedTime), and then so we won't generate
		// and event the next time we process it, and also so the user looking at the status
		// can see easily that there was a missed execution.
		t := nextScheduleTimeDuration(cronJob, now, sched)
		return ctrl.Result{RequeueAfter: *t}, nil
	}
	if inActiveListByName(cronJob, &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getJobName(cronJob, *scheduledTime),
			Namespace: cronJob.Namespace,
		},
	}) || cronJob.Status.LastScheduleTime.Equal(&metav1.Time{Time: *scheduledTime}) {
		log.V(4).Info("Not starting job because the scheduled time is already processed", "cronjob", klog.KObj(cronJob), "schedule", scheduledTime)
		t := nextScheduleTimeDuration(cronJob, now, sched)
		return ctrl.Result{RequeueAfter: *t}, nil
	}
	if cronJob.Spec.ConcurrencyPolicy == v1beta1.ForbidConcurrent && len(cronJob.Status.Active) > 0 {
		// Regardless which source of information we use for the set of active jobs,
		// there is some risk that we won't see an active job when there is one.
		// (because we haven't seen the status update to the SJ or the created pod).
		// So it is theoretically possible to have concurrency with Forbid.
		// As long the as the invocations are "far enough apart in time", this usually won't happen.
		//
		// TODO: for Forbid, we could use the same name for every execution, as a lock.
		// With replace, we could use a name that is deterministic per execution time.
		// But that would mean that you could not inspect prior successes or failures of Forbid jobs.
		log.V(4).Info("Not starting job because prior execution is still running and concurrency policy is Forbid", "cronjob", klog.KObj(cronJob))
		r.recorder.Eventf(cronJob, corev1.EventTypeNormal, "JobAlreadyActive", "Not starting job because prior execution is running and concurrency policy is Forbid")
		t := nextScheduleTimeDuration(cronJob, now, sched)
		return ctrl.Result{RequeueAfter: *t}, nil
	}
	if cronJob.Spec.ConcurrencyPolicy == v1beta1.ReplaceConcurrent {
		for _, jr := range cronJob.Status.Active {
			log.V(4).Info("Deleting job that was still running at next scheduled start time", "job", klog.KRef(jr.Namespace, jr.Name))
			job, err := r.jobControl.GetJob(ctx, jr.Namespace, jr.Name)
			if err != nil {
				r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "FailedGet", "Get job: %v", err)
				return ctrl.Result{}, err
			}
			deleteJob(ctx, cronJob, job, r.jobControl, r.recorder)
			return ctrl.Result{}, fmt.Errorf("could not replace job %s/%s", job.Namespace, job.Name)
		}
	}

	jobAlreadyExists := false
	jobReq, err := getJobFromTemplate2(cronJob, *scheduledTime)
	if err != nil {
		log.Error(err, "Unable to make Job from template", "cronjob", klog.KObj(cronJob))
		return ctrl.Result{}, err
	}

	jobResp, err := r.jobControl.CreateJob(ctx, jobReq)
	switch {
	case errors.HasStatusCause(err, corev1.NamespaceTerminatingCause):
		// if the namespace is being terminated, we don't have to do
		// anything because any creation will fail
		return ctrl.Result{}, err
	case errors.IsAlreadyExists(err):
		// If the job is created by other actor, assume it has updated the cronjob status accordingly.
		// However, if the job was created by cronjob controller, this means we've previously created the job
		// but failed to update the active list in the status, in which case we should reattempt to add the job
		// into the active list and update the status.
		jobAlreadyExists = true
		job, err := r.jobControl.GetJob(ctx, jobReq.GetNamespace(), jobReq.GetName())
		if err != nil {
			return ctrl.Result{}, err
		}
		jobResp = job

		// check that this job is owned by cronjob controller, otherwise do nothing and assume external controller
		// is updating the status.
		if !metav1.IsControlledBy(job, cronJob) {
			return ctrl.Result{}, nil
		}

		// Recheck if the job is missing from the active list before attempting to update the status again.
		if found := inActiveList(cronJob, job.ObjectMeta.UID); found {
			return ctrl.Result{}, nil
		}
	case err != nil:
		// default error handling
		r.recorder.Eventf(cronJob, corev1.EventTypeWarning, "FailedCreate", "Error creating job: %v", err)
		return ctrl.Result{}, err
	}

	if jobAlreadyExists {
		log.Info("Job already exists", "cronjob", klog.KObj(cronJob), "job", klog.KObj(jobReq))
	} else {
		// metrics.CronJobCreationSkew.Observe(jobResp.ObjectMeta.GetCreationTimestamp().Sub(*scheduledTime).Seconds())
		log.V(4).Info("Created Job", "job", klog.KObj(jobResp), "cronjob", klog.KObj(cronJob))
		r.recorder.Eventf(cronJob, corev1.EventTypeNormal, "SuccessfulCreate", "Created job %v", jobResp.Name)
	}

	// ------------------------------------------------------------------ //

	// If this process restarts at this point (after posting a job, but
	// before updating the status), then we might try to start the job on
	// the next time.  Actually, if we re-list the SJs and Jobs on the next
	// iteration of syncAll, we might not see our own status update, and
	// then post one again.  So, we need to use the job name as a lock to
	// prevent us from making the job twice (name the job with hash of its
	// scheduled time).

	// Add the just-started job to the status list.
	jobRef, err := getRef(jobResp)
	if err != nil {
		log.V(2).Info("Unable to make object reference", "cronjob", klog.KObj(cronJob), "err", err)
		return ctrl.Result{}, fmt.Errorf("unable to make object reference for job for %s", klog.KObj(cronJob))
	}
	cronJob.Status.Active = append(cronJob.Status.Active, *jobRef)
	cronJob.Status.LastScheduleTime = &metav1.Time{Time: *scheduledTime}

	t := nextScheduleTimeDuration(cronJob, now, sched)
	return ctrl.Result{RequeueAfter: *t}, nil
}

// shouldExcludeJob returns true if the job should be filtered out, false otherwise.
func shouldExcludeJob(cronJob *v1beta1.CronJob, job *v1beta1.Job) bool {
	if metav1.GetControllerOf(job) != nil && !metav1.IsControlledBy(job, cronJob) {
		return true
	}

	return false
}

// adoptOrphan sets the CronJob as a controller OwnerReference to the Miner.
func (r *Reconciler) adoptOrphan(ctx context.Context, cronJob *v1beta1.CronJob, job *v1beta1.Job) error {
	patch := client.MergeFrom(job.DeepCopy())
	newRef := *metav1.NewControllerRef(cronJob, controllerKind)
	job.OwnerReferences = append(job.OwnerReferences, newRef)
	return r.client.Patch(ctx, job, patch)
}

// JobToCronJobs is a handler.ToRequestsFunc to be used to enqueue rquests for reconciliation
// for CronJobs that might adopt an orphaned Miner.
func (r *Reconciler) JobToCronJobs(ctx context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	j, ok := o.(*v1beta1.Job)
	if !ok {
		panic(fmt.Sprintf("Expected a Job but got a %T", o))
	}

	log := ctrl.LoggerFrom(ctx, "Job", klog.KObj(j)) // TODO: test here

	// Check if the controller reference is already set and
	// return an empty result when one is found.
	for _, ref := range j.ObjectMeta.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return result
		}
	}

	cronJobs, err := r.getCronJobsForJob(ctx, j)
	if err != nil {
		log.Error(err, "Failed getting CronJobs for Miner")
		return nil
	}
	if len(cronJobs) == 0 {
		return nil
	}

	for _, cronJob := range cronJobs {
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cronJob)})
	}

	return result
}

func (r *Reconciler) getCronJobsForJob(ctx context.Context, j *v1beta1.Job) ([]*v1beta1.CronJob, error) {
	cronJobList := &v1beta1.CronJobList{}
	if err := r.client.List(ctx, cronJobList, client.InNamespace(j.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list CronJobs, err: %w", err)
	}

	// Any CronJob in the same namespace is a candidate to adopt an orphaned Job:
	// the Job's deterministic name encodes its owning CronJob and schedule, so the
	// right parent is resolved during the CronJob's own reconciliation.
	cronJobs := make([]*v1beta1.CronJob, 0, len(cronJobList.Items))
	for i := range cronJobList.Items {
		cronJobs = append(cronJobs, &cronJobList.Items[i])
	}

	return cronJobs, nil
}

func (r *Reconciler) getJobsToBeReconciled(ctx context.Context, cronJob *v1beta1.CronJob) ([]*v1beta1.Job, error) {
	// List all Jobs in the namespace and keep the ones that belong to this CronJob,
	// excluding those owned by another controller and adopting orphaned Jobs whose
	// deterministic name matches this CronJob's prefix.
	var jobList v1beta1.JobList
	if err := r.client.List(ctx, &jobList, client.InNamespace(cronJob.Namespace)); err != nil {
		return nil, err
	}

	var childJobs []*v1beta1.Job
	for i := range jobList.Items {
		job := &jobList.Items[i]
		if shouldExcludeJob(cronJob, job) {
			continue
		}

		// Adopt orphan Jobs that were created by a previous CronJob instance but
		// never got (or lost) their controller reference.
		if metav1.GetControllerOf(job) == nil && strings.HasPrefix(job.Name, cronJob.Name+"-") {
			if err := r.adoptOrphan(ctx, cronJob, job); err != nil {
				return nil, err
			}
		}

		childJobs = append(childJobs, job)
	}

	return childJobs, nil
}

func (r *Reconciler) getFinishedStatus(job *v1beta1.Job) (bool, v1beta1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == v1beta1.JobComplete || c.Type == v1beta1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

// removeOldestJobs removes the oldest jobs from a list of jobs
func (r *Reconciler) removeOldestJobs(ctx context.Context, cronJob *v1beta1.CronJob, jobs []*v1beta1.Job, maxJobs int32) {
	numToDelete := len(jobs) - int(maxJobs)
	if numToDelete <= 0 {
		return
	}
	log := klog.FromContext(ctx)
	log.V(4).Info("Cleaning up jobs from CronJob list", "deletejobnum", numToDelete, "jobnum", len(jobs), "cronjob", klog.KObj(cronJob))

	sort.Sort(byJobStartedAt(jobs))
	for i := 0; i < numToDelete; i++ {
		log.V(4).Info("Removing job from CronJob list", "job", jobs[i].Name, "cronjob", klog.KObj(cronJob))
		deleteJob(ctx, cronJob, jobs[i], r.jobControl, r.recorder)
	}
	return
}

// deleteJob reaps a job, deleting the job, the pods and the reference in the active list
func deleteJob(ctx context.Context, cronJob *v1beta1.CronJob, job *v1beta1.Job, jc jobControlInterface, recorder record.EventRecorder) {
	log := klog.FromContext(ctx)
	// delete the job itself...
	if err := jc.DeleteJob(ctx, job); err != nil {
		recorder.Eventf(cronJob, corev1.EventTypeWarning, "FailedDelete", "Deleted job: %v", err)
		log.Error(err, "Error deleting job from cronjob", "job", klog.KObj(job), "cronjob", klog.KObj(cronJob))
		return
	}
	// ... and its reference from active list
	deleteFromActiveList(cronJob, job.ObjectMeta.UID)
	recorder.Eventf(cronJob, corev1.EventTypeNormal, "SuccessfulDelete", "Deleted job %v", job.Name)

	return
}

func getRef(object runtime.Object) (*corev1.ObjectReference, error) {
	return ref.GetReference(scheme.Scheme, object)
}

func formatSchedule(cronJob *v1beta1.CronJob, recorder record.EventRecorder) string {
	if strings.Contains(cronJob.Spec.Schedule, "TZ") {
		if recorder != nil {
			recorder.Eventf(
				cronJob,
				corev1.EventTypeWarning,
				"UnsupportedSchedule",
				"CRON_TZ or TZ used in schedule %q is not officially supported",
				cronJob.Spec.Schedule,
			)
		}

		return cronJob.Spec.Schedule
	}

	if cronJob.Spec.TimeZone != nil {
		if _, err := time.LoadLocation(*cronJob.Spec.TimeZone); err != nil {
			return cronJob.Spec.Schedule
		}

		return fmt.Sprintf("TZ=%s %s", *cronJob.Spec.TimeZone, cronJob.Spec.Schedule)
	}

	return cronJob.Spec.Schedule
}

func getJobName(cronJob *v1beta1.CronJob, scheduledTime time.Time) string {
	return fmt.Sprintf("%s-%d", cronJob.Name, getTimeHashInMinutes(scheduledTime))
}
