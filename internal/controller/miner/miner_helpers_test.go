// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHasMatchingLabels(t *testing.T) {
	testCases := []struct {
		name     string
		selector metav1.LabelSelector
		labels   map[string]string
		expected bool
	}{
		{
			name: "selector matches labels",
			selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
			labels: map[string]string{
				"foo":  "bar",
				"more": "labels",
			},
			expected: true,
		},
		{
			name: "selector does not match labels",
			selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
			labels: map[string]string{
				"no": "match",
			},
			expected: false,
		},
		{
			name:     "selector is empty",
			selector: metav1.LabelSelector{},
			labels:   map[string]string{},
			expected: false,
		},
		{
			name: "selector is invalid",
			selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Operator: "bad-operator",
					},
				},
			},
			labels: map[string]string{
				"foo": "bar",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			got := HasMatchingLabels(tc.selector, tc.labels)
			g.Expect(got).To(gomega.Equal(tc.expected))
		})
	}
}
