// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package annotations implements annotation helper functions.
package annotations

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Set will set the value of an annotation on the supplied object. If there is no annotation it will be created.
func Set(obj metav1.Object, name, value string) {
	_ = AddAnnotations(obj, map[string]string{name: value})
}

// Get will get the value of the supplied annotation.
func Get(obj metav1.Object, name string) (value string, found bool) {
	annotations := obj.GetAnnotations()
	if len(annotations) == 0 {
		return "", false
	}

	value, found = annotations[name]

	return
}

// Has returns true if the supplied object has the supplied annotation.
func Has(obj metav1.Object, name string) bool {
	annotations := obj.GetAnnotations()
	if len(annotations) == 0 {
		return false
	}

	_, found := annotations[name]

	return found
}

// IsPaused returns true if the object has the `paused` annotation.
func IsPaused(obj metav1.Object) bool {
	return Has(obj, v1beta1.PausedAnnotation)
}

// HasSkipRemediation returns true if the object has the `skip-remediation` annotation.
func HasSkipRemediation(o metav1.Object) bool {
	return Has(o, v1beta1.MinerSkipRemediationAnnotation)
}

// HasWithPrefix returns true if at least one of the annotations has the prefix specified.
func HasWithPrefix(prefix string, annotations map[string]string) bool {
	for key := range annotations {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// AddAnnotations sets the desired annotations on the object and returns true if the annotations have changed.
func AddAnnotations(o metav1.Object, desired map[string]string) bool {
	if len(desired) == 0 {
		return false
	}
	annotations := o.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	hasChanged := false
	for k, v := range desired {
		if cur, ok := annotations[k]; !ok || cur != v {
			annotations[k] = v
			hasChanged = true
		}
	}
	o.SetAnnotations(annotations)
	return hasChanged
}
