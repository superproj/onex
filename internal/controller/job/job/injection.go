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

	_ "github.com/onexstack/onex/internal/controller/job/job/providers/all"
	kubernetes "github.com/onexstack/onex/internal/controller/job/job/providers/kubernetes"
	"github.com/onexstack/onex/internal/controller/job/job/providers/registry"
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

// realProviderControl dispatches to the concrete provider registered for the
// Job's spec.type, defaulting to the kubernetes backend when the type is empty.
type realProviderControl struct {
	providers map[v1beta1.JobType]registry.Provider
}

// newRealProviderControl returns a provider dispatcher wired to the given client.
func newRealProviderControl(c client.Client) registry.Provider {
	return &realProviderControl{providers: registry.Build(c)}
}

var _ registry.Provider = &realProviderControl{}

func (p *realProviderControl) Start(ctx context.Context, job *v1beta1.Job) error {
	return p.providerFor(job).Start(ctx, job)
}

func (p *realProviderControl) Stop(ctx context.Context, job *v1beta1.Job) error {
	return p.providerFor(job).Stop(ctx, job)
}

func (p *realProviderControl) Status(ctx context.Context, job *v1beta1.Job) (registry.Status, error) {
	return p.providerFor(job).Status(ctx, job)
}

func (p *realProviderControl) Children(ctx context.Context, job *v1beta1.Job) ([]client.Object, error) {
	return p.providerFor(job).Children(ctx, job)
}

func (p *realProviderControl) providerFor(job *v1beta1.Job) registry.Provider {
	t := job.Spec.Type
	if t == "" {
		t = kubernetes.Type
	}
	if prov, ok := p.providers[t]; ok {
		return prov
	}
	return &unsupportedProvider{typ: t}
}

// unsupportedProvider is returned for Job types that have no backend registered.
type unsupportedProvider struct {
	typ v1beta1.JobType
}

var _ registry.Provider = &unsupportedProvider{}

func (u *unsupportedProvider) Start(context.Context, *v1beta1.Job) error {
	return fmt.Errorf("unsupported job type %q", u.typ)
}

func (u *unsupportedProvider) Stop(context.Context, *v1beta1.Job) error { return nil }

func (u *unsupportedProvider) Status(context.Context, *v1beta1.Job) (registry.Status, error) {
	return registry.Status{}, nil
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

// fakeProviderControl is a test implementation of registry.Provider.
type fakeProviderControl struct {
	StartErr     error
	StatusResp   registry.Status
	ChildrenResp []client.Object
}

var _ registry.Provider = &fakeProviderControl{}

func (f *fakeProviderControl) Start(context.Context, *v1beta1.Job) error { return f.StartErr }
func (f *fakeProviderControl) Stop(context.Context, *v1beta1.Job) error  { return nil }

func (f *fakeProviderControl) Status(context.Context, *v1beta1.Job) (registry.Status, error) {
	return f.StatusResp, nil
}

func (f *fakeProviderControl) Children(context.Context, *v1beta1.Job) ([]client.Object, error) {
	return f.ChildrenResp, nil
}
