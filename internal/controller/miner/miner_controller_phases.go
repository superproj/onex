// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/superproj/onex/internal/pkg/known"
	"github.com/superproj/onex/internal/pkg/util/conditions"
	minerutil "github.com/superproj/onex/internal/pkg/util/miner"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/record"
)

var externalReadyWait = 30 * time.Second

func (r *Reconciler) reconcilePhase(_ context.Context, m *v1beta1.Miner) {
	originalPhase := m.Status.Phase

	// Set the phase to "Pending" if nil.
	if m.Status.Phase == "" {
		m.Status.SetTypedPhase(v1beta1.MinerPhasePending)
	}

	// Set phase to "Provisioning" if podRef has been set and the pod is not ready.
	if m.Status.PodRef != nil && !conditions.IsTrue(m, v1beta1.MinerPodHealthyCondition) {
		m.Status.SetTypedPhase(v1beta1.MinerPhaseProvisioning)
	}

	// Set phase to "Running" if podRef has been set and the pod is ready
	if m.Status.PodRef != nil && conditions.IsTrue(m, v1beta1.MinerPodHealthyCondition) {
		m.Status.SetTypedPhase(v1beta1.MinerPhaseRunning)
	}

	// Set the phase to "Failed" if any of Status.FailureReason or Status.FailureMessage is not-nil.
	if m.Status.FailureReason != nil || m.Status.FailureMessage != nil {
		m.Status.SetTypedPhase(v1beta1.MinerPhaseFailed)
	}

	// Set the phase to "Deleting" if the deletion timestamp is set.
	if !m.DeletionTimestamp.IsZero() {
		m.Status.SetTypedPhase(v1beta1.MinerPhaseDeleting)
	}

	// If the phase has changed, update the LastUpdated timestamp
	if m.Status.Phase != originalPhase {
		now := metav1.Now()
		m.Status.LastUpdated = &now
	}
}

func (r *Reconciler) reconcileAnnotations(ctx context.Context, m *v1beta1.Miner) (ctrl.Result, error) {
	needReconcile := false
	for _, annotation := range []string{known.CPUAnnotation, known.MemoryAnnotation} {
		if _, ok := m.Annotations[annotation]; !ok {
			needReconcile = true
		}
	}

	if !needReconcile {
		return ctrl.Result{}, nil
	}

	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	cpu := r.ComponentConfig.Types[m.Spec.MinerType].CPU
	memory := r.ComponentConfig.Types[m.Spec.MinerType].Memory
	m.Annotations[known.CPUAnnotation] = cpu.String()
	m.Annotations[known.MemoryAnnotation] = memory.String()

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileProviderService(ctx context.Context, m *v1beta1.Miner) (ctrl.Result, error) {
	// only create service for genesis miners
	if !minerutil.IsGenesisMiner(m) {
		return ctrl.Result{}, nil
	}

	log := ctrl.LoggerFrom(ctx)
	if _, err := r.ProviderClient.CoreV1().Services(m.Namespace).Get(ctx, minerutil.GetProviderServiceName(m), metav1.GetOptions{}); err == nil {
		return ctrl.Result{}, nil
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.Namespace,
			Name:      minerutil.GetProviderServiceName(m),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{v1beta1.MinerSetNameLabel: m.Name},
			// ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Name:       "websocket",
					Protocol:   corev1.ProtocolTCP,
					Port:       6001,
					TargetPort: intstr.IntOrString{IntVal: 6001},
				},
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.IntOrString{IntVal: 38080},
				},
			},
		},
	}

	_, err := r.ProviderClient.CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return ctrl.Result{}, nil
		}

		record.Warnf(m, "FailedCreate", "Failed to get Service %s: %v", svc.Name, err)
		log.Error(err, "Failed to create service", "service", klog.KObj(svc))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
