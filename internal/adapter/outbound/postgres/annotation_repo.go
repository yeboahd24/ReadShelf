package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type annotationRepo struct {
	db *pgxpool.Pool
}

func NewAnnotationRepository(db *pgxpool.Pool) outbound.AnnotationRepository {
	return &annotationRepo{db: db}
}

func (r *annotationRepo) Create(ctx context.Context, a *domain.Annotation) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO annotations (id, book_id, user_id, type, content, page, chapter, user_note, char_start, char_end, rects, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		a.ID, a.BookID, a.UserID, a.Type, a.Content, a.Page, a.Chapter, a.UserNote,
		a.CharStart, a.CharEnd, a.Rects, a.CreatedAt,
	)
	return err
}

func (r *annotationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Annotation, error) {
	a := &domain.Annotation{}
	err := r.db.QueryRow(ctx,
		`SELECT id, book_id, user_id, type, content, page, chapter, user_note, char_start, char_end, rects, created_at
		 FROM annotations WHERE id = $1`, id,
	).Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter, &a.UserNote,
		&a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return a, err
}

func (r *annotationRepo) ListByBook(ctx context.Context, bookID, userID uuid.UUID) ([]domain.Annotation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, book_id, user_id, type, content, page, chapter, user_note, char_start, char_end, rects, created_at
		 FROM annotations WHERE book_id = $1 AND user_id = $2 ORDER BY page, created_at`, bookID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var annotations []domain.Annotation
	for rows.Next() {
		var a domain.Annotation
		if err := rows.Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter,
			&a.UserNote, &a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt); err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}
	return annotations, rows.Err()
}

func (r *annotationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM annotations WHERE id = $1`, id)
	return err
}

func (r *annotationRepo) UpdateNote(ctx context.Context, id uuid.UUID, note string) error {
	_, err := r.db.Exec(ctx, `UPDATE annotations SET user_note = $1 WHERE id = $2`, note, id)
	return err
}

func (r *annotationRepo) UpdateEmbedding(ctx context.Context, id uuid.UUID, embedding pgvector.Vector) error {
	_, err := r.db.Exec(ctx, `UPDATE annotations SET embedding = $1 WHERE id = $2`, embedding, id)
	return err
}

func (r *annotationRepo) FindRecent(ctx context.Context, userID, bookID uuid.UUID, content string, page int, annotationType string) (*domain.Annotation, error) {
	a := &domain.Annotation{}
	err := r.db.QueryRow(ctx,
		`SELECT id, book_id, user_id, type, content, page, chapter, user_note, char_start, char_end, rects, created_at
		 FROM annotations
		 WHERE user_id = $1 AND book_id = $2 AND content = $3 AND page = $4 AND type = $5
		   AND created_at > now() - interval '30 seconds'
		 ORDER BY created_at DESC LIMIT 1`,
		userID, bookID, content, page, annotationType,
	).Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter, &a.UserNote,
		&a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *annotationRepo) FindConflicting(ctx context.Context, userID, bookID uuid.UUID, page int, charStart, charEnd int, annotationType string) (*domain.Annotation, error) {
	a := &domain.Annotation{}
	err := r.db.QueryRow(ctx,
		`SELECT id, book_id, user_id, type, content, page, chapter, user_note, char_start, char_end, rects, created_at
		 FROM annotations
		 WHERE user_id = $1 AND book_id = $2 AND page = $3
		   AND char_start = $4 AND char_end = $5 AND type != $6
		 LIMIT 1`,
		userID, bookID, page, charStart, charEnd, annotationType,
	).Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter, &a.UserNote,
		&a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *annotationRepo) SearchFullText(ctx context.Context, userID uuid.UUID, query string) ([]domain.Annotation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT a.id, a.book_id, a.user_id, a.type, a.content, a.page, a.chapter, a.user_note,
		        a.char_start, a.char_end, a.rects, a.created_at, b.title
		 FROM annotations a
		 JOIN books b ON b.id = a.book_id
		 WHERE a.user_id = $1 AND a.tsv @@ plainto_tsquery('english', $2)
		 ORDER BY ts_rank(a.tsv, plainto_tsquery('english', $2)) DESC
		 LIMIT 50`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("full-text search: %w", err)
	}
	defer rows.Close()

	var results []domain.Annotation
	for rows.Next() {
		var a domain.Annotation
		if err := rows.Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter,
			&a.UserNote, &a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt, &a.BookTitle); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

func (r *annotationRepo) SearchVector(ctx context.Context, userID uuid.UUID, embedding pgvector.Vector, limit int) ([]domain.Annotation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT a.id, a.book_id, a.user_id, a.type, a.content, a.page, a.chapter, a.user_note,
		        a.char_start, a.char_end, a.rects, a.created_at, b.title
		 FROM annotations a
		 JOIN books b ON b.id = a.book_id
		 WHERE a.user_id = $1 AND a.embedding IS NOT NULL
		 ORDER BY a.embedding <=> $2
		 LIMIT $3`,
		userID, embedding, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()

	var results []domain.Annotation
	for rows.Next() {
		var a domain.Annotation
		if err := rows.Scan(&a.ID, &a.BookID, &a.UserID, &a.Type, &a.Content, &a.Page, &a.Chapter,
			&a.UserNote, &a.CharStart, &a.CharEnd, &a.Rects, &a.CreatedAt, &a.BookTitle); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, rows.Err()
}
