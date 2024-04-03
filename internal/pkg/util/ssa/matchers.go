// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package ssa

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/superproj/onex/internal/pkg/contract"
)

// MatchManagedFieldsEntry is a gomega Matcher to check if a ManagedFieldsEntry has the given name and operation.
func MatchManagedFieldsEntry(manager string, operation metav1.ManagedFieldsOperationType) types.GomegaMatcher {
	return &managedFieldMatcher{
		manager:   manager,
		operation: operation,
	}
}

type managedFieldMatcher struct {
	manager   string
	operation metav1.ManagedFieldsOperationType
}

func (mf *managedFieldMatcher) Match(actual any) (bool, error) {
	managedFieldsEntry, ok := actual.(metav1.ManagedFieldsEntry)
	if !ok {
		return false, fmt.Errorf("expecting metav1.ManagedFieldsEntry got %T", actual)
	}

	return managedFieldsEntry.Manager == mf.manager && managedFieldsEntry.Operation == mf.operation, nil
}

func (mf *managedFieldMatcher) FailureMessage(actual any) string {
	managedFieldsEntry := actual.(metav1.ManagedFieldsEntry)
	return fmt.Sprintf("Expected ManagedFieldsEntry to match Manager:%s and Operation:%s, got Manager:%s, Operation:%s",
		mf.manager, mf.operation, managedFieldsEntry.Manager, managedFieldsEntry.Operation)
}

func (mf *managedFieldMatcher) NegatedFailureMessage(actual any) string {
	managedFieldsEntry := actual.(metav1.ManagedFieldsEntry)
	return fmt.Sprintf("Expected ManagedFieldsEntry to not match Manager:%s and Operation:%s, got Manager:%s, Operation:%s",
		mf.manager, mf.operation, managedFieldsEntry.Manager, managedFieldsEntry.Operation)
}

// MatchFieldOwnership is a gomega Matcher to check if path is owned by the given manager and operation.
// Note: The path has to be specified as is observed in managed fields. Example: to check if the labels are owned
// by the correct manager the correct way to pass the path is contract.Path{"f:metadata","f:labels"}.
func MatchFieldOwnership(manager string, operation metav1.ManagedFieldsOperationType, path contract.Path) types.GomegaMatcher {
	return &fieldOwnershipMatcher{
		path:      path,
		manager:   manager,
		operation: operation,
	}
}

type fieldOwnershipMatcher struct {
	path      contract.Path
	manager   string
	operation metav1.ManagedFieldsOperationType
}

func (fom *fieldOwnershipMatcher) Match(actual any) (bool, error) {
	managedFields, ok := actual.([]metav1.ManagedFieldsEntry)
	if !ok {
		return false, fmt.Errorf("expecting []metav1.ManagedFieldsEntry got %T", actual)
	}
	for _, managedFieldsEntry := range managedFields {
		if managedFieldsEntry.Manager == fom.manager && managedFieldsEntry.Operation == fom.operation {
			fieldsV1 := map[string]any{}
			if err := json.Unmarshal(managedFieldsEntry.FieldsV1.Raw, &fieldsV1); err != nil {
				return false, errors.Wrap(err, "failed to parse managedFieldsEntry.FieldsV1")
			}
			FilterIntent(&FilterIntentInput{
				Path:         contract.Path{},
				Value:        fieldsV1,
				ShouldFilter: IsPathNotAllowed([]contract.Path{fom.path}),
			})
			return len(fieldsV1) > 0, nil
		}
	}
	return false, nil
}

func (fom *fieldOwnershipMatcher) FailureMessage(actual any) string {
	managedFields := actual.([]metav1.ManagedFieldsEntry)
	return fmt.Sprintf("Expected Path %s to be owned by Manager:%s and Operation:%s, did not find correct ownership: %s",
		fom.path, fom.manager, fom.operation, managedFields)
}

func (fom *fieldOwnershipMatcher) NegatedFailureMessage(actual any) string {
	managedFields := actual.([]metav1.ManagedFieldsEntry)
	return fmt.Sprintf("Expected Path %s to not be owned by Manager:%s and Operation:%s, did not find correct ownership: %s",
		fom.path, fom.manager, fom.operation, managedFields)
}
