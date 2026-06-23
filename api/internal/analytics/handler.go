package analytics

// ⊹ ࣪ ˖ analytics API — live listener count + historical stats

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"radioooooo/internal/auth"
	"radioooooo/internal/channel"
)

type Handler struct {
	store    *Store
	channels *channel.Store
	source   ListenerSource
}

func NewHandler(store *Store, channels *channel.Store, source ListenerSource) *Handler {
	return &Handler{store: store, channels: channels, source: source}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-live-listeners",
		Method:      http.MethodGet,
		Path:        "/channels/{id}/listeners/live",
		Summary:     "Get live listener count",
		Tags:        []string{"Analytics"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.liveListeners)

	huma.Register(api, huma.Operation{
		OperationID: "get-channel-stats",
		Method:      http.MethodGet,
		Path:        "/channels/{id}/stats",
		Summary:     "Get historical listener stats for a channel",
		Tags:        []string{"Analytics"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.channelStats)

	huma.Register(api, huma.Operation{
		OperationID: "get-channel-timeseries",
		Method:      http.MethodGet,
		Path:        "/channels/{id}/stats/timeseries",
		Summary:     "Get hourly listener counts for a channel",
		Tags:        []string{"Analytics"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.timeSeries)
}

// --- types ---

type channelIDInput struct {
	ID string `path:"id"`
}

type liveOutput struct {
	Body struct {
		Listeners int `json:"listeners"`
	}
}

type channelStatsInput struct {
	ID   string `path:"id"`
	From string `query:"from" format:"date-time"`
	To   string `query:"to"   format:"date-time"`
}

type channelStatsOutput struct {
	Body ChannelStats
}

type timeSeriesInput struct {
	ID   string `path:"id"`
	From string `query:"from" format:"date-time"`
	To   string `query:"to"   format:"date-time"`
}

type timeSeriesOutput struct {
	Body struct {
		Points []HourlyCount `json:"points"`
	}
}

// --- handlers ---

// ✮⋆‧° polls icecast directly for live count — no db round-trip
func (h *Handler) liveListeners(ctx context.Context, input *channelIDInput) (*liveOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}

	ch, err := h.channels.Get(ctx, input.ID, stationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		return nil, huma.Error500InternalServerError("internal error")
	}

	snapshots, err := h.source.Poll(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to poll listener source")
	}

	count := 0
	for _, s := range snapshots {
		if s.Mount == ch.Mount {
			count++
		}
	}

	out := &liveOutput{}
	out.Body.Listeners = count
	return out, nil
}

func (h *Handler) channelStats(ctx context.Context, input *channelStatsInput) (*channelStatsOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}

	if _, err := h.channels.Get(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		return nil, huma.Error500InternalServerError("internal error")
	}

	from, _ := time.Parse(time.RFC3339, input.From)
	to, _ := time.Parse(time.RFC3339, input.To)
	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}

	current, err := h.store.CurrentListeners(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal error")
	}

	countries, err := h.store.CountryBreakdown(ctx, input.ID, from, to)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal error")
	}

	peak, err := h.store.PeakListeners(ctx, input.ID, from, to)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal error")
	}

	return &channelStatsOutput{Body: ChannelStats{
		CurrentListeners: current,
		PeakListeners:    peak,
		Countries:        countries,
	}}, nil
}

func (h *Handler) timeSeries(ctx context.Context, input *timeSeriesInput) (*timeSeriesOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}

	if _, err := h.channels.Get(ctx, input.ID, stationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("channel not found")
		}
		return nil, huma.Error500InternalServerError("internal error")
	}

	from, _ := time.Parse(time.RFC3339, input.From)
	to, _ := time.Parse(time.RFC3339, input.To)
	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}

	points, err := h.store.TimeSeries(ctx, input.ID, from, to)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal error")
	}

	out := &timeSeriesOutput{}
	out.Body.Points = points
	return out, nil
}

