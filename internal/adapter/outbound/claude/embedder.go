package claude

import (
	"context"

	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/pgvector/pgvector-go"
)

// embedder is a placeholder for Claude API embeddings.
// Replace with actual Claude embedding API calls when available.
type embedder struct {
	apiKey string
}

func NewEmbedder(apiKey string) outbound.Embedder {
	return &embedder{apiKey: apiKey}
}

func (e *embedder) Embed(_ context.Context, _ string) (pgvector.Vector, error) {
	// Placeholder: returns zero vector. Swap to real implementation when using paid tier.
	vec := make([]float32, 768)
	return pgvector.NewVector(vec), nil
}
