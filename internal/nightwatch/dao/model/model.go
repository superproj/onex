package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	nwv1 "github.com/superproj/onex/pkg/api/nightwatch/v1"
)

var (
	ErrCronJobStatusInvalidType = errors.New("invalid type for CronJobStatus")
	ErrJobMInvalidType          = errors.New("invalid type for JobM")
	ErrJobParamsInvalidType     = errors.New("invalid type for JobParams")
	ErrJobResultsInvalidType    = errors.New("invalid type for JobResults")
	ErrJobConditionsInvalidType = errors.New("invalid type for JobConditions")
)

const (
	ConditionTrue    string = "True"
	ConditionFalse   string = "False"
	ConditionUnknown string = "Unknown"
)

type CronJobStatus nwv1.CronJobStatus

// Scan implements the sql Scanner interface
func (status *CronJobStatus) Scan(value interface{}) error {
	if value == nil {
		*status = CronJobStatus{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrCronJobStatusInvalidType
	}

	return json.Unmarshal(bytes, status)
}

// Value implements the sql Valuer interface
func (status *CronJobStatus) Value() (driver.Value, error) {
	return json.Marshal(status)
}

type JobParams nwv1.JobParams

// Scan implements the sql Scanner interface
func (params *JobParams) Scan(value interface{}) error {
	if value == nil {
		*params = JobParams{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrJobParamsInvalidType
	}

	return json.Unmarshal(bytes, params)
}

// Value implements the sql Valuer interface
func (params *JobParams) Value() (driver.Value, error) {
	return json.Marshal(params)
}

type JobResults nwv1.JobResults

// Scan implements the sql Scanner interface
func (result *JobResults) Scan(value interface{}) error {
	if value == nil {
		*result = JobResults{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrJobResultsInvalidType
	}

	return json.Unmarshal(bytes, result)
}

// Value implements the sql Valuer interface
func (result *JobResults) Value() (driver.Value, error) {
	return json.Marshal(result)
}

type JobConditions []*nwv1.JobCondition

// Scan implements the sql Scanner interface
func (conds *JobConditions) Scan(value interface{}) error {
	if value == nil {
		*conds = JobConditions{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrJobConditionsInvalidType
	}

	return json.Unmarshal(bytes, conds)
}

// Value implements the sql Valuer interface
func (conds *JobConditions) Value() (driver.Value, error) {
	return json.Marshal(conds)
}

// Scan implements the sql Scanner interface
func (job *JobM) Scan(value interface{}) error {
	if value == nil {
		*job = JobM{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrJobMInvalidType
	}

	return json.Unmarshal(bytes, job)
}

// Value implements the sql Valuer interface
func (job *JobM) Value() (driver.Value, error) {
	return json.Marshal(job)
}
