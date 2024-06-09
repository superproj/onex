// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// IsMinerHealthy returns true if the the miner is running and miner node is healthy.
func IsMinerHealthy(c client.Client, miner *v1beta1.Miner) bool {
	return true
}

func PhaseFromPodHealthyCondition(cond *v1beta1.Condition) v1beta1.MinerPhase {
	if cond == nil {
		return v1beta1.MinerPhaseProvisioning
	}

	if cond.Status == corev1.ConditionTrue {
		return v1beta1.MinerPhaseRunning
	}

	switch cond.Reason {
	case v1beta1.PodNotFoundReason, v1beta1.PodConditionsFailedReason:
		return v1beta1.MinerPhaseFailed
	case v1beta1.WaitingForPodRefReason, v1beta1.PodProvisioningReason:
		return v1beta1.MinerPhaseProvisioning
	default:
	}

	return v1beta1.MinerPhaseUnknown
}

// IsMinerAvailable returns true if the pod is ready and minReadySeconds have elapsed or is 0. False otherwise.
func IsMinerAvailable(m *v1beta1.Miner, minReadySeconds int32, now metav1.Time) bool {
	if !IsMinerReady(m) {
		return false
	}

	if minReadySeconds == 0 {
		return true
	}

	minReadySecondsDuration := time.Duration(minReadySeconds) * time.Second
	readyCondition := GetReadyCondition(&m.Status)

	if !readyCondition.LastTransitionTime.IsZero() &&
		readyCondition.LastTransitionTime.Add(minReadySecondsDuration).Before(now.Time) {
		return true
	}

	return false
}

// GetReadyCondition extracts the ready condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetReadyCondition(status *v1beta1.MinerStatus) *v1beta1.Condition {
	if status == nil {
		return nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == v1beta1.MinerPodHealthyCondition {
			return &status.Conditions[i]
		}
	}
	return nil
}

// IsMinerReady returns true if a miner is ready; false otherwise.
func IsMinerReady(m *v1beta1.Miner) bool {
	if m == nil {
		return false
	}
	for _, c := range m.Status.Conditions {
		if c.Type == v1beta1.MinerPodHealthyCondition {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func ChainDNSServiceNameFromMiner(namespace, name string) string {
	return fmt.Sprintf("%s.%s.svc.superproj.com", name, namespace)
}

func GenesisDNSServiceNameFromMiner(name string) string {
	return ChainDNSServiceNameFromMiner(metav1.NamespaceSystem, name)
}

func GetProviderServiceName(m *v1beta1.Miner) string {
	// If miner is an genesis node machine, the service name must be the same as chan name
	chainName, ok := m.Labels[v1beta1.ChainNameLabel]
	if ok {
		return chainName
	}
	return m.Name
}

func IsGenesisMiner(m *v1beta1.Miner) bool {
	_, ok := m.Labels[v1beta1.ChainNameLabel]
	return ok
}
