package episode

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
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
		OperationID:   "create-episode",
		Method:        http.MethodPost,
		Path:          "/channels/{channelId}/episodes",
		Summary:       "Create an episode",
		Tags:          []string{"Episodes"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-episodes",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/episodes",
		Summary:     "List episodes for a channel",
		Tags:        []string{"Episodes"},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-episode",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/episodes/{id}",
		Summary:     "Get an episode",
		Tags:        []string{"Episodes"},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-episode",
		Method:      http.MethodPut,
		Path:        "/channels/{channelId}/episodes/{id}",
		Summary:     "Update an episode",
		Tags:        []string{"Episodes"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-episode",
		Method:        http.MethodDelete,
		Path:          "/channels/{channelId}/episodes/{id}",
		Summary:       "Delete an episode",
		Tags:          []string{"Episodes"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// --- types ---

type episodeBody struct {
	Title         string    `json:"title"                  minLength:"1" maxLength:"200"`
	Description   string    `json:"description,omitempty"  maxLength:"2000"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Type          string    `json:"type"                   enum:"live,recorded,external,playlist"`
	SourceAdapter string    `json:"sourceAdapter"          minLength:"1"`
	SourceRef     string    `json:"sourceRef"              minLength:"1"`
}

type createInput struct {
	ChannelID string `path:"channelId"`
	Body      episodeBody
}

type updateInput struct {
	ChannelID string `path:"channelId"`
	ID        string `path:"id"`
	Body      episodeBody
}

type channelInput struct {
	ChannelID string `path:"channelId"`
}

type episodeInput struct {
	ChannelID string `path:"channelId"`
	ID        string `path:"id"`
}

type episodeOutput struct {
	Body Episode
}

type episodeListBody struct {
	Episodes []Episode `json:"episodes"`
}

type listOutput struct {
	Body episodeListBody
}

// --- handlers ---

func (h *Handler) create(ctx context.Context, input *createInput) (*episodeOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	ep, err := h.store.Create(ctx, CreateParams{
		ChannelID:     input.ChannelID,
		StationID:     stationID,
		Title:         input.Body.Title,
		Description:   input.Body.Description,
		StartTime:     input.Body.StartTime,
		EndTime:       input.Body.EndTime,
		Type:          input.Body.Type,
		SourceAdapter: input.Body.SourceAdapter,
		SourceRef:     input.Body.SourceRef,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		slog.Error("failed to create episode", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &episodeOutput{Body: ep}, nil
}

func (h *Handler) list(ctx context.Context, input *channelInput) (*listOutput, error) {
	episodes, err := h.store.List(ctx, input.ChannelID)
	if err != nil {
		slog.Error("failed to list episodes", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Episodes = episodes
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *episodeInput) (*episodeOutput, error) {
	ep, err := h.store.Get(ctx, input.ID, input.ChannelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("episode not found")
		}
		slog.Error("failed to get episode", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &episodeOutput{Body: ep}, nil
}

func (h *Handler) update(ctx context.Context, input *updateInput) (*episodeOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	ep, err := h.store.Update(ctx, input.ID, input.ChannelID, stationID, UpdateParams{
		Title:         input.Body.Title,
		Description:   input.Body.Description,
		StartTime:     input.Body.StartTime,
		EndTime:       input.Body.EndTime,
		Type:          input.Body.Type,
		SourceAdapter: input.Body.SourceAdapter,
		SourceRef:     input.Body.SourceRef,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("episode not found")
		}
		slog.Error("failed to update episode", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &episodeOutput{Body: ep}, nil
}

func (h *Handler) delete(ctx context.Context, input *episodeInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID, input.ChannelID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("episode not found")
		}
		slog.Error("failed to delete episode", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
