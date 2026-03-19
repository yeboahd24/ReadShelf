package postgres

import (
	"context"
	"errors"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookRepo struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) outbound.BookRepository {
	return &bookRepo{db: db}
}

func (r *bookRepo) Create(ctx context.Context, b *domain.Book) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO books (id, user_id, title, author, file_key, page_count, color, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		b.ID, b.UserID, b.Title, b.Author, b.FileKey, b.PageCount, b.Color, b.CreatedAt,
	)
	return err
}

func (r *bookRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	b := &domain.Book{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, title, author, file_key, page_count, color, created_at
		 FROM books WHERE id = $1`, id,
	).Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.FileKey, &b.PageCount, &b.Color, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return b, err
}

func (r *bookRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Book, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, title, author, file_key, page_count, color, created_at
		 FROM books WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var b domain.Book
		if err := rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.FileKey, &b.PageCount, &b.Color, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func (r *bookRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM books WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *bookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	return err
}
