package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"radioooooo/internal/auth"
	"radioooooo/internal/channel"
	"radioooooo/internal/config"
	"radioooooo/internal/episode"
	"radioooooo/internal/login"
	"radioooooo/internal/media"
	"radioooooo/internal/playlist"
	"radioooooo/internal/station"
	"radioooooo/internal/user"
)

type Server struct {
	router http.Handler
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	r := chi.NewMux()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	stationStore := station.NewStore(db)
	channelStore := channel.NewStore(db)
	episodeStore := episode.NewStore(db)
	mediaStore := media.NewStore(db)
	playlistStore := playlist.NewStore(db)
	userStore := user.NewStore(db)

	// authMiddleware runs on every request. tries JWT first (no db hit), then
	// falls back to API key. either way stamps the station id on the context.
	r.Use(authMiddleware(stationStore, cfg.JWTSecret))

	humaConfig := huma.DefaultConfig("Radiooo API", "0.1.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {Type: "http", Scheme: "bearer"},
	}

	api := humachi.New(r, humaConfig)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	login.NewHandler(userStore, cfg.JWTSecret).Register(api)
	station.NewHandler(stationStore, userStore, cfg.JWTSecret).Register(api)
	user.NewHandler(userStore).Register(api)
	channel.NewHandler(channelStore).Register(api)
	episode.NewHandler(episodeStore).Register(api)
	media.NewHandler(mediaStore).Register(api)
	playlist.NewHandler(playlistStore).Register(api)

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// authMiddleware tries JWT first (stateless, no db), then API key (db lookup).
// whichever succeeds stamps the station id on the context.
func authMiddleware(store *station.Store, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != "" {
				if claims, err := auth.ParseAccessToken(jwtSecret, token); err == nil {
					if stationID, ok := claims["station_id"].(string); ok && stationID != "" {
						r = r.WithContext(auth.WithStationID(r.Context(), stationID))
					}
				} else if stationID, err := store.VerifyAPIKey(r.Context(), token); err == nil {
					r = r.WithContext(auth.WithStationID(r.Context(), stationID))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
