package outbound

import (
	"context"

	"github.com/pgvector/pgvector-go"
)

type Embedder interface {
	Embed(ctx context.Context, text string) (pgvector.Vector, error)
}
