/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	apimachineryapivalidation "k8s.io/apimachinery/pkg/api/validation"
	apimachineryvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	api "k8s.io/kubernetes/pkg/apis/core"
	apivalidation "k8s.io/kubernetes/pkg/apis/core/validation"
	"k8s.io/utils/ptr"

	"github.com/onexstack/onex/pkg/apis/batch"
)

// ValidateJob validates a Job and returns an ErrorList with any errors.
func ValidateJob(job *batch.Job) field.ErrorList {
	// Jobs and rcs have the same name validation
	allErrs := apivalidation.ValidateObjectMeta(&job.ObjectMeta, true, apimachineryapivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateJobSpec(&job.Spec, field.NewPath("spec"))...)
	return allErrs
}

// ValidateJobSpec validates a JobSpec and returns an ErrorList with any errors.
func ValidateJobSpec(spec *batch.JobSpec, fldPath *field.Path) field.ErrorList {
	allErrs := validateJobSpec(spec, fldPath)
	return allErrs
}

func validateJobSpec(spec *batch.JobSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if spec.ActiveDeadlineSeconds != nil {
		allErrs = append(allErrs, apivalidation.ValidateNonnegativeField(int64(*spec.ActiveDeadlineSeconds), fldPath.Child("activeDeadlineSeconds"))...)
	}
	return allErrs
}

// validateJobStatus validates a JobStatus and returns an ErrorList with any errors.
func validateJobStatus(job *batch.Job, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

// ValidateJobUpdate validates an update to a Job and returns an ErrorList with any errors.
func ValidateJobUpdate(job, oldJob *batch.Job) field.ErrorList {
	allErrs := apivalidation.ValidateObjectMetaUpdate(&job.ObjectMeta, &oldJob.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateJobSpecUpdate(job.Spec, oldJob.Spec, field.NewPath("spec"))...)
	return allErrs
}

// ValidateJobUpdateStatus validates an update to the status of a Job and returns an ErrorList with any errors.
func ValidateJobUpdateStatus(job, oldJob *batch.Job) field.ErrorList {
	allErrs := apivalidation.ValidateObjectMetaUpdate(&job.ObjectMeta, &oldJob.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateJobStatusUpdate(job, oldJob)...)
	return allErrs
}

// ValidateJobSpecUpdate validates an update to a JobSpec and returns an ErrorList with any errors.
func ValidateJobSpecUpdate(spec, oldSpec batch.JobSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

// ValidateJobStatusUpdate validates an update to a JobStatus and returns an ErrorList with any errors.
func ValidateJobStatusUpdate(job, oldJob *batch.Job) field.ErrorList {
	allErrs := field.ErrorList{}
	statusFld := field.NewPath("status")
	allErrs = append(allErrs, validateJobStatus(job, statusFld)...)

	return allErrs
}

// ValidateCronJobCreate validates a CronJob on creation and returns an ErrorList with any errors.
func ValidateCronJobCreate(cronJob *batch.CronJob) field.ErrorList {
	// CronJobs and rcs have the same name validation
	allErrs := apivalidation.ValidateObjectMeta(&cronJob.ObjectMeta, true, apimachineryapivalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, validateCronJobSpec(&cronJob.Spec, nil, field.NewPath("spec"))...)
	if len(cronJob.ObjectMeta.Name) > apimachineryvalidation.DNS1035LabelMaxLength-11 {
		// The cronjob controller appends a 11-character suffix to the cronjob (`-$TIMESTAMP`) when
		// creating a job. The job name length limit is 63 characters.
		// Therefore cronjob names must have length <= 63-11=52. If we don't validate this here,
		// then job creation will fail later.
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("name"), cronJob.ObjectMeta.Name, "must be no more than 52 characters"))
	}
	return allErrs
}

// ValidateCronJobUpdate validates an update to a CronJob and returns an ErrorList with any errors.
func ValidateCronJobUpdate(job, oldJob *batch.CronJob) field.ErrorList {
	allErrs := apivalidation.ValidateObjectMetaUpdate(&job.ObjectMeta, &oldJob.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, validateCronJobSpec(&job.Spec, &oldJob.Spec, field.NewPath("spec"))...)

	// skip the 52-character name validation limit on update validation
	// to allow old cronjobs with names > 52 chars to be updated/deleted
	return allErrs
}

