package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/google/uuid"
)

type bookService struct {
	books outbound.BookRepository
	files outbound.FileStore
}

func NewBookService(books outbound.BookRepository, files outbound.FileStore) inbound.BookService {
	return &bookService{books: books, files: files}
}

func (s *bookService) Upload(ctx context.Context, userID uuid.UUID, title, author string, file io.Reader, filename string) (*domain.Book, error) {
	count, err := s.books.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxBooksFree {
		return nil, domain.ErrBookLimitReached
	}

	bookID := uuid.New()
	fileKey := fmt.Sprintf("users/%s/books/%s/%s", userID, bookID, filename)

	if err := s.files.Upload(ctx, fileKey, file, "application/pdf"); err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	book := &domain.Book{
		ID:        bookID,
		UserID:    userID,
		Title:     title,
		Author:    author,
		FileKey:   fileKey,
		CreatedAt: time.Now(),
	}

	if err := s.books.Create(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) List(ctx context.Context, userID uuid.UUID) ([]domain.Book, error) {
	return s.books.ListByUser(ctx, userID)
}

func (s *bookService) Get(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) (*domain.Book, error) {
	book, err := s.books.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if book.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return book, nil
}

func (s *bookService) GetSignedURL(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) (string, error) {
	book, err := s.Get(ctx, userID, bookID)
	if err != nil {
		return "", err
	}
	return s.files.SignedURL(ctx, book.FileKey)
}

func (s *bookService) Delete(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) error {
	book, err := s.Get(ctx, userID, bookID)
	if err != nil {
		return err
	}

	if err := s.files.Delete(ctx, book.FileKey); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return s.books.Delete(ctx, bookID)
}
