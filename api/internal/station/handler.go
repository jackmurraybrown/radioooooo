package station

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"radioooooo/internal/auth"
	"radioooooo/internal/user"
)

type Handler struct {
	store  *Store
	users  *user.Store
	secret string
}

func NewHandler(store *Store, users *user.Store, secret string) *Handler {
	return &Handler{store: store, users: users, secret: secret}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-station",
		Method:        http.MethodPost,
		Path:          "/stations",
		Summary:       "Create a station and admin user (signup)",
		Tags:          []string{"Stations"},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-stations",
		Method:      http.MethodGet,
		Path:        "/stations",
		Summary:     "List all stations",
		Tags:        []string{"Stations"},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-station",
		Method:      http.MethodGet,
		Path:        "/stations/{id}",
		Summary:     "Get a station by ID",
		Tags:        []string{"Stations"},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-station",
		Method:        http.MethodDelete,
		Path:          "/stations/{id}",
		Summary:       "Delete a station",
		Tags:          []string{"Stations"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// --- types ---

type createInput struct {
	Body struct {
		Name     string `json:"name"     minLength:"1" maxLength:"100" doc:"display name of the station"`
		Slug     string `json:"slug"     minLength:"1" maxLength:"50"  pattern:"^[a-z0-9-]+$" doc:"url-safe identifier"`
		Email    string `json:"email"    format:"email"                doc:"admin user email"`
		Password string `json:"password" minLength:"8"                 doc:"admin user password"`
	}
}

type createOutput struct {
	Body struct {
		Station      Station `json:"station"`
		APIKey       string  `json:"apiKey"       doc:"store securely — shown only once"`
		AccessToken  string  `json:"accessToken"`
		RefreshToken string  `json:"refreshToken"`
		TokenType    string  `json:"tokenType"`
	}
}

type listOutput struct {
	Body struct {
		Stations []Station `json:"stations"`
	}
}

type getInput struct {
	ID string `path:"id"`
}

type getOutput struct {
	Body Station
}

type deleteInput struct {
	ID string `path:"id"`
}

// --- handlers ---

func (h *Handler) create(ctx context.Context, input *createInput) (*createOutput, error) {
	st, err := h.store.Create(ctx, input.Body.Name, input.Body.Slug)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("slug already taken")
		}
		slog.Error("failed to create station", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	u, err := h.users.Create(ctx, st.ID, input.Body.Email, input.Body.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("email already in use")
		}
		slog.Error("failed to create admin user", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	apiKey, err := h.store.CreateAPIKey(ctx, st.ID)
	if err != nil {
		slog.Error("failed to create api key", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	accessToken, err := auth.IssueAccessToken(h.secret, u.ID, st.ID)
	if err != nil {
		slog.Error("failed to issue access token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	refreshToken, err := h.users.CreateRefreshToken(ctx, u.ID)
	if err != nil {
		slog.Error("failed to create refresh token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	out := &createOutput{}
	out.Body.Station = st
	out.Body.APIKey = apiKey
	out.Body.AccessToken = accessToken
	out.Body.RefreshToken = refreshToken
	out.Body.TokenType = "Bearer"
	return out, nil
}

func (h *Handler) list(ctx context.Context, _ *struct{}) (*listOutput, error) {
	stations, err := h.store.List(ctx)
	if err != nil {
		slog.Error("failed to list stations", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Stations = stations
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *getInput) (*getOutput, error) {
	st, err := h.store.Get(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("station not found")
		}
		slog.Error("failed to get station", "error", err, "id", input.ID)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &getOutput{Body: st}, nil
}

func (h *Handler) delete(ctx context.Context, input *deleteInput) (*struct{}, error) {
	authedID, ok := auth.StationIDFromContext(ctx)
	if !ok || authedID != input.ID {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("station not found")
		}
		slog.Error("failed to delete station", "error", err, "id", input.ID)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
