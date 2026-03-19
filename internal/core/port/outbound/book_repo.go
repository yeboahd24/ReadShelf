package outbound

import (
	"context"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/google/uuid"
)

type BookRepository interface {
	Create(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Book, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
