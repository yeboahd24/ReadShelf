package inbound

import (
	"context"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/google/uuid"
)

type AnnotationService interface {
	Create(ctx context.Context, annotation *domain.Annotation) (*domain.Annotation, error)
	ListByBook(ctx context.Context, userID, bookID uuid.UUID) ([]domain.Annotation, error)
	Delete(ctx context.Context, userID uuid.UUID, annotationID uuid.UUID) error
	UpdateNote(ctx context.Context, userID uuid.UUID, annotationID uuid.UUID, note string) error
	Search(ctx context.Context, userID uuid.UUID, query string) ([]domain.Annotation, error)
}
