package ical

// ⊹ ࣪ ˖ ical feed management — add/list/remove calendar feeds per channel

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

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
		OperationID:   "add-ical-feed",
		Method:        http.MethodPost,
		Path:          "/channels/{channelId}/ical-feeds",
		Summary:       "Add an iCal feed to a channel",
		Tags:          []string{"iCal"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-ical-feeds",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/ical-feeds",
		Summary:     "List iCal feeds for a channel",
		Tags:        []string{"iCal"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-ical-feed",
		Method:        http.MethodDelete,
		Path:          "/ical-feeds/{id}",
		Summary:       "Remove an iCal feed",
		Tags:          []string{"iCal"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

type createInput struct {
	ChannelID string `path:"channelId"`
	Body      struct {
		Type         string  `json:"type"                    enum:"ical,caldav"`
		URL          string  `json:"url"                     minLength:"1"`
		Username     *string `json:"username,omitempty"`
		Password     *string `json:"password,omitempty"`
		CalendarPath *string `json:"calendarPath,omitempty"`
	}
}

type feedOutput struct {
	Body Feed
}

type feedListBody struct {
	Feeds []Feed `json:"feeds"`
}

type listOutput struct {
	Body feedListBody
}

type deleteInput struct {
	ID string `path:"id"`
}

func (h *Handler) create(ctx context.Context, input *createInput) (*feedOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	feedType := input.Body.Type
	if feedType == "" {
		feedType = "ical"
	}
	feed, err := h.store.Create(ctx, CreateParams{
		StationID:    stationID,
		ChannelID:    input.ChannelID,
		Type:         feedType,
		URL:          input.Body.URL,
		Username:     input.Body.Username,
		Password:     input.Body.Password,
		CalendarPath: input.Body.CalendarPath,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		slog.Error("failed to create ical feed", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &feedOutput{Body: feed}, nil
}

func (h *Handler) list(ctx context.Context, input *struct {
	ChannelID string `path:"channelId"`
}) (*listOutput, error) {
	_, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	feeds, err := h.store.List(ctx, input.ChannelID)
	if err != nil {
		slog.Error("failed to list ical feeds", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Feeds = feeds
	return out, nil
}

func (h *Handler) delete(ctx context.Context, input *deleteInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("feed not found")
		}
		slog.Error("failed to delete ical feed", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
