package auth

import "context"

type contextKey string

const stationIDKey contextKey = "stationID"

// withStationID returns a new context carrying the authenticated station ID.
func WithStationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, stationIDKey, id)
}

// stationIDFromContext returns the authenticated station ID and whether one was set.
func StationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(stationIDKey).(string)
	return id, ok
}
