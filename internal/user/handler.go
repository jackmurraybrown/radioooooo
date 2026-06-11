package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"radioooooo/internal/auth"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-user",
		Method:        http.MethodPost,
		Path:          "/users",
		Summary:       "Invite a user to the station",
		Tags:          []string{"Users"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List users for the station",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)
}

// --- types ˚₊✧ ---

type createInput struct {
	Body struct {
		Email    string `json:"email"    format:"email"`
		Password string `json:"password" minLength:"8"`
	}
}

type userOutput struct {
	Body User
}

type listOutput struct {
	Body struct {
		Users []User `json:"users"`
	}
}

// --- handlers ✦ ✧ ✦ ---

func (h *Handler) create(ctx context.Context, input *createInput) (*userOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	u, err := h.store.Create(ctx, stationID, input.Body.Email, input.Body.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("email already in use")
		}
		slog.Error("failed to create user", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &userOutput{Body: u}, nil
}

func (h *Handler) list(ctx context.Context, _ *struct{}) (*listOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	users, err := h.store.ListByStation(ctx, stationID)
	if err != nil {
		slog.Error("failed to list users", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Users = users
	return out, nil
}
