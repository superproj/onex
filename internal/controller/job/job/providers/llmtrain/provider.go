// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package llmtrain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onexstack/onex/internal/controller/job/job/providers/registry"
	"github.com/onexstack/onex/internal/pkg/client/minio"
	fakeminio "github.com/onexstack/onex/internal/pkg/client/minio/fake"
	"github.com/onexstack/onex/internal/pkg/client/train"
	"github.com/onexstack/onex/pkg/apis/batch/v1beta1"
)

// Provider runs a Job as an in-process LLM training pipeline (download -> embed
// -> train). Unlike the kubernetes provider it materializes no child workload;
// its progress is persisted in job.Status.ProviderStatus and advanced on each
// Status poll.
type Provider struct {
	minio    minio.IMinio
	trainer  trainer
	embedder embedder
}

var _ registry.Provider = (*Provider)(nil)

// newProvider returns a Provider wired with the given storage, trainer and
// embedder. It is unexported because it is only used by init and white-box tests.
func newProvider(m minio.IMinio, tr trainer, e embedder) *Provider {
	return &Provider{minio: m, trainer: tr, embedder: e}
}

// Start decodes the provider spec and seeds the pipeline in the Pending phase.
func (p *Provider) Start(ctx context.Context, job *v1beta1.Job) error {
	if _, err := decodeSpec(job); err != nil {
		return err
	}
	return encodeStatus(job, &Status{Phase: phasePending})
}

// Stop is a no-op: the pipeline owns no child workload to cancel.
func (p *Provider) Stop(context.Context, *v1beta1.Job) error {
	return nil
}

// Children returns no children: the in-process pipeline creates no objects.
func (p *Provider) Children(context.Context, *v1beta1.Job) ([]client.Object, error) {
	return nil, nil
}

// Status advances the pipeline by one phase and reports the result.
func (p *Provider) Status(ctx context.Context, job *v1beta1.Job) (registry.Status, error) {
	spec, err := decodeSpec(job)
	if err != nil {
		return registry.Status{}, err
	}
	st, err := decodeStatus(job)
	if err != nil {
		return registry.Status{}, err
	}

	switch st.Phase {
	case phaseSucceeded:
		return registry.Status{Finished: true, Succeeded: true}, nil
	case phaseFailed:
		return registry.Status{Finished: true, Succeeded: false, Message: st.Message}, nil
	}

	if isTimeout(job, spec) {
		st.Phase = phaseFailed
		st.Message = fmt.Sprintf("llm train task exceeded %d seconds", timeoutSeconds(spec))
		if err := encodeStatus(job, st); err != nil {
			return registry.Status{}, err
		}
		return registry.Status{Finished: true, Succeeded: false, Message: st.Message}, nil
	}

	next, err := p.advance(ctx, job.Name, spec, st)
	if err != nil {
		st.Phase = phaseFailed
		st.Message = err.Error()
		if err := encodeStatus(job, st); err != nil {
			return registry.Status{}, err
		}
		return registry.Status{Finished: true, Succeeded: false, Message: err.Error()}, nil
	}
	st.Phase = next
	if err := encodeStatus(job, st); err != nil {
		return registry.Status{}, err
	}
	return registry.Status{}, nil
}

// decodeSpec decodes the LLMTrain provider spec, applying defaults.
func decodeSpec(job *v1beta1.Job) (*Spec, error) {
	spec := &Spec{}
	if job.Spec.ProviderSpec != nil && len(job.Spec.ProviderSpec.Raw) > 0 {
		if err := json.Unmarshal(job.Spec.ProviderSpec.Raw, spec); err != nil {
			return nil, fmt.Errorf("failed to decode llmtrain providerSpec for job %s/%s: %w", job.Namespace, job.Name, err)
		}
	}
	if spec.SourceObject == "" {
		spec.SourceObject = fakeminio.FakeObjectName
	}
	return spec, nil
}

// decodeStatus decodes the LLMTrain pipeline state, defaulting to Pending.
func decodeStatus(job *v1beta1.Job) (*Status, error) {
	st := &Status{Phase: phasePending}
	if job.Status.ProviderStatus != nil && len(job.Status.ProviderStatus.Raw) > 0 {
		if err := json.Unmarshal(job.Status.ProviderStatus.Raw, st); err != nil {
			return nil, fmt.Errorf("failed to decode llmtrain providerStatus for job %s/%s: %w", job.Namespace, job.Name, err)
		}
	}
	if st.Phase == "" {
		st.Phase = phasePending
	}
	return st, nil
}

// encodeStatus serializes the pipeline state back into the Job status.
func encodeStatus(job *v1beta1.Job, st *Status) error {
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	job.Status.ProviderStatus = &runtime.RawExtension{Raw: raw}
	return nil
}

func timeoutSeconds(spec *Spec) int64 {
	if spec.JobTimeoutSec > 0 {
		return spec.JobTimeoutSec
	}
	return defaultJobTimeout
}

func isTimeout(job *v1beta1.Job, spec *Spec) bool {
	if job.Status.StartedAt == nil {
		return false
	}
	return time.Since(job.Status.StartedAt.Time) > time.Duration(timeoutSeconds(spec))*time.Second
}

func init() {
	registry.Register(Type, func(client.Client) registry.Provider {
		// NewFakeMinioClient cannot fail for a valid bucket name.
		m, _ := fakeminio.NewFakeMinioClient("demo")
		return newProvider(m, train.NewTrainManager(), deterministicEmbedder{})
	})
}
