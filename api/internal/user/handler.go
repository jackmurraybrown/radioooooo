package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
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

	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/users/me",
		Summary:     "Get the authenticated user",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.me)

	huma.Register(api, huma.Operation{
		OperationID: "change-password",
		Method:      http.MethodPut,
		Path:        "/users/me/password",
		Summary:     "Change password",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.changePassword)
}

// --- types ˚₊✧ ---

type inviteBody struct {
	Email    string `json:"email"    format:"email"`
	Password string `json:"password" minLength:"8"`
}

type createInput struct {
	Body inviteBody
}

type changePasswordBody struct {
	CurrentPassword string `json:"currentPassword" minLength:"1"`
	NewPassword     string `json:"newPassword"     minLength:"8"`
}

type userListBody struct {
	Users []User `json:"users"`
}

type changePasswordInput struct {
	Body changePasswordBody
}

type userOutput struct {
	Body User
}

type listOutput struct {
	Body userListBody
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

func (h *Handler) me(ctx context.Context, _ *struct{}) (*userOutput, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	stationID, _ := auth.StationIDFromContext(ctx)
	u, err := h.store.GetByID(ctx, userID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("user not found")
		}
		slog.Error("failed to get user", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &userOutput{Body: u}, nil
}

func (h *Handler) changePassword(ctx context.Context, input *changePasswordInput) (*struct{}, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	stationID, _ := auth.StationIDFromContext(ctx)
	err := h.store.UpdatePassword(ctx, userID, stationID, input.Body.CurrentPassword, input.Body.NewPassword)
	if err != nil {
		if errors.Is(err, errWrongPassword) {
			return nil, huma.Error422UnprocessableEntity("current password is incorrect")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("user not found")
		}
		slog.Error("failed to update password", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
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
