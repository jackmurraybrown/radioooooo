package playlist

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
		OperationID:   "create-playlist",
		Method:        http.MethodPost,
		Path:          "/playlists",
		Summary:       "Create a playlist",
		Tags:          []string{"Playlists"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-playlists",
		Method:      http.MethodGet,
		Path:        "/playlists",
		Summary:     "List playlists for a station",
		Tags:        []string{"Playlists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-playlist",
		Method:      http.MethodGet,
		Path:        "/playlists/{id}",
		Summary:     "Get a playlist with its items",
		Tags:        []string{"Playlists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-playlist",
		Method:      http.MethodPut,
		Path:        "/playlists/{id}",
		Summary:     "Update a playlist",
		Tags:        []string{"Playlists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-playlist",
		Method:        http.MethodDelete,
		Path:          "/playlists/{id}",
		Summary:       "Delete a playlist",
		Tags:          []string{"Playlists"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID:   "add-playlist-item",
		Method:        http.MethodPost,
		Path:          "/playlists/{id}/items",
		Summary:       "Add a media item to a playlist",
		Tags:          []string{"Playlists"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.addItem)

	huma.Register(api, huma.Operation{
		OperationID:   "remove-playlist-item",
		Method:        http.MethodDelete,
		Path:          "/playlists/{id}/items/{itemId}",
		Summary:       "Remove an item from a playlist",
		Tags:          []string{"Playlists"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.removeItem)

	huma.Register(api, huma.Operation{
		OperationID: "set-playlist-items",
		Method:      http.MethodPut,
		Path:        "/playlists/{id}/items",
		Summary:     "Replace all playlist items (for reordering)",
		Tags:        []string{"Playlists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.setItems)
}

// --- types ˚₊✧ ---

type idInput struct {
	ID string `path:"id"`
}

type itemInput struct {
	ID     string `path:"id"`
	ItemID string `path:"itemId"`
}

type createBody struct {
	Name          string  `json:"name"                    minLength:"1" maxLength:"200"`
	Shuffle       bool    `json:"shuffle"`
	Loop          bool    `json:"loop"`
	SourceAdapter *string `json:"sourceAdapter,omitempty" minLength:"1"`
	SourceRef     *string `json:"sourceRef,omitempty"     minLength:"1"`
}

type updateBody struct {
	Name    string `json:"name"    minLength:"1" maxLength:"200"`
	Shuffle bool   `json:"shuffle"`
	Loop    bool   `json:"loop"`
}

type addItemBody struct {
	MediaID string `json:"mediaId" minLength:"1"`
}

type setItemsBody struct {
	MediaIDs []string `json:"mediaIds"`
}

type createInput struct {
	Body createBody
}

type updateInput struct {
	ID   string `path:"id"`
	Body updateBody
}

type addItemInput struct {
	ID   string `path:"id"`
	Body addItemBody
}

type setItemsInput struct {
	ID   string `path:"id"`
	Body setItemsBody
}

type playlistOutput struct {
	Body Playlist
}

type playlistWithItemsOutput struct {
	Body struct {
		Playlist Playlist       `json:"playlist"`
		Items    []PlaylistItem `json:"items"`
	}
}

type listOutput struct {
	Body struct {
		Playlists []Playlist `json:"playlists"`
	}
}

type itemOutput struct {
	Body PlaylistItem
}

type itemsOutput struct {
	Body struct {
		Items []PlaylistItem `json:"items"`
	}
}

// --- handlers ✦ ✧ ✦ ---

func (h *Handler) create(ctx context.Context, input *createInput) (*playlistOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	pl, err := h.store.Create(ctx, CreateParams{
		StationID:     stationID,
		Name:          input.Body.Name,
		Shuffle:       input.Body.Shuffle,
		Loop:          input.Body.Loop,
		SourceAdapter: input.Body.SourceAdapter,
		SourceRef:     input.Body.SourceRef,
	})
	if err != nil {
		slog.Error("failed to create playlist", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &playlistOutput{Body: pl}, nil
}

func (h *Handler) list(ctx context.Context, input *struct{}) (*listOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	playlists, err := h.store.List(ctx, stationID)
	if err != nil {
		slog.Error("failed to list playlists", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &listOutput{}
	out.Body.Playlists = playlists
	return out, nil
}

func (h *Handler) get(ctx context.Context, input *idInput) (*playlistWithItemsOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	pl, err := h.store.Get(ctx, input.ID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("playlist not found")
		}
		slog.Error("failed to get playlist", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	items, err := h.store.ListItems(ctx, pl.ID)
	if err != nil {
		slog.Error("failed to list playlist items", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &playlistWithItemsOutput{}
	out.Body.Playlist = pl
	out.Body.Items = items
	return out, nil
}

func (h *Handler) update(ctx context.Context, input *updateInput) (*playlistOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	pl, err := h.store.Update(ctx, input.ID, stationID, UpdateParams{
		Name:    input.Body.Name,
		Shuffle: input.Body.Shuffle,
		Loop:    input.Body.Loop,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("playlist not found")
		}
		slog.Error("failed to update playlist", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &playlistOutput{Body: pl}, nil
}

func (h *Handler) delete(ctx context.Context, input *idInput) (*struct{}, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.Delete(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("playlist not found")
		}
		slog.Error("failed to delete playlist", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}

func (h *Handler) addItem(ctx context.Context, input *addItemInput) (*itemOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	item, err := h.store.AddItem(ctx, input.ID, input.Body.MediaID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("playlist or media not found")
		}
		slog.Error("failed to add playlist item", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &itemOutput{Body: item}, nil
}

func (h *Handler) removeItem(ctx context.Context, input *itemInput) (*struct{}, error) {
	_, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := h.store.RemoveItem(ctx, input.ItemID, input.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("item not found")
		}
		slog.Error("failed to remove playlist item", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	return nil, nil
}

func (h *Handler) setItems(ctx context.Context, input *setItemsInput) (*itemsOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	items, err := h.store.SetItems(ctx, input.ID, stationID, input.Body.MediaIDs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("playlist not found")
		}
		slog.Error("failed to set playlist items", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &itemsOutput{}
	out.Body.Items = items
	return out, nil
}
