package jobconditions

import (
	"github.com/superproj/onex/internal/nightwatch/dao/model"
	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
)

// Get retrieves the JobCondition for a specified condition type from the provided JobConditions.
func Get(conditions *model.JobConditions, condType string) *nwv1.JobCondition {
	if conditions == nil {
		return nil
	}

	for _, condition := range *conditions {
		if condition.Type == condType {
			return condition
		}
	}

	return nil
}

// IsTrue checks if the specified condition type in JobConditions is true.
func IsTrue(conditions *model.JobConditions, condType string) bool {
	if cond := Get(conditions, condType); cond != nil {
		return cond.Status == model.ConditionTrue
	}

	return false
}

// IsFalse checks if the specified condition type in JobConditions is false.
func IsFalse(conditions *model.JobConditions, condType string) bool {
	return !IsTrue(conditions, condType)
}
