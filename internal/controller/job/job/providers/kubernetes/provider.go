// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package kubernetes implements the "kubernetes" Job provider, which runs a Job
// by dispatching a single Pod and mapping the Pod phase back onto the Job
// lifecycle.
package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onexstack/onex/internal/controller/job/job/providers/registry"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

const (
	// Type is the provider type for running a Job as a Kubernetes Pod. It is
	// also the implicit default when spec.type is left empty.
	Type v1beta1.JobType = "kubernetes"

	// jobNameLabel is the label key used to associate the Pod dispatched for a
	// Job with that Job, so that the provider can locate its children.
	jobNameLabel = "job.onex.io/name"
)

// controllerKind is the GroupVersionKind used when writing the Pod's controller
// reference back to the owning Job.
var controllerKind = v1beta1.SchemeGroupVersion.WithKind("Job")

// Provider runs a Job by dispatching a single Pod and mapping the Pod phase back
// onto the Job lifecycle. It expects job.Spec.ProviderSpec.Raw to carry a
// JSON-encoded corev1.PodSpec.
type Provider struct {
	client client.Client
}

var _ registry.Provider = (*Provider)(nil)

// Start creates the Pod that executes the Job.
func (p *Provider) Start(ctx context.Context, job *v1beta1.Job) error {
	spec, err := podSpecFromJob(job)
	if err != nil {
		return err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    job.Name + "-",
			Namespace:       job.Namespace,
			Labels:          map[string]string{jobNameLabel: job.Name},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
		},
		Spec: *spec,
	}

	if err := p.client.Create(ctx, pod); err != nil {
		return fmt.Errorf("failed to create pod for job %s/%s: %w", job.Namespace, job.Name, err)
	}
	return nil
}

// Stop deletes all Pods owned by the Job.
func (p *Provider) Stop(ctx context.Context, job *v1beta1.Job) error {
	pods, err := p.getPods(ctx, job)
	if err != nil {
		return err
	}
	for i := range pods {
		if err := p.client.Delete(ctx, &pods[i]); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete pod %s/%s: %w", pods[i].Namespace, pods[i].Name, err)
		}
	}
	return nil
}

// Status maps the Job's Pod phase to a provider status. A Job runs a single
// Pod, so its terminal state is derived from that Pod's phase.
func (p *Provider) Status(ctx context.Context, job *v1beta1.Job) (registry.Status, error) {
	pods, err := p.getPods(ctx, job)
	if err != nil {
		return registry.Status{}, err
	}
	if len(pods) == 0 {
		// No Pod observed yet (e.g. immediately after Start, before the
		// informer has caught up).
		return registry.Status{}, nil
	}

	pod := pods[0]
	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return registry.Status{Finished: true, Succeeded: true}, nil
	case corev1.PodFailed:
		return registry.Status{Finished: true, Succeeded: false, Message: podFailureMessage(&pod)}, nil
	default:
		return registry.Status{}, nil
	}
}

// Children returns the Pods owned by the Job, as client objects, for
// controller-reference adoption/release.
func (p *Provider) Children(ctx context.Context, job *v1beta1.Job) ([]client.Object, error) {
	pods, err := p.getPods(ctx, job)
	if err != nil {
		return nil, err
	}
	objs := make([]client.Object, 0, len(pods))
	for i := range pods {
		objs = append(objs, &pods[i])
	}
	return objs, nil
}

func (p *Provider) getPods(ctx context.Context, job *v1beta1.Job) ([]corev1.Pod, error) {
	list := &corev1.PodList{}
	if err := p.client.List(ctx, list,
		client.InNamespace(job.Namespace),
		client.MatchingLabels{jobNameLabel: job.Name},
	); err != nil {
		return nil, fmt.Errorf("failed to list pods for job %s/%s: %w", job.Namespace, job.Name, err)
	}
	return list.Items, nil
}

// podSpecFromJob decodes the kubernetes provider spec from the Job's
// providerSpec raw extension.
func podSpecFromJob(job *v1beta1.Job) (*corev1.PodSpec, error) {
	if job.Spec.ProviderSpec == nil || len(job.Spec.ProviderSpec.Raw) == 0 {
		return nil, fmt.Errorf("job %s/%s has no providerSpec", job.Namespace, job.Name)
	}
	spec := &corev1.PodSpec{}
	if err := json.Unmarshal(job.Spec.ProviderSpec.Raw, spec); err != nil {
		return nil, fmt.Errorf("failed to decode providerSpec for job %s/%s: %w", job.Namespace, job.Name, err)
	}
	return spec, nil
}

// podFailureMessage extracts a human-readable failure reason from a failed Pod.
func podFailureMessage(pod *corev1.Pod) string {
	for i := range pod.Status.ContainerStatuses {
		if st := pod.Status.ContainerStatuses[i].State.Terminated; st != nil && st.ExitCode != 0 {
			if st.Message != "" {
				return st.Message
			}
			return st.Reason
		}
	}
	return string(pod.Status.Phase)
}

func init() {
	registry.Register(Type, func(c client.Client) registry.Provider {
		return &Provider{client: c}
	})
}
