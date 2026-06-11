// package auth handles bearer token authentication for station requests.
//
// two token types are supported, both stamping the station id on the context:
//
//  1. jwt access tokens — issued by POST /auth/login, valid for 15 minutes.
//     the authMiddleware in server.go parses and validates these first (no db hit).
//
//  2. api keys — long-lived, sha-256 hashed in api_keys. for machine-to-machine
//     access (liquidsoap callbacks, external integrations). the middleware falls
//     back to an api key lookup if jwt parsing fails.
//
// handlers call StationIDFromContext to retrieve the authenticated station id.
// if no station id is present the request was unauthenticated — return 403.
package auth

import "context"

type contextKey string

const stationIDKey contextKey = "stationID"

// WithStationID returns a new context carrying the authenticated station ID.
func WithStationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, stationIDKey, id)
}

// StationIDFromContext returns the authenticated station ID and whether one was set.
// handlers use the bool to distinguish unauthenticated requests (return 403) from
// authenticated ones.
func StationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(stationIDKey).(string)
	return id, ok
}
