package domain

import (
	"time"

	"github.com/google/uuid"
)

const MaxBooksFree = 3

type Book struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Author    string    `json:"author,omitempty"`
	FileKey   string    `json:"-"`
	PageCount int       `json:"page_count,omitempty"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
