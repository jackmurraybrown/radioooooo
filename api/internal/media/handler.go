package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"radioooooo/internal/auth"
	"radioooooo/internal/storage"
)

type Handler struct {
	store *Store
	files storage.Store
	river *river.Client[pgx.Tx]
}

func NewHandler(store *Store, files storage.Store, rc *river.Client[pgx.Tx]) *Handler {
	return &Handler{store: store, files: files, river: rc}
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

// ⊹ ࣪ ˖ file upload — raw chi route, not huma (multipart streaming)
func (h *Handler) RegisterUpload(r chi.Router) {
	r.Post("/media/upload", h.upload)
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
	// ⋆˙⟡ delete file from storage before removing the record
	item, err := h.store.Get(ctx, input.ID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("media not found")
		}
		return nil, huma.Error500InternalServerError("internal error")
	}
	if item.LocalRef != nil && h.files != nil {
		if err := h.files.Delete(ctx, *item.LocalRef); err != nil {
			slog.Warn("media: file delete failed", "path", *item.LocalRef, "error", err)
		}
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

var allowedFormats = map[string]bool{
	".mp3": true,
	".aac": true,
	".m4a": true,
}

// ✮⋆‧° multi-file upload — streams to storage, creates media records
// POST /media/upload with multipart form, field name "files"
// accepts multiple files in one request
func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	stationID, ok := auth.StationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// 500MB max across all files
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "multipart expected", http.StatusBadRequest)
		return
	}

	type uploadResult struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Filename string `json:"filename"`
		Error    string `json:"error,omitempty"`
	}
	var results []uploadResult

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("media: multipart read failed", "error", err)
			break
		}

		if part.FormName() != "files" {
			part.Close()
			continue
		}

		filename := part.FileName()
		ext := strings.ToLower(filepath.Ext(filename))

		if !allowedFormats[ext] {
			results = append(results, uploadResult{
				Filename: filename,
				Error:    fmt.Sprintf("unsupported format: %s (mp3, aac, m4a only)", ext),
			})
			part.Close()
			continue
		}

		// . ݁₊ ✶ stream directly to storage — no temp file
		key := fmt.Sprintf("stations/%s/media/%s%s", stationID, uuid.New().String(), ext)
		if err := h.files.Upload(r.Context(), key, part); err != nil {
			results = append(results, uploadResult{
				Filename: filename,
				Error:    "upload failed",
			})
			slog.Error("media: upload failed", "filename", filename, "error", err)
			part.Close()
			continue
		}
		part.Close()

		title := strings.TrimSuffix(filename, ext)
		format := strings.TrimPrefix(ext, ".")

		item, err := h.store.Create(r.Context(), CreateParams{
			StationID:     stationID,
			Title:         title,
			FileFormat:    &format,
			SourceAdapter: "local",
			SourceRef:     key,
		})
		if err != nil {
			results = append(results, uploadResult{
				Filename: filename,
				Error:    "db record failed",
			})
			slog.Error("media: create record failed", "filename", filename, "error", err)
			continue
		}

		if err := h.store.SetLocalRef(r.Context(), item.ID, key); err != nil {
			slog.Warn("media: set local_ref failed", "id", item.ID, "error", err)
		}

		// ⋆˙⟡ enqueue loudness analysis immediately
		if _, err := h.river.Insert(r.Context(), LoudnessAnalysisArgs{
			MediaID:  item.ID,
			FilePath: h.files.URL(key),
		}, nil); err != nil {
			slog.Warn("media: enqueue loudness failed", "id", item.ID, "error", err)
		}

		results = append(results, uploadResult{
			ID:       item.ID,
			Title:    title,
			Filename: filename,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	type response struct {
		Uploads []uploadResult `json:"uploads"`
	}
	json.NewEncoder(w).Encode(response{Uploads: results})
}
