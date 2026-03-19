package outbound

import (
	"context"
	"io"
)

type FileStore interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	SignedURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}
