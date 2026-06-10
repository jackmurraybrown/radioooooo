// package auth handles API key authentication for station requests.
//
// how it works:
//
//  1. when a station is created, an API key is generated and returned once
//     (plain text, never stored — only its SHA-256 hash lives in api_keys).
//
//  2. the frontend stores the key (e.g. in localStorage or a cookie) and sends
//     it on every request as a bearer token:
//     Authorization: Bearer <api-key>
//
//  3. the apiKeyMiddleware in server.go intercepts every request, hashes the
//     token, looks it up in api_keys, and — if valid — stamps the station ID
//     onto the request context via WithStationID.
//
//  4. handlers that require auth call StationIDFromContext. if no station ID is
//     present the request was unauthenticated and the handler returns 403.
//
// future: JWT auth with user roles (super admin / station admin / DJ)
// will replace or sit alongside this. the API key model stays for machine-to-machine
// access (e.g. Liquidsoap callbacks, external integrations).
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
