package inbound

import (
	"context"

	"github.com/google/uuid"
)

type RecallSource struct {
	AnnotationID string `json:"annotation_id"`
	BookID       string `json:"book_id"`
	BookTitle    string `json:"book_title"`
	Page         int    `json:"page"`
	Content      string `json:"content"`
}

type RecallResult struct {
	Answer  string         `json:"answer"`
	Sources []RecallSource `json:"sources"`
}

type RecallService interface {
	Query(ctx context.Context, userID uuid.UUID, query string) (*RecallResult, error)
}
