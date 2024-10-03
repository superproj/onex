package image

import (
	"context"
	"encoding/json"
	"time"

	"github.com/superproj/onex/internal/pkg/embedding/embedder/onex"
	"github.com/superproj/onex/pkg/log"
)

// EmbeddingData holds the information needed for embedding operations.
// It includes both common embedding data and additional attributes for image embedding.
type EmbeddingData struct {
	// Common embedding data structure.
	Data      onex.EmbeddingData
	ImagePath string
}

// embedder implements the embedding interface for image data.
type embedder struct {
	// Client to interact with the embedding service.
	client *ollama.LLM
}

// NewEmbedder initializes a new embedder with the specified options.
func NewEmbedder(client *ollama.LLM) *embedder {
	return &embedder{client: client}
}

// Embedding performs the embedding operation on the provided input.
func (emb *embedder) Embedding(ctx context.Context, input any) string {
	data, ok := input.(EmbeddingData)
	if !ok {
		log.C(ctx).Warnw("Invalid input type for embedding")
		return ""
	}

	embs, err := emb.client.CreateEmbedding(ctx, []string{data.ImagePath})
	if err != nil {
		log.C(ctx).Warnw("Failed to embed image", "err", err)
		return ""
	}

	if len(embs) != 1 {
		log.C(ctx).Warnw("Embedding output is not equal to 1")
		return ""
	}

	// Store the embedding results.
	data.Data.Emb = embs[0]
	ret, _ := json.Marshal(data)

	return string(ret)
}
