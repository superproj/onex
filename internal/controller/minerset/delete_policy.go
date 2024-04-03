// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerset

import (
	"fmt"
	"math"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/superproj/onex/internal/pkg/util/conditions"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

type (
	deletePriority     float64
	deletePriorityFunc func(miner *v1beta1.Miner) deletePriority
)

const (
	mustDelete    deletePriority = 100.0
	betterDelete  deletePriority = 50.0
	couldDelete   deletePriority = 20.0
	mustNotDelete deletePriority = 0.0

	secondsPerTenDays float64 = 864000
)

// maps the creation timestamp onto the 0-100 priority range.
func oldestDeletePriority(miner *v1beta1.Miner) deletePriority {
	if !miner.DeletionTimestamp.IsZero() {
		return mustDelete
	}
	if _, ok := miner.ObjectMeta.Annotations[v1beta1.DeleteMinerAnnotation]; ok {
		return mustDelete
	}
	if !isMinerHealthy(miner) {
		return mustDelete
	}
	if miner.ObjectMeta.CreationTimestamp.Time.IsZero() {
		return mustNotDelete
	}
	d := metav1.Now().Sub(miner.ObjectMeta.CreationTimestamp.Time)
	if d.Seconds() < 0 {
		return mustNotDelete
	}
	return deletePriority(float64(mustDelete) * (1.0 - math.Exp(-d.Seconds()/secondsPerTenDays)))
}

func newestDeletePriority(miner *v1beta1.Miner) deletePriority {
	if !miner.DeletionTimestamp.IsZero() {
		return mustDelete
	}
	if _, ok := miner.ObjectMeta.Annotations[v1beta1.DeleteMinerAnnotation]; ok {
		return mustDelete
	}
	if !isMinerHealthy(miner) {
		return mustDelete
	}
	return mustDelete - oldestDeletePriority(miner)
}

func randomDeletePolicy(miner *v1beta1.Miner) deletePriority {
	if !miner.DeletionTimestamp.IsZero() {
		return mustDelete
	}
	if _, ok := miner.ObjectMeta.Annotations[v1beta1.DeleteMinerAnnotation]; ok {
		return betterDelete
	}
	if !isMinerHealthy(miner) {
		return betterDelete
	}
	return couldDelete
}

type sortableMiners struct {
	miners   []*v1beta1.Miner
	priority deletePriorityFunc
}

func (m sortableMiners) Len() int      { return len(m.miners) }
func (m sortableMiners) Swap(i, j int) { m.miners[i], m.miners[j] = m.miners[j], m.miners[i] }
func (m sortableMiners) Less(i, j int) bool {
	priorityI, priorityJ := m.priority(m.miners[i]), m.priority(m.miners[j])
	if priorityI == priorityJ {
		// In cases where the priority is identical, it should be ensured that the same miner order is returned each time.
		// Ordering by name is a simple way to do this.
		return m.miners[i].Name < m.miners[j].Name
	}
	return priorityJ < priorityI // high to low
}

func getMinersToDeletePrioritized(filteredMiners []*v1beta1.Miner, diff int, fun deletePriorityFunc) []*v1beta1.Miner {
	if diff >= len(filteredMiners) {
		return filteredMiners
	} else if diff <= 0 {
		return []*v1beta1.Miner{}
	}

	sortable := sortableMiners{
		miners:   filteredMiners,
		priority: fun,
	}
	sort.Sort(sortable)

	return sortable.miners[:diff]
}

func getDeletePriorityFunc(ms *v1beta1.MinerSet) (deletePriorityFunc, error) {
	// Map the Spec.DeletePolicy value to the appropriate delete priority function
	switch msdp := v1beta1.MinerSetDeletePolicy(ms.Spec.DeletePolicy); msdp {
	case v1beta1.RandomMinerSetDeletePolicy:
		return randomDeletePolicy, nil
	case v1beta1.NewestMinerSetDeletePolicy:
		return newestDeletePriority, nil
	case v1beta1.OldestMinerSetDeletePolicy:
		return oldestDeletePriority, nil
	case "":
		return randomDeletePolicy, nil
	default:
		return nil, fmt.Errorf("unsupported delete policy %s. Must be one of 'Random', 'Newest', or 'Oldest'", msdp)
	}
}

func isMinerHealthy(miner *v1beta1.Miner) bool {
	if miner.Status.PodRef == nil {
		return false
	}
	if miner.Status.FailureReason != nil || miner.Status.FailureMessage != nil {
		return false
	}
	podHealthyCondition := conditions.Get(miner, v1beta1.MinerPodHealthyCondition)
	if podHealthyCondition != nil && podHealthyCondition.Status != corev1.ConditionTrue {
		return false
	}
	return true
}
