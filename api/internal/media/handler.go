package media

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
		OperationID:   "create-media",
		Method:        http.MethodPost,
		Path:          "/media",
		Summary:       "Add a media item",
		Tags:          []string{"Media"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-media",
		Method:      http.MethodGet,
		Path:        "/media",
		Summary:     "List media for a station",
		Tags:        []string{"Media"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-media",
		Method:      http.MethodGet,
		Path:        "/media/{id}",
		Summary:     "Get a media item",
		Tags:        []string{"Media"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-media",
		Method:      http.MethodPut,
		Path:        "/media/{id}",
		Summary:     "Update media metadata",
		Tags:        []string{"Media"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-media",
		Method:        http.MethodDelete,
		Path:          "/media/{id}",
		Summary:       "Delete a media item",
		Tags:          []string{"Media"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// --- types ˚₊✧ ---

type idInput struct {
	ID string `path:"id"`
}

type mediaCreateBody struct {
	Title         string  `json:"title"                   minLength:"1" maxLength:"200"`
	Artist        *string `json:"artist,omitempty"        maxLength:"200"`
	ArtworkRef    *string `json:"artworkRef,omitempty"`
	FileFormat    *string `json:"fileFormat,omitempty"    enum:"mp3,aac,m4a"`
	FileSizeBytes *int64  `json:"fileSizeBytes,omitempty"`
	SourceAdapter string  `json:"sourceAdapter"           minLength:"1"`
	SourceRef     string  `json:"sourceRef"               minLength:"1"`
}

type mediaUpdateBody struct {
	Title      string  `json:"title"                minLength:"1" maxLength:"200"`
	Artist     *string `json:"artist,omitempty"     maxLength:"200"`
	ArtworkRef *string `json:"artworkRef,omitempty"`
	FileFormat *string `json:"fileFormat,omitempty" enum:"mp3,aac,m4a"`
}

type createInput struct {
	Body mediaCreateBody
}

type updateInput struct {
	ID   string `path:"id"`
	Body mediaUpdateBody
}

type mediaOutput struct {
	Body Media
}

type mediaListBody struct {
	Media []Media `json:"media"`
}

type listOutput struct {
	Body mediaListBody
}

// --- handlers ✦ ✧ ✦ ---

func (h *Handler) create(ctx context.Context, input *createInput) (*mediaOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	item, err := h.store.Create(ctx, CreateParams{
		StationID:     stationID,
		Title:         input.Body.Title,
		Artist:        input.Body.Artist,
		ArtworkRef:    input.Body.ArtworkRef,
		FileFormat:    input.Body.FileFormat,
		FileSizeBytes: input.Body.FileSizeBytes,
		SourceAdapter: input.Body.SourceAdapter,
		SourceRef:     input.Body.SourceRef,
	})
	if err != nil {
		slog.Error("failed to create media", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &mediaOutput{Body: item}, nil
}

func (h *Handler) list(ctx context.Context, input *struct{}) (*listOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	items, err := h.store.List(ctx, stationID)
	if err != nil {
		slog.Error("failed to list media", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Media = items
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *idInput) (*mediaOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	item, err := h.store.Get(ctx, input.ID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("media not found")
		}
		slog.Error("failed to get media", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &mediaOutput{Body: item}, nil
}

func (h *Handler) update(ctx context.Context, input *updateInput) (*mediaOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	item, err := h.store.Update(ctx, input.ID, stationID, UpdateParams{
		Title:      input.Body.Title,
		Artist:     input.Body.Artist,
		ArtworkRef: input.Body.ArtworkRef,
		FileFormat: input.Body.FileFormat,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("media not found")
		}
		slog.Error("failed to update media", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &mediaOutput{Body: item}, nil
}

func (h *Handler) delete(ctx context.Context, input *idInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("media not found")
		}
		slog.Error("failed to delete media", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}
