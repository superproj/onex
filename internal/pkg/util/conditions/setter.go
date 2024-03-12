// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package conditions

import (
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Setter interface defines methods that a Cluster API object should implement in order to
// use the conditions package for setting conditions.
type Setter interface {
	Getter
	SetConditions(v1beta1.Conditions)
}

// Set sets the given condition.
//
// NOTE: If a condition already exists, the LastTransitionTime is updated only if a change is detected
// in any of the following fields: Status, Reason, Severity and Message.
func Set(to Setter, condition *v1beta1.Condition) {
	if to == nil || condition == nil {
		return
	}

	// Check if the new conditions already exists, and change it only if there is a status
	// transition (otherwise we should preserve the current last transition time)-
	conditions := to.GetConditions()
	exists := false
	for i := range conditions {
		existingCondition := conditions[i]
		if existingCondition.Type == condition.Type {
			exists = true
			if !hasSameState(&existingCondition, condition) {
				condition.LastTransitionTime = metav1.NewTime(time.Now().UTC().Truncate(time.Second))
				conditions[i] = *condition
				break
			}
			condition.LastTransitionTime = existingCondition.LastTransitionTime
			break
		}
	}

	// If the condition does not exist, add it, setting the transition time only if not already set
	if !exists {
		if condition.LastTransitionTime.IsZero() {
			condition.LastTransitionTime = metav1.NewTime(time.Now().UTC().Truncate(time.Second))
		}
		conditions = append(conditions, *condition)
	}

	// Sorts conditions for convenience of the consumer, i.e. kubectl.
	sort.Slice(conditions, func(i, j int) bool {
		return lexicographicLess(&conditions[i], &conditions[j])
	})

	to.SetConditions(conditions)
}

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t v1beta1.ConditionType) *v1beta1.Condition {
	return &v1beta1.Condition{
		Type:   t,
		Status: corev1.ConditionTrue,
	}
}

// FalseCondition returns a condition with Status=False and the given type.
func FalseCondition(t v1beta1.ConditionType, reason string, severity v1beta1.ConditionSeverity, messageFormat string, messageArgs ...any) *v1beta1.Condition {
	return &v1beta1.Condition{
		Type:     t,
		Status:   corev1.ConditionFalse,
		Reason:   reason,
		Severity: severity,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// UnknownCondition returns a condition with Status=Unknown and the given type.
func UnknownCondition(t v1beta1.ConditionType, reason string, messageFormat string, messageArgs ...any) *v1beta1.Condition {
	return &v1beta1.Condition{
		Type:    t,
		Status:  corev1.ConditionUnknown,
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// MarkTrue sets Status=True for the condition with the given type.
func MarkTrue(to Setter, t v1beta1.ConditionType) {
	Set(to, TrueCondition(t))
}

// MarkUnknown sets Status=Unknown for the condition with the given type.
func MarkUnknown(to Setter, t v1beta1.ConditionType, reason, messageFormat string, messageArgs ...any) {
	Set(to, UnknownCondition(t, reason, messageFormat, messageArgs...))
}

// MarkFalse sets Status=False for the condition with the given type.
func MarkFalse(to Setter, t v1beta1.ConditionType, reason string, severity v1beta1.ConditionSeverity, messageFormat string, messageArgs ...any) {
	Set(to, FalseCondition(t, reason, severity, messageFormat, messageArgs...))
}

// SetSummary sets a Ready condition with the summary of all the conditions existing
// on an object. If the object does not have other conditions, no summary condition is generated.
func SetSummary(to Setter, options ...MergeOption) {
	Set(to, summary(to, options...))
}

// SetMirror creates a new condition by mirroring the the Ready condition from a dependent object;
// if the Ready condition does not exists in the source object, no target conditions is generated.
func SetMirror(to Setter, targetCondition v1beta1.ConditionType, from Getter, options ...MirrorOptions) {
	Set(to, mirror(from, targetCondition, options...))
}

// SetAggregate creates a new condition with the aggregation of all the the Ready condition
// from a list of dependent objects; if the Ready condition does not exists in one of the source object,
// the object is excluded from the aggregation; if none of the source object have ready condition,
// no target conditions is generated.
func SetAggregate(to Setter, targetCondition v1beta1.ConditionType, from []Getter, options ...MergeOption) {
	Set(to, aggregate(from, targetCondition, options...))
}

// Delete deletes the condition with the given type.
func Delete(to Setter, t v1beta1.ConditionType) {
	if to == nil {
		return
	}

	conditions := to.GetConditions()
	newConditions := make(v1beta1.Conditions, 0, len(conditions))
	for _, condition := range conditions {
		if condition.Type != t {
			newConditions = append(newConditions, condition)
		}
	}
	to.SetConditions(newConditions)
}

// lexicographicLess returns true if a condition is less than another with regards to the
// to order of conditions designed for convenience of the consumer, i.e. kubectl.
// According to this order the Ready condition always goes first, followed by all the other
// conditions sorted by Type.
func lexicographicLess(i, j *v1beta1.Condition) bool {
	return (i.Type == v1beta1.ReadyCondition || i.Type < j.Type) && j.Type != v1beta1.ReadyCondition
}

// hasSameState returns true if a condition has the same state of another; state is defined
// by the union of following fields: Type, Status, Reason, Severity and Message (it excludes LastTransitionTime).
func hasSameState(i, j *v1beta1.Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Reason == j.Reason &&
		i.Severity == j.Severity &&
		i.Message == j.Message
}