// validateCronJobSpec validates a CronJobSpec and returns an ErrorList with any errors.
func validateCronJobSpec(spec, oldSpec *batch.CronJobSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(spec.Schedule) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("schedule"), ""))
	} else {
		allowTZInSchedule := false
		if oldSpec != nil {
			allowTZInSchedule = strings.Contains(oldSpec.Schedule, "TZ")
		}
		allErrs = append(allErrs, validateScheduleFormat(spec.Schedule, allowTZInSchedule, spec.TimeZone, fldPath.Child("schedule"))...)
	}

	if spec.StartingDeadlineSeconds != nil {
		allErrs = append(allErrs, apivalidation.ValidateNonnegativeField(int64(*spec.StartingDeadlineSeconds), fldPath.Child("startingDeadlineSeconds"))...)
	}

	if oldSpec == nil || !ptr.Equal(oldSpec.TimeZone, spec.TimeZone) {
		allErrs = append(allErrs, validateTimeZone(spec.TimeZone, fldPath.Child("timeZone"))...)
	}

	allErrs = append(allErrs, validateConcurrencyPolicy(&spec.ConcurrencyPolicy, fldPath.Child("concurrencyPolicy"))...)

	if spec.SuccessfulJobsHistoryLimit != nil {
		// zero is a valid SuccessfulJobsHistoryLimit
		allErrs = append(allErrs, apivalidation.ValidateNonnegativeField(int64(*spec.SuccessfulJobsHistoryLimit), fldPath.Child("successfulJobsHistoryLimit"))...)
	}
	if spec.FailedJobsHistoryLimit != nil {
		// zero is a valid SuccessfulJobsHistoryLimit
		allErrs = append(allErrs, apivalidation.ValidateNonnegativeField(int64(*spec.FailedJobsHistoryLimit), fldPath.Child("failedJobsHistoryLimit"))...)
	}

	return allErrs
}

func validateConcurrencyPolicy(concurrencyPolicy *batch.ConcurrencyPolicy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch *concurrencyPolicy {
	case batch.AllowConcurrent, batch.ForbidConcurrent, batch.ReplaceConcurrent:
		break
	case "":
		allErrs = append(allErrs, field.Required(fldPath, ""))
	default:
		validValues := []batch.ConcurrencyPolicy{batch.AllowConcurrent, batch.ForbidConcurrent, batch.ReplaceConcurrent}
		allErrs = append(allErrs, field.NotSupported(fldPath, *concurrencyPolicy, validValues))
	}

	return allErrs
}

func validateScheduleFormat(schedule string, allowTZInSchedule bool, timeZone *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if _, err := cron.ParseStandard(schedule); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, schedule, err.Error()))
	}
	switch {
	case allowTZInSchedule && strings.Contains(schedule, "TZ") && timeZone != nil:
		allErrs = append(allErrs, field.Invalid(fldPath, schedule, "cannot use both timeZone field and TZ or CRON_TZ in schedule"))
	case !allowTZInSchedule && strings.Contains(schedule, "TZ"):
		allErrs = append(allErrs, field.Invalid(fldPath, schedule, "cannot use TZ or CRON_TZ in schedule, use timeZone field instead"))
	}

	return allErrs
}

// https://data.iana.org/time-zones/theory.html#naming
// * A name must not be empty, or contain '//', or start or end with '/'.
// * Do not use the file name components '.' and '..'.
// * Within a file name component, use only ASCII letters, '.', '-' and '_'.
// * Do not use digits, as that might create an ambiguity with POSIX TZ strings.
// * A file name component must not exceed 14 characters or start with '-'
//
// 0-9 and + characters are tolerated to accommodate legacy compatibility names
var validTimeZoneCharacters = regexp.MustCompile(`^[A-Za-z\.\-_0-9+]{1,14}$`)

func validateTimeZone(timeZone *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if timeZone == nil {
		return allErrs
	}

	if len(*timeZone) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, "timeZone must be nil or non-empty string"))
		return allErrs
	}

	for _, part := range strings.Split(*timeZone, "/") {
		if part == "." || part == ".." || strings.HasPrefix(part, "-") || !validTimeZoneCharacters.MatchString(part) {
			allErrs = append(allErrs, field.Invalid(fldPath, timeZone, fmt.Sprintf("unknown time zone %s", *timeZone)))
			return allErrs
		}
	}

	if strings.EqualFold(*timeZone, "Local") {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, "timeZone must be an explicit time zone as defined in https://www.iana.org/time-zones"))
	}

	if _, err := time.LoadLocation(*timeZone); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, timeZone, err.Error()))
	}

	return allErrs
}

func IsJobFinished(job *batch.Job) bool {
	for _, c := range job.Status.Conditions {
		if (c.Type == batch.JobComplete || c.Type == batch.JobFailed) && c.Status == api.ConditionTrue {
			return true
		}
	}
	return false
}

