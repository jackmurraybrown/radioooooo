package storage

// ✮⋆‧° local filesystem storage 

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create root dir: %w", err)
	}
	return &Local{root: root}, nil
}

func (l *Local) Upload(_ context.Context, key string, r io.Reader) error {
	path := filepath.Join(l.root, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("storage: mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("storage: write file: %w", err)
	}
	return nil
}

func (l *Local) Download(_ context.Context, key string, w io.Writer) error {
	f, err := os.Open(filepath.Join(l.root, key))
	if err != nil {
		return fmt.Errorf("storage: open file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("storage: read file: %w", err)
	}
	return nil
}

func (l *Local) Delete(_ context.Context, key string) error {
	if err := os.Remove(filepath.Join(l.root, key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete file: %w", err)
	}
	return nil
}

func (l *Local) URL(key string) string {
	return filepath.Join(l.root, key)
}
