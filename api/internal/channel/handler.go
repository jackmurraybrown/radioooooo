package channel

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
		OperationID:   "create-channel",
		Method:        http.MethodPost,
		Path:          "/channels",
		Summary:       "Create a channel",
		Tags:          []string{"Channels"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-channels",
		Method:      http.MethodGet,
		Path:        "/channels",
		Summary:     "List channels for a station",
		Tags:        []string{"Channels"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-channel",
		Method:      http.MethodGet,
		Path:        "/channels/{id}",
		Summary:     "Get a channel",
		Tags:        []string{"Channels"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-channel",
		Method:        http.MethodDelete,
		Path:          "/channels/{id}",
		Summary:       "Delete a channel",
		Tags:          []string{"Channels"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// --- types ---

type channelBody struct {
	Name string `json:"name" minLength:"1" maxLength:"100"`
	Slug string `json:"slug" minLength:"1" maxLength:"50" pattern:"^[a-z0-9-]+$"`
}

type channelListBody struct {
	Channels []Channel `json:"channels"`
}

type createInput struct {
	Body channelBody
}

type idInput struct {
	ID string `path:"id"`
}

type channelOutput struct {
	Body Channel
}

type listOutput struct {
	Body channelListBody
}

// --- handlers ---

func (h *Handler) create(ctx context.Context, input *createInput) (*channelOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	ch, err := h.store.Create(ctx, stationID, input.Body.Name, input.Body.Slug)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, huma.Error409Conflict("slug already taken")
		}
		slog.Error("failed to create channel", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &channelOutput{Body: ch}, nil
}

func (h *Handler) list(ctx context.Context, input *struct{}) (*listOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	channels, err := h.store.List(ctx, stationID)
	if err != nil {
		slog.Error("failed to list channels", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Channels = channels
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *idInput) (*channelOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	ch, err := h.store.Get(ctx, input.ID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		slog.Error("failed to get channel", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &channelOutput{Body: ch}, nil
}

func (h *Handler) delete(ctx context.Context, input *idInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	channels, err := h.store.List(ctx, stationID)
	if err != nil {
		slog.Error("failed to list channels", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	if len(channels) <= 1 {
		return nil, huma.Error409Conflict("cannot delete the last channel")
	}
	if err := h.store.Delete(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		slog.Error("failed to delete channel", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
