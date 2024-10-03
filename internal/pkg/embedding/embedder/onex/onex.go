package onex

import (
	"context"
	"fmt"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/superproj/onex/pkg/log"
	"go.uber.org/ratelimit"
)

const (
	// EmbeddingQualificationThreshold defines the minimum acceptable ratio of
	// successful embeddings to inputs.
	EmbeddingQualificationThreshold = 0.8
)

// Embedder interface defines a method for embedding input data.
type Embedder interface {
	Embedding(ctx context.Context, input any) string
}

type Extra struct {
	ID string `json:"id,omitempty"`
}

// EmbeddingData represents the structure of data to be embedded.
type EmbeddingData struct {
	Extra Extra     `json:"extra,omitempty"`
	Label string    `json:"label,omitempty"`
	Text  string    `json:"text,omitempty"`
	Emb   []float32 `json:"emb,omitempty"`
}

// EmbeddingType represents the type of embedding being performed.
type EmbeddingType int

const (
	// TextEmbeddingType indicates an embedding type for text data.
	TextEmbeddingType EmbeddingType = iota
	// ImageEmbeddingType indicates an embedding type for image data.
	ImageEmbeddingType
)

// onexEmbedder implements the Embedder interface and manages concurrent embeddings.
type onexEmbedder struct {
	concurrency int
	rl          ratelimit.Limiter
	embedder    Embedder
}

// NewEmbedder creates a new instance of *onexEmbedder with specified options.
func NewEmbedder(typed Embedder, opts ...Option) *onexEmbedder {
	emb := &onexEmbedder{
		concurrency: defaultMaxConcurrency,
		rl:          ratelimit.New(defaultRateLimit),
		embedder:    typed,
	}
	// Apply options to configure the embedder
	for _, opt := range opts {
		opt(emb)
	}

	return emb
}

// Embedding performs embedding on a slice of inputs and returns a slice of results.
func (emb *onexEmbedder) Embedding(ctx context.Context, inputs []any) ([]string, error) {
	if len(inputs) == 0 {
		log.C(ctx).Errorw(nil, "Failed to embedding empty inputs")
		return nil, fmt.Errorf("failed to embedding empty inputs")
	}

	var mu sync.RWMutex
	wp := workerpool.New(emb.concurrency)
	retList := make([]string, 0)

	for i := range inputs {
		_ = emb.rl.Take()
		wp.Submit(func() {
			ret := emb.embedder.Embedding(ctx, inputs[i])
			if ret == "" {
				log.C(ctx).Warnw("Received empty embedding data, ignoring")
				return
			}

			mu.Lock()
			defer mu.Unlock()
			retList = append(retList, ret)
		})
	}

	wp.StopWait() // Wait for all workers to finish

	log.C(ctx).Infow("Successfully completed embedding", "count", len(retList))
	return retList, nil
}
