// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

const testPodSpec = `{"restartPolicy":"Never","containers":[{"name":"echo","image":"busybox"}]}`

func routingScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	return scheme
}

// TestProviderRouting verifies that the registry dispatches each Job type to the
// registered provider and falls back to the kubernetes provider for an empty
// type, erroring for unknown types.
func TestProviderRouting(t *testing.T) {
	p := newRealProviderControl(fake.NewClientBuilder().WithScheme(routingScheme()).Build())
	ctx := context.Background()

	// Both the explicit and the empty type route to the kubernetes provider.
	for _, typ := range []v1beta1.JobType{"", "kubernetes"} {
		name := "j-" + string(typ)
		if name == "j-" {
			name = "j-default"
		}
		job := &v1beta1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: v1beta1.JobSpec{
				Type:         typ,
				ProviderSpec: &runtime.RawExtension{Raw: []byte(testPodSpec)},
			},
		}
		if err := p.Start(ctx, job); err != nil {
			t.Fatalf("type %q: start: %v", typ, err)
		}
	}

	// LLMTrain seeds its pipeline rather than creating a Pod.
	llm := &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "j-llm", Namespace: "default"},
		Spec:       v1beta1.JobSpec{Type: "LLMTrain"},
	}
	if err := p.Start(ctx, llm); err != nil {
		t.Fatalf("LLMTrain start: %v", err)
	}
	if llm.Status.ProviderStatus == nil || len(llm.Status.ProviderStatus.Raw) == 0 {
		t.Fatal("expected LLMTrain start to seed providerStatus")
	}

	// Unknown types report an unsupported job type error.
	unsupported := &v1beta1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "j-x", Namespace: "default"},
		Spec:       v1beta1.JobSpec{Type: "aws-batch"},
	}
	err := p.Start(ctx, unsupported)
	if err == nil || !strings.Contains(err.Error(), "unsupported job type") {
		t.Fatalf("expected unsupported job type error, got %v", err)
	}
}