func IsJobComplete(job *batch.Job) bool {
	return IsConditionTrue(job.Status.Conditions, batch.JobComplete)
}

func IsJobFailed(job *batch.Job) bool {
	return IsConditionTrue(job.Status.Conditions, batch.JobFailed)
}

func isJobSuccessCriteriaMet(job *batch.Job) bool {
	return IsConditionTrue(job.Status.Conditions, batch.JobSuccessCriteriaMet)
}

func isJobFailureTarget(job *batch.Job) bool {
	return IsConditionTrue(job.Status.Conditions, batch.JobFailureTarget)
}

func IsConditionTrue(list []batch.JobCondition, cType batch.JobConditionType) bool {
	for _, c := range list {
		if c.Type == cType && c.Status == api.ConditionTrue {
			return true
		}
	}
	return false
}

func validateFailedIndexesNotOverlapCompleted(completedIndexesStr string, failedIndexesStr string, completions int32) error {
	if len(completedIndexesStr) == 0 || len(failedIndexesStr) == 0 {
		return nil
	}
	completedIndexesIntervals := strings.Split(completedIndexesStr, ",")
	failedIndexesIntervals := strings.Split(failedIndexesStr, ",")
	var completedPos, failedPos int
	cX, cY, cErr := parseIndexInterval(completedIndexesIntervals[completedPos], completions)
	fX, fY, fErr := parseIndexInterval(failedIndexesIntervals[failedPos], completions)
	for completedPos < len(completedIndexesIntervals) && failedPos < len(failedIndexesIntervals) {
		if cErr != nil {
			// Failure to parse "completed" interval. We go to the next interval,
			// the error will be reported to the user when validating the format.
			completedPos++
			if completedPos < len(completedIndexesIntervals) {
				cX, cY, cErr = parseIndexInterval(completedIndexesIntervals[completedPos], completions)
			}
		} else if fErr != nil {
			// Failure to parse "failed" interval. We go to the next interval,
			// the error will be reported to the user when validating the format.
			failedPos++
			if failedPos < len(failedIndexesIntervals) {
				fX, fY, fErr = parseIndexInterval(failedIndexesIntervals[failedPos], completions)
			}
		} else {
			// We have one failed and one completed interval parsed.
			if cX <= fY && fX <= cY {
				return fmt.Errorf("failedIndexes and completedIndexes overlap at index: %d", max(cX, fX))
			}
			// No overlap, let's move to the next one.
			if cX <= fX {
				completedPos++
				if completedPos < len(completedIndexesIntervals) {
					cX, cY, cErr = parseIndexInterval(completedIndexesIntervals[completedPos], completions)
				}
			} else {
				failedPos++
				if failedPos < len(failedIndexesIntervals) {
					fX, fY, fErr = parseIndexInterval(failedIndexesIntervals[failedPos], completions)
				}
			}
		}
	}
	return nil
}

func validateIndexesFormat(indexesStr string, completions int32) (int32, error) {
	if len(indexesStr) == 0 {
		return 0, nil
	}
	var lastIndex *int32
	var total int32
	for _, intervalStr := range strings.Split(indexesStr, ",") {
		x, y, err := parseIndexInterval(intervalStr, completions)
		if err != nil {
			return 0, err
		}
		if lastIndex != nil && *lastIndex >= x {
			return 0, fmt.Errorf("non-increasing order, previous: %d, current: %d", *lastIndex, x)
		}
		total += y - x + 1
		lastIndex = &y
	}
	return total, nil
}

func parseIndexInterval(intervalStr string, completions int32) (int32, int32, error) {
	limitsStr := strings.Split(intervalStr, "-")
	if len(limitsStr) > 2 {
		return 0, 0, fmt.Errorf("the fragment %q violates the requirement that an index interval can have at most two parts separated by '-'", intervalStr)
	}
	x, err := strconv.Atoi(limitsStr[0])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert string to integer for index: %q", limitsStr[0])
	}
	if x >= int(completions) {
		return 0, 0, fmt.Errorf("too large index: %q", limitsStr[0])
	}
	if len(limitsStr) > 1 {
		y, err := strconv.Atoi(limitsStr[1])
		if err != nil {
			return 0, 0, fmt.Errorf("cannot convert string to integer for index: %q", limitsStr[1])
		}
		if y >= int(completions) {
			return 0, 0, fmt.Errorf("too large index: %q", limitsStr[1])
		}
		if x >= y {
			return 0, 0, fmt.Errorf("non-increasing order, previous: %d, current: %d", x, y)
		}
		return int32(x), int32(y), nil
	}
	return int32(x), int32(x), nil
}
