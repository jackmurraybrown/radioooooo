package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"radioooooo/internal/auth"
	"radioooooo/internal/channel"
	"radioooooo/internal/config"
	"radioooooo/internal/episode"
	"radioooooo/internal/media"
	"radioooooo/internal/station"
)

type Server struct {
	router http.Handler
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	r := chi.NewMux()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.StripSlashes)

	stationStore := station.NewStore(db)
	channelStore := channel.NewStore(db)
	episodeStore := episode.NewStore(db)
	mediaStore := media.NewStore(db)

	// apiKeyMiddleware runs on every request. if a valid bearer token is present it
	// populates the context with the authenticated station id. handlers that need
	// auth check for this value themselves via auth.StationIDFromContext.
	r.Use(apiKeyMiddleware(stationStore))

	humaConfig := huma.DefaultConfig("Radiooo API", "0.1.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {Type: "http", Scheme: "bearer"},
	}

	api := humachi.New(r, humaConfig)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	station.NewHandler(stationStore).Register(api)
	channel.NewHandler(channelStore).Register(api)
	episode.NewHandler(episodeStore).Register(api)
	media.NewHandler(mediaStore).Register(api)

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func apiKeyMiddleware(store *station.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != "" {
				stationID, err := store.VerifyAPIKey(r.Context(), token)
				if err == nil {
					r = r.WithContext(auth.WithStationID(r.Context(), stationID))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
