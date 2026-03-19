package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Annotation struct {
	ID        uuid.UUID       `json:"id"`
	BookID    uuid.UUID       `json:"book_id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Content   string          `json:"content"`
	Page      int             `json:"page"`
	Chapter   string          `json:"chapter,omitempty"`
	UserNote  string          `json:"user_note,omitempty"`
	CharStart *int            `json:"char_start,omitempty"`
	CharEnd   *int            `json:"char_end,omitempty"`
	Rects     json.RawMessage `json:"rects,omitempty"`
	Embedding pgvector.Vector `json:"-"`
	CreatedAt time.Time       `json:"created_at"`

	// BookTitle is populated for search/recall results.
	BookTitle string `json:"book_title,omitempty"`
}

const DuplicateDebounceSeconds = 30

// EmbedText builds the text to embed for semantic search.
// Format: "{book_title} | {chapter} | {content} | {user_note}" (missing fields omitted).
func (a *Annotation) EmbedText(bookTitle string) string {
	parts := make([]string, 0, 4)
	if bookTitle != "" {
		parts = append(parts, bookTitle)
	}
	if a.Chapter != "" {
		parts = append(parts, a.Chapter)
	}
	parts = append(parts, a.Content)
	if a.UserNote != "" {
		parts = append(parts, a.UserNote)
	}
	return strings.Join(parts, " | ")
}

// IsDuplicateOf returns true if other has the same content, page, and type
// and was created within the debounce window.
func (a *Annotation) IsDuplicateOf(other *Annotation) bool {
	if a.Content != other.Content || a.Page != other.Page || a.Type != other.Type {
		return false
	}
	diff := a.CreatedAt.Sub(other.CreatedAt)
	if diff < 0 {
		diff = -diff
	}
	return diff.Seconds() < DuplicateDebounceSeconds
}

// ConflictsWith returns true if the other annotation covers the same character
// range on the same page but has a different type (highlight vs strikethrough).
func (a *Annotation) ConflictsWith(other *Annotation) bool {
	if a.Page != other.Page || a.Type == other.Type {
		return false
	}
	if a.CharStart == nil || a.CharEnd == nil || other.CharStart == nil || other.CharEnd == nil {
		return false
	}
	return *a.CharStart == *other.CharStart && *a.CharEnd == *other.CharEnd
}
