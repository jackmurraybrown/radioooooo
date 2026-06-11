package login

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"radioooooo/internal/auth"
	"radioooooo/internal/user"
)

// Handler handles login, token refresh, and logout.
type Handler struct {
	users  *user.Store
	secret string
}

func NewHandler(users *user.Store, secret string) *Handler {
	return &Handler{users: users, secret: secret}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login with email and password",
		Tags:        []string{"Auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh access token",
		Tags:        []string{"Auth"},
	}, h.refresh)

	huma.Register(api, huma.Operation{
		OperationID:   "logout",
		Method:        http.MethodPost,
		Path:          "/auth/logout",
		Summary:       "Logout and revoke refresh token",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.logout)
}

// --- types ˚₊✧ ---

type loginInput struct {
	Body struct {
		Email    string `json:"email"    format:"email"`
		Password string `json:"password" minLength:"1"`
	}
}

type refreshInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken" minLength:"1"`
	}
}

type logoutInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken" minLength:"1"`
	}
}

type tokenOutput struct {
	Body struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		TokenType    string `json:"tokenType"`
	}
}

// --- handlers ✦ ✧ ✦ ---

func (h *Handler) login(ctx context.Context, input *loginInput) (*tokenOutput, error) {
	u, hash, err := h.users.GetByEmail(ctx, input.Body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error401Unauthorized("invalid credentials")
		}
		slog.Error("login: get user", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	if !user.CheckPassword(hash, input.Body.Password) {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}
	accessToken, err := auth.IssueAccessToken(h.secret, u.ID, u.StationID)
	if err != nil {
		slog.Error("login: issue access token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	refreshToken, err := h.users.CreateRefreshToken(ctx, u.ID)
	if err != nil {
		slog.Error("login: create refresh token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &tokenOutput{}
	out.Body.AccessToken = accessToken
	out.Body.RefreshToken = refreshToken
	out.Body.TokenType = "Bearer"
	return out, nil
}

func (h *Handler) refresh(ctx context.Context, input *refreshInput) (*tokenOutput, error) {
	u, newRefresh, err := h.users.RotateRefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error401Unauthorized("invalid or expired refresh token")
		}
		slog.Error("refresh: rotate token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	accessToken, err := auth.IssueAccessToken(h.secret, u.ID, u.StationID)
	if err != nil {
		slog.Error("refresh: issue access token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &tokenOutput{}
	out.Body.AccessToken = accessToken
	out.Body.RefreshToken = newRefresh
	out.Body.TokenType = "Bearer"
	return out, nil
}

func (h *Handler) logout(ctx context.Context, input *logoutInput) (*struct{}, error) {
	if err := h.users.DeleteRefreshToken(ctx, input.Body.RefreshToken); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("logout: delete token", "error", err)
			return nil, huma.Error500InternalServerError("internal error")
		}
	}
	return nil, nil
}
