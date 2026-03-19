package outbound

import (
	"context"

	"github.com/dominic/readshelf/internal/core/domain"
)

type AIClient interface {
	Recall(ctx context.Context, query string, passages []domain.Annotation) (string, error)
}
