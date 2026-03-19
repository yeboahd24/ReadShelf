package inbound

import (
	"context"
	"io"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/google/uuid"
)

type BookService interface {
	Upload(ctx context.Context, userID uuid.UUID, title, author string, file io.Reader, filename string) (*domain.Book, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.Book, error)
	Get(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) (*domain.Book, error)
	GetSignedURL(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) (string, error)
	Delete(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) error
}
