package show

// ⊹ ࣪ ˖ show CRUD — recurring programme identity

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"radioooooo/internal/auth"
	"radioooooo/internal/station"
)

type Handler struct {
	store    *Store
	stations *station.Store
}

func NewHandler(store *Store, stations *station.Store) *Handler {
	return &Handler{store: store, stations: stations}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-show",
		Method:        http.MethodPost,
		Path:          "/channels/{channelId}/shows",
		Summary:       "Create a show",
		Tags:          []string{"Shows"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-shows",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/shows",
		Summary:     "List shows for a channel",
		Tags:        []string{"Shows"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-show",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/shows/{id}",
		Summary:     "Get a show",
		Tags:        []string{"Shows"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-show",
		Method:      http.MethodPut,
		Path:        "/channels/{channelId}/shows/{id}",
		Summary:     "Update a show",
		Tags:        []string{"Shows"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-show",
		Method:        http.MethodDelete,
		Path:          "/channels/{channelId}/shows/{id}",
		Summary:       "Delete a show and all its episodes",
		Tags:          []string{"Shows"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// --- types ---

type showBody struct {
	Title           string `json:"title"           minLength:"1" maxLength:"200"`
	Description     string `json:"description"`
	RecurrenceRule  string `json:"recurrenceRule"  minLength:"1"`
	DurationMinutes int    `json:"durationMinutes" minimum:"1"`
	Type            string `json:"type"            enum:"live,recorded,external,playlist"`
	SourceAdapter   string `json:"sourceAdapter"   minLength:"1"`
	SourceRef       string `json:"sourceRef"       minLength:"1"`
	AllowRepeat     bool   `json:"allowRepeat"`
}

type channelPathInput struct {
	ChannelID string `path:"channelId"`
}

type showPathInput struct {
	ChannelID string `path:"channelId"`
	ID        string `path:"id"`
}

type createInput struct {
	channelPathInput
	Body showBody
}

type updateInput struct {
	showPathInput
	Body showBody
}

type showOutput struct {
	Body Show
}

type listOutput struct {
	Body struct {
		Shows []Show `json:"shows"`
	}
}

// --- handlers ---

func (h *Handler) create(ctx context.Context, input *createInput) (*showOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	s, err := h.store.Create(ctx, CreateParams{
		ChannelID:       input.ChannelID,
		StationID:       stationID,
		Title:           input.Body.Title,
		Description:     input.Body.Description,
		RecurrenceRule:  input.Body.RecurrenceRule,
		DurationMinutes: input.Body.DurationMinutes,
		Type:            input.Body.Type,
		SourceAdapter:   input.Body.SourceAdapter,
		SourceRef:       input.Body.SourceRef,
		AllowRepeat:     input.Body.AllowRepeat,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		slog.Error("failed to create show", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	// ⊹ ࣪ ˖ expand initial episodes immediately
	tzName, err := h.stations.TimezoneForChannel(ctx, s.ChannelID)
	if err == nil {
		if loc, err := time.LoadLocation(tzName); err == nil {
			count, err := ExpandShow(ctx, h.store, s, loc)
			if err != nil {
				slog.Warn("show created but expansion failed", "show", s.ID, "error", err)
			} else if count > 0 {
				slog.Info("show created, episodes expanded", "show", s.ID, "count", count)
			}
		}
	}

	return &showOutput{Body: s}, nil
}

func (h *Handler) list(ctx context.Context, input *channelPathInput) (*listOutput, error) {
	_, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	shows, err := h.store.List(ctx, input.ChannelID)
	if err != nil {
		slog.Error("failed to list shows", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Shows = shows
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *showPathInput) (*showOutput, error) {
	_, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	s, err := h.store.Get(ctx, input.ID, input.ChannelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("show not found")
		}
		slog.Error("failed to get show", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &showOutput{Body: s}, nil
}

func (h *Handler) update(ctx context.Context, input *updateInput) (*showOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	s, err := h.store.Update(ctx, input.ID, input.ChannelID, stationID, UpdateParams{
		Title:           input.Body.Title,
		Description:     input.Body.Description,
		RecurrenceRule:  input.Body.RecurrenceRule,
		DurationMinutes: input.Body.DurationMinutes,
		Type:            input.Body.Type,
		SourceAdapter:   input.Body.SourceAdapter,
		SourceRef:       input.Body.SourceRef,
		AllowRepeat:     input.Body.AllowRepeat,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("show not found")
		}
		slog.Error("failed to update show", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &showOutput{Body: s}, nil
}

func (h *Handler) delete(ctx context.Context, input *showPathInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID, input.ChannelID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("show not found")
		}
		slog.Error("failed to delete show", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
