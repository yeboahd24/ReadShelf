package service

import (
	"context"

	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/google/uuid"
)

type recallService struct {
	annotations outbound.AnnotationRepository
	embedder    outbound.Embedder
	ai          outbound.AIClient
}

func NewRecallService(
	annotations outbound.AnnotationRepository,
	embedder outbound.Embedder,
	ai outbound.AIClient,
) inbound.RecallService {
	return &recallService{
		annotations: annotations,
		embedder:    embedder,
		ai:          ai,
	}
}

func (s *recallService) Query(ctx context.Context, userID uuid.UUID, query string) (*inbound.RecallResult, error) {
	// Embed the query.
	qVec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	// Vector search for top 5 similar annotations.
	matches, err := s.annotations.SearchVector(ctx, userID, qVec, 5)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return &inbound.RecallResult{
			Answer:  "No relevant annotations found.",
			Sources: nil,
		}, nil
	}

	// Ask Claude to synthesise an answer.
	answer, err := s.ai.Recall(ctx, query, matches)
	if err != nil {
		return nil, err
	}

	sources := make([]inbound.RecallSource, len(matches))
	for i, m := range matches {
		sources[i] = inbound.RecallSource{
			AnnotationID: m.ID.String(),
			BookTitle:    m.BookTitle,
			Page:         m.Page,
			Content:      m.Content,
		}
	}

	return &inbound.RecallResult{
		Answer:  answer,
		Sources: sources,
	}, nil
}
