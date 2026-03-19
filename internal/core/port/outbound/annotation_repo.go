package outbound

import (
	"context"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type AnnotationRepository interface {
	Create(ctx context.Context, a *domain.Annotation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Annotation, error)
	ListByBook(ctx context.Context, bookID, userID uuid.UUID) ([]domain.Annotation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateNote(ctx context.Context, id uuid.UUID, note string) error
	UpdateEmbedding(ctx context.Context, id uuid.UUID, embedding pgvector.Vector) error

	// FindRecent returns annotations matching content+page+type within the debounce window.
	FindRecent(ctx context.Context, userID, bookID uuid.UUID, content string, page int, annotationType string) (*domain.Annotation, error)

	// FindConflicting returns an annotation on the same char range/page with a different type.
	FindConflicting(ctx context.Context, userID, bookID uuid.UUID, page int, charStart, charEnd int, annotationType string) (*domain.Annotation, error)

	// SearchFullText performs tsvector full-text search across a user's annotations.
	SearchFullText(ctx context.Context, userID uuid.UUID, query string) ([]domain.Annotation, error)

	// SearchVector performs pgvector cosine similarity search.
	SearchVector(ctx context.Context, userID uuid.UUID, embedding pgvector.Vector, limit int) ([]domain.Annotation, error)
}
