package livenow

// : ) live now — SSE endpoint pushing current on-air state per channel
// public-facing, no auth required. used by the embeddable player.

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"radioooooo/internal/episode"
)

type State struct {
	Show *ShowInfo `json:"show"`
}

type ShowInfo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Type      string    `json:"type"`
}

type Handler struct {
	episodes      *episode.Store
	nowCache      *cache[State]
	scheduleCache *cache[[]episode.Episode]
}

func NewHandler(episodes *episode.Store) *Handler {
	return &Handler{
		episodes:      episodes,
		nowCache:      newCache[State](5 * time.Second),
		scheduleCache: newCache[[]episode.Episode](60 * time.Second),
	}
}

// registers as raw chi routes (SSE + public REST — not behind auth)
func (h *Handler) Register(r chi.Router) {
	r.Get("/channels/{id}/live", h.stream)
	r.Get("/channels/{id}/live/now", h.now)
	r.Get("/channels/{id}/schedule", h.schedule)
	r.Get("/channels/{id}/schedule/range", h.scheduleRange)
}

func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	var lastID string

	// . ݁₊ ✶ poll every 2s, push state when it changes
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// send initial state immediately
	h.sendState(ctx, w, flusher, channelID, &lastID)

	for {
		select {
		case <-ticker.C:
			h.sendState(ctx, w, flusher, channelID, &lastID)
		case <-ctx.Done():
			return
		}
	}
}

func (h *Handler) sendState(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, channelID string, lastID *string) {
	ep, err := h.episodes.GetCurrent(ctx, channelID)

	var state State
	currentID := ""

	if err == nil {
		state.Show = &ShowInfo{
			ID:        ep.ID,
			Title:     ep.Title,
			StartTime: ep.StartTime,
			EndTime:   ep.EndTime,
			Type:      ep.Type,
		}
		currentID = ep.ID
	}

	if currentID == *lastID {
		return
	}
	*lastID = currentID

	data, err := json.Marshal(state)
	if err != nil {
		slog.Error("livenow: marshal failed", "error", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// one-shot current state — for clients that don't need SSE ✮ ⋆ ˚
func (h *Handler) now(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")

	if cached, ok := h.nowCache.get(channelID); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(cached)
		return
	}

	ep, err := h.episodes.GetCurrent(r.Context(), channelID)

	var state State
	if err == nil {
		state.Show = &ShowInfo{
			ID:        ep.ID,
			Title:     ep.Title,
			StartTime: ep.StartTime,
			EndTime:   ep.EndTime,
			Type:      ep.Type,
		}
	}

	h.nowCache.set(channelID, state)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(state)
}

// ⋆˙upcoming episodes for the embeddable player
func (h *Handler) schedule(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")

	type scheduleResponse struct {
		Episodes []episode.Episode `json:"episodes"`
	}

	if cached, ok := h.scheduleCache.get(channelID); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(scheduleResponse{Episodes: cached})
		return
	}

	episodes, err := h.episodes.ListUpcoming(r.Context(), channelID, 20)
	resp := scheduleResponse{Episodes: episodes}
	if err != nil {
		slog.Error("livenow: schedule query failed", "error", err)
		resp.Episodes = []episode.Episode{}
	} else {
		h.scheduleCache.set(channelID, episodes)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}

// : ) episodes between start and end — ?start=2026-06-24T00:00:00Z&end=2026-06-25T00:00:00Z
func (h *Handler) scheduleRange(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")

	start, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	if err != nil {
		http.Error(w, "start must be RFC3339", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	if err != nil {
		http.Error(w, "end must be RFC3339", http.StatusBadRequest)
		return
	}

	episodes, err := h.episodes.ListRange(r.Context(), channelID, start, end)

	type rangeResponse struct {
		Episodes []episode.Episode `json:"episodes"`
	}
	resp := rangeResponse{Episodes: episodes}
	if err != nil {
		slog.Error("livenow: range query failed", "error", err)
		resp.Episodes = []episode.Episode{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}
