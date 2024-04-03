// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package labels contains useful functions to process onex labels.
package labels

import (
	"encoding/base64"
	"fmt"
	"hash/fnv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// HasMatchingLabels verifies that the Label Selector matches the given Labels.
func HasMatchingLabels(matchSelector metav1.LabelSelector, matchLabels map[string]string) bool {
	// This should never fail, validating webhook should catch this first
	selector, err := metav1.LabelSelectorAsSelector(&matchSelector)
	if err != nil {
		return false
	}
	// If a nil or empty selector creeps in, it should match nothing, not everything.
	if selector.Empty() {
		return false
	}
	if !selector.Matches(labels.Set(matchLabels)) {
		return false
	}
	return true
}

// HasWatchLabel returns true if the object has a label with the WatchLabel key matching the given value.
func HasWatchLabel(o metav1.Object, labelValue string) bool {
	val, ok := o.GetLabels()[v1beta1.WatchLabel]
	if !ok {
		return false
	}
	return val == labelValue
}

// MustFormatValue returns the passed inputLabelValue if it meets the standards for a Kubernetes label value.
// If the name is not a valid label value this function returns a hash which meets the requirements.
func MustFormatValue(str string) string {
	// a valid Kubernetes label value must:
	// - be less than 64 characters long.
	// - be an empty string OR consist of alphanumeric characters, '-', '_' or '.'.
	// - start and end with an alphanumeric character
	if len(validation.IsValidLabelValue(str)) == 0 {
		return str
	}
	hasher := fnv.New32a()
	_, err := hasher.Write([]byte(str))
	if err != nil {
		// At time of writing the implementation of fnv's Write function can never return an error.
		// If this changes in a future go version this function will panic.
		panic(err)
	}
	return fmt.Sprintf("hash_%s_z", base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hasher.Sum(nil)))
}

// MustEqualValue returns true if the actualLabelValue equals either the inputLabelValue or the hashed
// value of the inputLabelValue.
func MustEqualValue(str, labelValue string) bool {
	return labelValue == MustFormatValue(str)
}
