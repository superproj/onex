// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package registry provides the pluggable provider registry used by the
// batch/v1beta1 Job controller. Providers register a Factory for a Job type and
// the controller resolves the concrete backend via Build.
package registry

import (
	"context"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

// Status is a provider's view of the job's current execution state.
type Status struct {
	// Finished is true when the provider workload has reached a terminal state.
	Finished bool
	// Succeeded is true when the provider workload finished successfully.
	Succeeded bool
	// Message is an optional human-readable detail of the current/final state.
	Message string
}

// Provider abstracts interaction with the workload provider that actually
// executes a Job. It is the extension point for plugging in kubernetes /
// LLMTrain / aws-batch / spark-on-yarn backends.
type Provider interface {
	// Start dispatches the job to the provider and returns an error if the
	// dispatch could not be submitted.
	Start(ctx context.Context, job *v1beta1.Job) error
	// Stop cancels a running provider workload.
	Stop(ctx context.Context, job *v1beta1.Job) error
	// Status returns the provider's view of the job's current state.
	Status(ctx context.Context, job *v1beta1.Job) (Status, error)
	// Children lists provider-managed child objects owned by the job, used for
	// controller-reference adoption/release.
	Children(ctx context.Context, job *v1beta1.Job) ([]client.Object, error)
}

// Factory constructs a Provider bound to the given controller-runtime client.
type Factory func(c client.Client) Provider

var (
	mu        sync.Mutex
	factories = map[v1beta1.JobType]Factory{}
)

// Register registers a provider factory for a Job type. It panics if a factory
// has already been registered for the same type.
func Register(t v1beta1.JobType, f Factory) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := factories[t]; ok {
		panic("duplicate provider entry: " + string(t))
	}

	factories[t] = f
}

// Build constructs every registered provider using the provided client and
// returns them keyed by Job type. The returned map is detached from the registry
// and safe to read concurrently.
func Build(c client.Client) map[v1beta1.JobType]Provider {
	mu.Lock()
	defer mu.Unlock()

	out := make(map[v1beta1.JobType]Provider, len(factories))
	for t, f := range factories {
		out[t] = f(c)
	}
	return out
}
