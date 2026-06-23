package storage

// ⋆˙⟡ file storage interface

import (
	"context"
	"io"
)

type Store interface {
	Upload(ctx context.Context, key string, r io.Reader) error
	Download(ctx context.Context, key string, w io.Writer) error
	Delete(ctx context.Context, key string) error
	URL(key string) string
}
