// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package pod

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	providerIDPrefix = "kind://"
)

// IsPodAvailable returns true if the pod is ready and minReadySeconds have elapsed or is 0. False otherwise.
func IsPodAvailable(pod *corev1.Pod, minReadySeconds int32, now metav1.Time) bool {
	if !IsPodReady(pod) {
		return false
	}

	if minReadySeconds == 0 {
		return true
	}

	minReadySecondsDuration := time.Duration(minReadySeconds) * time.Second
	readyCondition := GetReadyCondition(&pod.Status)

	if !readyCondition.LastTransitionTime.IsZero() &&
		readyCondition.LastTransitionTime.Add(minReadySecondsDuration).Before(now.Time) {
		return true
	}

	return false
}

// GetReadyCondition extracts the ready condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetReadyCondition(status *corev1.PodStatus) *corev1.PodCondition {
	if status == nil {
		return nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == corev1.PodReady {
			return &status.Conditions[i]
		}
	}
	return nil
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// IsPodUnreachable returns true if a pod is unreachable.
// Pod is considered unreachable when its ready status is "Unknown".
func IsPodUnreachable(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionUnknown
		}
	}
	return false
}
