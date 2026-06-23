package analytics

// ✮ ⋆ ˚｡𖦹 listener source interface — abstracts where listener data comes from
// icecast today, liquidsoap stats or hls session counting in the future

import "context"

type ListenerSnapshot struct {
	Mount      string
	IP         string
	UserAgent  string
}

type ListenerSource interface {
	Poll(ctx context.Context) ([]ListenerSnapshot, error)
}
