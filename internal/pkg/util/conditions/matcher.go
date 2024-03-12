// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package conditions

import (
	"fmt"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// MatchConditions returns a custom matcher to check equality of v1beta1.Conditions.
func MatchConditions(expected v1beta1.Conditions) types.GomegaMatcher {
	return &matchConditions{
		expected: expected,
	}
}

type matchConditions struct {
	expected v1beta1.Conditions
}

func (m matchConditions) Match(actual any) (success bool, err error) {
	elems := []any{}
	for _, condition := range m.expected {
		elems = append(elems, MatchCondition(condition))
	}

	return gomega.ConsistOf(elems...).Match(actual)
}

func (m matchConditions) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("expected\n\t%#v\nto match\n\t%#v\n", actual, m.expected)
}

func (m matchConditions) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("expected\n\t%#v\nto not match\n\t%#v\n", actual, m.expected)
}

// MatchCondition returns a custom matcher to check equality of v1beta1.Condition.
func MatchCondition(expected v1beta1.Condition) types.GomegaMatcher {
	return &matchCondition{
		expected: expected,
	}
}

type matchCondition struct {
	expected v1beta1.Condition
}

func (m matchCondition) Match(actual any) (success bool, err error) {
	actualCondition, ok := actual.(v1beta1.Condition)
	if !ok {
		return false, fmt.Errorf("actual should be of type Condition")
	}

	ok, err = gomega.Equal(m.expected.Type).Match(actualCondition.Type)
	if !ok {
		return ok, err
	}
	ok, err = gomega.Equal(m.expected.Status).Match(actualCondition.Status)
	if !ok {
		return ok, err
	}
	ok, err = gomega.Equal(m.expected.Severity).Match(actualCondition.Severity)
	if !ok {
		return ok, err
	}
	ok, err = gomega.Equal(m.expected.Reason).Match(actualCondition.Reason)
	if !ok {
		return ok, err
	}
	ok, err = gomega.Equal(m.expected.Message).Match(actualCondition.Message)
	if !ok {
		return ok, err
	}

	return ok, err
}

func (m matchCondition) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("expected\n\t%#v\nto match\n\t%#v\n", actual, m.expected)
}

func (m matchCondition) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("expected\n\t%#v\nto not match\n\t%#v\n", actual, m.expected)
}
