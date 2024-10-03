package jobconditions

import (
	"fmt"
	"time"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
)

// Set updates or adds a JobCondition to the provided JobConditions.
func Set(conditions *model.JobConditions, cond *nwv1.JobCondition) *model.JobConditions {
	if cond == nil {
		return conditions
	}
	if conditions == nil {
		conditions = &model.JobConditions{}
	}

	exists := false
	for i := range *conditions {
		existingCondition := (*conditions)[i]
		if existingCondition.Type == cond.Type {
			exists = true
			if !hasSameState(existingCondition, cond) {
				cond.LastTransitionTime = time.Now().Format(time.DateTime)
				(*conditions)[i] = cond
				break
			}
			cond.LastTransitionTime = existingCondition.LastTransitionTime
			break
		}
	}

	// If the condition does not exist, add it, setting the transition time only if not already set
	if !exists {
		cond.LastTransitionTime = time.Now().Format(time.DateTime)
		*conditions = append(*conditions, cond)
	}

	return conditions
}

// Delete deletes the condition with the given type.
func Delete(conditions *model.JobConditions, condType string) {
	if conditions == nil {
		return
	}

	newConditions := make(model.JobConditions, 0)
	for _, condition := range *conditions {
		if condition.Type != condType {
			newConditions = append(newConditions, condition)
		}
	}
	*conditions = newConditions
}

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t string) *nwv1.JobCondition {
	return &nwv1.JobCondition{
		Type:   t,
		Status: model.ConditionTrue,
	}
}

// FalseCondition returns a condition with Status=False and the given type.
func FalseCondition(t string, messageFormat string, messageArgs ...any) *nwv1.JobCondition {
	return &nwv1.JobCondition{
		Type:    t,
		Status:  model.ConditionFalse,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// UnknownCondition returns a condition with Status=Unknown and the given type.
func UnknownCondition(t string, messageFormat string, messageArgs ...any) *nwv1.JobCondition {
	return &nwv1.JobCondition{
		Type:    t,
		Status:  model.ConditionUnknown,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// hasSameState returns true if a condition has the same state of another; state is defined
// by the union of following fields: Type, Status, Reason, Severity and Message (it excludes LastTransitionTime).
func hasSameState(i, j *nwv1.JobCondition) bool {
	return i.Type == j.Type && i.Status == j.Status && i.Message == j.Message
}
