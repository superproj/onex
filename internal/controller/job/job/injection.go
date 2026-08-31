// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"context"
	"fmt"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

// jobControlInterface abstracts the writes the controller performs against the
// apiserver so that the controller can be unit tested.
type jobControlInterface interface {
	GetJob(ctx context.Context, namespace, name string) (*v1beta1.Job, error)
	UpdateStatus(ctx context.Context, job *v1beta1.Job) error
	Update(ctx context.Context, job *v1beta1.Job) error
}

// realJobControl is the default implementation of jobControlInterface.
type realJobControl struct {
	client client.Client
}

var _ jobControlInterface = &realJobControl{}

func (c *realJobControl) GetJob(ctx context.Context, namespace, name string) (*v1beta1.Job, error) {
	job := new(v1beta1.Job)
	if err := c.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *realJobControl) UpdateStatus(ctx context.Context, job *v1beta1.Job) error {
	return c.client.Status().Update(ctx, job)
}

func (c *realJobControl) Update(ctx context.Context, job *v1beta1.Job) error {
	return c.client.Update(ctx, job)
}

// providerStatus is the provider's view of the job's current execution state.
type providerStatus struct {
	// Finished is true when the provider workload has reached a terminal state.
	Finished bool
	// Succeeded is true when the provider workload finished successfully.
	Succeeded bool
	// Message is an optional human-readable detail of the current/final state.
	Message string
}

// providerControlInterface abstracts interaction with the workload provider that
// actually executes a Job. It is the extension point for plugging in
// kubernetes / aws-batch / spark-on-yarn backends.
type providerControlInterface interface {
	// Start dispatches the job to the provider and returns an error if the
	// dispatch could not be submitted.
	Start(ctx context.Context, job *v1beta1.Job) error
	// Stop cancels a running provider workload.
	Stop(ctx context.Context, job *v1beta1.Job) error
	// Status returns the provider's view of the job's current state.
	Status(ctx context.Context, job *v1beta1.Job) (providerStatus, error)
	// Children lists provider-managed child objects owned by the job, used for
	// controller-reference adoption/release.
	Children(ctx context.Context, job *v1beta1.Job) ([]client.Object, error)
}

// realProviderControl dispatches to a concrete provider backend based on the
// Job's spec.type, defaulting to the kubernetes backend.
type realProviderControl struct {
	kubernetes providerControlInterface
}

// newRealProviderControl returns a provider backend wired to the given client.
func newRealProviderControl(c client.Client) providerControlInterface {
	return &realProviderControl{kubernetes: &kubernetesProvider{client: c}}
}

var _ providerControlInterface = &realProviderControl{}

func (p *realProviderControl) Start(ctx context.Context, job *v1beta1.Job) error {
	return p.providerFor(job).Start(ctx, job)
}

func (p *realProviderControl) Stop(ctx context.Context, job *v1beta1.Job) error {
	return p.providerFor(job).Stop(ctx, job)
}

func (p *realProviderControl) Status(ctx context.Context, job *v1beta1.Job) (providerStatus, error) {
	return p.providerFor(job).Status(ctx, job)
}

func (p *realProviderControl) Children(ctx context.Context, job *v1beta1.Job) ([]client.Object, error) {
	return p.providerFor(job).Children(ctx, job)
}

func (p *realProviderControl) providerFor(job *v1beta1.Job) providerControlInterface {
	switch job.Spec.Type {
	case "", JobTypeKubernetes:
		return p.kubernetes
	default:
		return &unsupportedProvider{typ: job.Spec.Type}
	}
}

// unsupportedProvider is returned for Job types that have no backend wired yet.
type unsupportedProvider struct {
	typ v1beta1.JobType
}

var _ providerControlInterface = &unsupportedProvider{}

func (u *unsupportedProvider) Start(context.Context, *v1beta1.Job) error {
	return fmt.Errorf("unsupported job type %q", u.typ)
}

func (u *unsupportedProvider) Stop(context.Context, *v1beta1.Job) error { return nil }

func (u *unsupportedProvider) Status(context.Context, *v1beta1.Job) (providerStatus, error) {
	return providerStatus{}, nil
}

func (u *unsupportedProvider) Children(context.Context, *v1beta1.Job) ([]client.Object, error) {
	return nil, nil
}

// fakeJobControl is a test implementation of jobControlInterface.
type fakeJobControl struct {
	sync.Mutex
	Job         *v1beta1.Job
	StatusCalls []*v1beta1.Job
	Err         error
}

var _ jobControlInterface = &fakeJobControl{}

func (f *fakeJobControl) GetJob(_ context.Context, _, _ string) (*v1beta1.Job, error) {
	return f.Job, f.Err
}

func (f *fakeJobControl) UpdateStatus(_ context.Context, job *v1beta1.Job) error {
	f.Lock()
	defer f.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.StatusCalls = append(f.StatusCalls, job.DeepCopy())
	return nil
}

func (f *fakeJobControl) Update(_ context.Context, job *v1beta1.Job) error {
	f.Lock()
	defer f.Unlock()
	return f.Err
}

// fakeProviderControl is a test implementation of providerControlInterface.
type fakeProviderControl struct {
	StartErr     error
	StatusResp   providerStatus
	ChildrenResp []client.Object
}

var _ providerControlInterface = &fakeProviderControl{}

func (f *fakeProviderControl) Start(context.Context, *v1beta1.Job) error { return f.StartErr }
func (f *fakeProviderControl) Stop(context.Context, *v1beta1.Job) error  { return nil }

func (f *fakeProviderControl) Status(context.Context, *v1beta1.Job) (providerStatus, error) {
	return f.StatusResp, nil
}

func (f *fakeProviderControl) Children(context.Context, *v1beta1.Job) ([]client.Object, error) {
	return f.ChildrenResp, nil
}
