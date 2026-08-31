// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

const testPodSpec = `{"restartPolicy":"Never","containers":[{"name":"echo","image":"busybox"}]}`

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	return scheme
}

func newTestJob(typ v1beta1.JobType) *v1beta1.Job {
	return &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "j", Namespace: "default", UID: "uid-1"},
		Spec: v1beta1.JobSpec{
			Type:         typ,
			ProviderSpec: &runtime.RawExtension{Raw: []byte(testPodSpec)},
		},
	}
}

func TestPodSpecFromJob(t *testing.T) {
	spec, err := podSpecFromJob(newTestJob(JobTypeKubernetes))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.RestartPolicy != corev1.RestartPolicyNever {
		t.Fatalf("expected restart policy Never, got %q", spec.RestartPolicy)
	}
	if len(spec.Containers) != 1 || spec.Containers[0].Name != "echo" {
		t.Fatalf("expected a single echo container, got %+v", spec.Containers)
	}

	empty := &v1beta1.Job{ObjectMeta: metav1.ObjectMeta{Name: "j", Namespace: "default"}}
	if _, err := podSpecFromJob(empty); err == nil {
		t.Fatal("expected error for empty providerSpec")
	}

	invalid := newTestJob(JobTypeKubernetes)
	invalid.Spec.ProviderSpec.Raw = []byte("not-json")
	if _, err := podSpecFromJob(invalid); err == nil {
		t.Fatal("expected error for invalid providerSpec")
	}
}

func TestProviderDispatcherUnsupportedType(t *testing.T) {
	p := newRealProviderControl(nil)
	job := newTestJob("aws-batch")

	if err := p.Start(context.Background(), job); err == nil {
		t.Fatal("expected unsupported type error")
	}
	// Non-terminal methods should be no-ops rather than panic/error.
	if _, err := p.Status(context.Background(), job); err != nil {
		t.Fatalf("unexpected status error: %v", err)
	}
}

func TestKubernetesProviderLifecycle(t *testing.T) {
	p := &kubernetesProvider{client: fake.NewClientBuilder().WithScheme(newTestScheme()).Build()}
	job := newTestJob(JobTypeKubernetes)
	ctx := context.Background()

	if err := p.Start(ctx, job); err != nil {
		t.Fatalf("start: %v", err)
	}

	pods, err := p.getPods(ctx, job)
	if err != nil {
		t.Fatalf("getPods: %v", err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
	if got := metav1.GetControllerOf(&pods[0]); got == nil || got.UID != job.UID {
		t.Fatalf("expected pod to be owned by the job, got controllerRef=%v", got)
	}

	// A pending/running pod is not finished.
	st, err := p.Status(ctx, job)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if st.Finished {
		t.Fatalf("expected running pod to be unfinished, got %+v", st)
	}

	// Mark the pod succeeded and re-check.
	pod := pods[0]
	pod.Status.Phase = corev1.PodSucceeded
	if err := p.client.Status().Update(ctx, &pod); err != nil {
		t.Fatalf("update pod status: %v", err)
	}
	st, err = p.Status(ctx, job)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !st.Finished || !st.Succeeded {
		t.Fatalf("expected finished+succeeded, got %+v", st)
	}

	// Children should surface the owned pod.
	children, err := p.Children(ctx, job)
	if err != nil {
		t.Fatalf("children: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}