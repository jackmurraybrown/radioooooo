package login

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/matcornic/hermes/v2"
	"radioooooo/internal/auth"
	"radioooooo/internal/notify"
	"radioooooo/internal/user"
)

// Handler handles login, token refresh, logout, and password reset ⋆˙⟡
type Handler struct {
	users    *user.Store
	mailer   notify.Mailer
	secret   string
	frontURL string
}

func NewHandler(users *user.Store, mailer notify.Mailer, secret, frontURL string) *Handler {
	return &Handler{users: users, mailer: mailer, secret: secret, frontURL: frontURL}
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

	huma.Register(api, huma.Operation{
		OperationID:   "forgot-password",
		Method:        http.MethodPost,
		Path:          "/auth/forgot-password",
		Summary:       "Request a password reset email",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.forgotPassword)

	huma.Register(api, huma.Operation{
		OperationID:   "reset-password",
		Method:        http.MethodPost,
		Path:          "/auth/reset-password",
		Summary:       "Reset password with token",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.resetPassword)
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

type forgotPasswordInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}

type resetPasswordInput struct {
	Body struct {
		Token       string `json:"token"       minLength:"1"`
		NewPassword string `json:"newPassword" minLength:"8"`
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

// ⊹ ࣪ ˖ always returns 204 even if email not found — no user enumeration
func (h *Handler) forgotPassword(ctx context.Context, input *forgotPasswordInput) (*struct{}, error) {
	u, _, err := h.users.GetByEmail(ctx, input.Body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("forgot-password: lookup", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	token, err := h.users.CreatePasswordResetToken(ctx, u.ID)
	if err != nil {
		slog.Error("forgot-password: create token", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.frontURL, token)

	html, plain, err := notify.RenderPlatformEmail(hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"someone requested a password reset for your account.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "click the button below to reset your password:",
					Button: hermes.Button{
						Text: "reset password",
						Link: resetURL,
					},
				},
			},
			Outros: []string{
				"this link expires in 1 hour. if you didn't request this, ignore this email.",
			},
		},
	})
	if err != nil {
		slog.Error("forgot-password: render", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	if err := h.mailer.Send(ctx, u.Email, "reset your password", html, plain); err != nil {
		slog.Error("forgot-password: send", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	slog.Info("forgot-password: sent", "to", u.Email)
	return nil, nil
}

func (h *Handler) resetPassword(ctx context.Context, input *resetPasswordInput) (*struct{}, error) {
	if err := h.users.ResetPassword(ctx, input.Body.Token, input.Body.NewPassword); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error400BadRequest("invalid or expired token")
		}
		slog.Error("reset-password: reset", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
