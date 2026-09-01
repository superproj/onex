// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package llmtrain

import (
	"context"
	"fmt"

	"github.com/onexstack/onex/internal/pkg/client/train"
)

// trainer abstracts the training platform so the pipeline can be unit tested
// without the in-memory TrainManager's simulated delay.
type trainer interface {
	CreateTask(ctx context.Context, dataPath, resultPath string) (string, error)
	GetTaskStatus(ctx context.Context, taskID string) (string, error)
}

// embedder produces a stable, serialized representation for each document.
type embedder interface {
	EmbedDocuments(ctx context.Context, docs []string) []string
}

// deterministicEmbedder derives a small fixed-shape vector per document so the
// demo runs without an external embedding service, mirroring the reference
// implementation's fake embedder.
type deterministicEmbedder struct{}

func (deterministicEmbedder) EmbedDocuments(_ context.Context, docs []string) []string {
	out := make([]string, 0, len(docs))
	for i, d := range docs {
		out = append(out, fmt.Sprintf("[%d,%d]", i, len(d)))
	}
	return out
}

func dataPath(jobName string) string     { return fmt.Sprintf("job/%s/data.json", jobName) }
func embeddedPath(jobName string) string { return fmt.Sprintf("job/%s/embedded.json", jobName) }
func resultPath(jobName string) string   { return fmt.Sprintf("job/%s/result.json", jobName) }

// advance executes the work of the current phase and returns the next phase. A
// nil error with next == st.Phase means the pipeline stays put (training still
// in progress). A non-nil error means the phase terminally failed.
func (p *Provider) advance(ctx context.Context, jobName string, spec *Spec, st *Status) (string, error) {
	switch st.Phase {
	case phasePending:
		return phaseDownloading, nil
	case phaseDownloading:
		if err := p.download(ctx, jobName, spec, st); err != nil {
			return "", err
		}
		return phaseDownloaded, nil
	case phaseDownloaded:
		return phaseEmbedding, nil
	case phaseEmbedding:
		if err := p.embed(ctx, jobName, st); err != nil {
			return "", err
		}
		return phaseEmbedded, nil
	case phaseEmbedded:
		return phaseTraining, nil
	case phaseTraining:
		done, err := p.train(ctx, jobName, st)
		if err != nil {
			return "", err
		}
		if !done {
			return phaseTraining, nil // stay; poll again on the next cycle
		}
		return phaseTrained, nil
	case phaseTrained:
		return phaseSucceeded, nil
	default:
		return "", fmt.Errorf("unexpected llmtrain phase %q", st.Phase)
	}
}

// download copies the source data to the job's own data object.
func (p *Provider) download(ctx context.Context, jobName string, spec *Spec, st *Status) error {
	if spec.IdempotentExecution && st.DataPath != nil {
		return nil // already performed
	}
	data, err := p.minio.Read(ctx, spec.SourceObject)
	if err != nil {
		return err
	}
	key := dataPath(jobName)
	if err := p.minio.Write(ctx, key, data); err != nil {
		return err
	}
	st.DataPath = &key
	return nil
}

// embed reads the raw data, embeds it and writes the embedded representation.
func (p *Provider) embed(ctx context.Context, jobName string, st *Status) error {
	if st.DataPath == nil {
		return fmt.Errorf("data path is nil")
	}
	docs, err := p.minio.Read(ctx, *st.DataPath)
	if err != nil {
		return err
	}
	embs := p.embedder.EmbedDocuments(ctx, docs)
	key := embeddedPath(jobName)
	if err := p.minio.Write(ctx, key, embs); err != nil {
		return err
	}
	st.EmbeddedDataPath = &key
	st.TaskID = nil // force a fresh training task on re-embed
	return nil
}

// train submits the training task once, then polls it to completion. It returns
// (false, nil) while still running so the pipeline stays in phaseTraining.
func (p *Provider) train(ctx context.Context, jobName string, st *Status) (bool, error) {
	if st.EmbeddedDataPath == nil {
		return false, fmt.Errorf("embedded data path is nil")
	}
	if st.TaskID == nil {
		key := resultPath(jobName)
		id, err := p.trainer.CreateTask(ctx, *st.EmbeddedDataPath, key)
		if err != nil {
			return false, err
		}
		st.TaskID = &id
		st.ResultPath = &key
	}
	status, err := p.trainer.GetTaskStatus(ctx, *st.TaskID)
	if err != nil {
		return false, err
	}
	return status == train.StatusCompleted, nil
}
