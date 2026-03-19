package service

import (
	"context"
	"log"
	"time"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/google/uuid"
)

type annotationService struct {
	annotations outbound.AnnotationRepository
	books       outbound.BookRepository
	embedder    outbound.Embedder
}

func NewAnnotationService(
	annotations outbound.AnnotationRepository,
	books outbound.BookRepository,
	embedder outbound.Embedder,
) inbound.AnnotationService {
	return &annotationService{
		annotations: annotations,
		books:       books,
		embedder:    embedder,
	}
}

func (s *annotationService) Create(ctx context.Context, a *domain.Annotation) (*domain.Annotation, error) {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()

	// Duplicate debounce: same content + page + type within 30s → reject.
	recent, _ := s.annotations.FindRecent(ctx, a.UserID, a.BookID, a.Content, a.Page, a.Type)
	if recent != nil && a.IsDuplicateOf(recent) {
		return nil, domain.ErrDuplicateAnnotation
	}

	// Conflict resolution: highlight and strikethrough on same char range — later wins.
	if a.CharStart != nil && a.CharEnd != nil {
		conflicting, _ := s.annotations.FindConflicting(ctx, a.UserID, a.BookID, a.Page, *a.CharStart, *a.CharEnd, a.Type)
		if conflicting != nil {
			if err := s.annotations.Delete(ctx, conflicting.ID); err != nil {
				return nil, err
			}
		}
	}

	if err := s.annotations.Create(ctx, a); err != nil {
		return nil, err
	}

	// Async embedding — don't block the response.
	go s.embedAnnotation(a)

	return a, nil
}

func (s *annotationService) ListByBook(ctx context.Context, userID, bookID uuid.UUID) ([]domain.Annotation, error) {
	return s.annotations.ListByBook(ctx, bookID, userID)
}

func (s *annotationService) Delete(ctx context.Context, userID uuid.UUID, annotationID uuid.UUID) error {
	a, err := s.annotations.GetByID(ctx, annotationID)
	if err != nil {
		return err
	}
	if a.UserID != userID {
		return domain.ErrNotFound
	}
	return s.annotations.Delete(ctx, annotationID)
}

func (s *annotationService) UpdateNote(ctx context.Context, userID uuid.UUID, annotationID uuid.UUID, note string) error {
	a, err := s.annotations.GetByID(ctx, annotationID)
	if err != nil {
		return err
	}
	if a.UserID != userID {
		return domain.ErrNotFound
	}

	if err := s.annotations.UpdateNote(ctx, annotationID, note); err != nil {
		return err
	}

	// Re-embed with updated note.
	a.UserNote = note
	go s.embedAnnotation(a)

	return nil
}

func (s *annotationService) Search(ctx context.Context, userID uuid.UUID, query string) ([]domain.Annotation, error) {
	return s.annotations.SearchFullText(ctx, userID, query)
}

func (s *annotationService) embedAnnotation(a *domain.Annotation) {
	// Fetch book title for embed text composition.
	book, err := s.books.GetByID(context.Background(), a.BookID)
	if err != nil {
		log.Printf("embed: failed to get book %s: %v", a.BookID, err)
		return
	}

	text := a.EmbedText(book.Title)
	vec, err := s.embedder.Embed(context.Background(), text)
	if err != nil {
		log.Printf("embed: failed to embed annotation %s: %v", a.ID, err)
		return
	}

	if err := s.annotations.UpdateEmbedding(context.Background(), a.ID, vec); err != nil {
		log.Printf("embed: failed to save embedding for annotation %s: %v", a.ID, err)
	}
}
