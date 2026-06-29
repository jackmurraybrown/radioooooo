package tracklist

// ⊹ ࣪ ˖ ၊၊||၊ tracklist handler — token-auth public endpoints + authenticated CRUD

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	frontURL string
}

func NewHandler(store *Store, channels *channel.Store, frontURL string) *Handler {
	return &Handler{store: store, channels: channels, frontURL: frontURL}
}

func (h *Handler) Register(api huma.API) {
	// ⋆˙⟡ public token-auth
	huma.Register(api, huma.Operation{
		OperationID: "get-tracklist-by-token",
		Method:      http.MethodGet,
		Path:        "/tracklists/{token}",
		Summary:     "Get tracklist via submission link",
		Tags:        []string{"Tracklists"},
	}, h.getByToken)

	huma.Register(api, huma.Operation{
		OperationID: "save-tracklist-by-token",
		Method:      http.MethodPut,
		Path:        "/tracklists/{token}",
		Summary:     "Save tracklist via submission link",
		Tags:        []string{"Tracklists"},
	}, h.saveByToken)

	// ✮ ⋆ ˚ authenticated
	huma.Register(api, huma.Operation{
		OperationID: "list-episode-tracks",
		Method:      http.MethodGet,
		Path:        "/channels/{channelId}/episodes/{episodeId}/tracks",
		Summary:     "List tracks for an episode",
		Tags:        []string{"Tracklists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listTracks)

	huma.Register(api, huma.Operation{
		OperationID: "set-episode-tracks",
		Method:      http.MethodPut,
		Path:        "/channels/{channelId}/episodes/{episodeId}/tracks",
		Summary:     "Replace tracks for an episode",
		Tags:        []string{"Tracklists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.setTracks)

	huma.Register(api, huma.Operation{
		OperationID: "create-submission-link",
		Method:      http.MethodPost,
		Path:        "/channels/{channelId}/episodes/{episodeId}/submission-link",
		Summary:     "Generate a submission link for an episode",
		Tags:        []string{"Tracklists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.createSubmissionLink)

	huma.Register(api, huma.Operation{
		OperationID: "receive-tracklist-webhook",
		Method:      http.MethodPost,
		Path:        "/webhooks/tracklist",
		Summary:     "Receive tracklist from external app",
		Tags:        []string{"Tracklists"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.receiveWebhook)
}

// --- types ✮⋆‧°—°‧⋆✮ ---

type tokenInput struct {
	Token string `path:"token"`
}

type tracklistOutput struct {
	Body struct {
		Episode EpisodeInfo `json:"episode"`
		Tracks  []Track     `json:"tracks"`
	}
}

type saveByTokenInput struct {
	Token string `path:"token"`
	Body  struct {
		Tracks []TrackInput `json:"tracks"`
	}
}

type episodePathInput struct {
	ChannelID string `path:"channelId"`
	EpisodeID string `path:"episodeId"`
}

type tracksOnlyOutput struct {
	Body struct {
		Tracks []Track `json:"tracks"`
	}
}

type setTracksInput struct {
	ChannelID string `path:"channelId"`
	EpisodeID string `path:"episodeId"`
	Body      struct {
		Tracks []TrackInput `json:"tracks"`
	}
}

type submissionLinkOutput struct {
	Body struct {
		URL string `json:"url"`
	}
}

type webhookInput struct {
	Body struct {
		ChannelID string       `json:"channelId" minLength:"1"`
		StartTime time.Time    `json:"startTime"`
		Tracks    []TrackInput `json:"tracks"`
	}
}

// --- public token handlers ⋆˙⟡ ---

func (h *Handler) getByToken(ctx context.Context, input *tokenInput) (*tracklistOutput, error) {
	episodeID, err := h.store.ValidateToken(ctx, input.Token)
	if err != nil {
		return nil, huma.Error404NotFound("invalid or expired link")
	}
	info, _, err := h.store.EpisodeWithStation(ctx, episodeID)
	if err != nil {
		return nil, huma.Error404NotFound("episode not found")
	}
	tracks, err := h.store.ListTracks(ctx, episodeID)
	if err != nil {
		slog.Error("failed to list tracks", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &tracklistOutput{}
	out.Body.Episode = info
	out.Body.Tracks = tracks
	return out, nil
}

func (h *Handler) saveByToken(ctx context.Context, input *saveByTokenInput) (*tracklistOutput, error) {
	episodeID, err := h.store.ValidateToken(ctx, input.Token)
	if err != nil {
		return nil, huma.Error404NotFound("invalid or expired link")
	}
	tracks, err := h.store.SetTracks(ctx, episodeID, input.Body.Tracks)
	if err != nil {
		slog.Error("failed to save tracks", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	info, webhookURL, err := h.store.EpisodeWithStation(ctx, episodeID)
	if err != nil {
		slog.Error("failed to get episode info", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	if webhookURL != nil {
		go ForwardToWebhook(context.Background(), *webhookURL, WebhookPayload{
			EpisodeID:    info.ID,
			EpisodeTitle: info.Title,
			StartTime:    info.StartTime,
			EndTime:      info.EndTime,
			Tracks:       tracks,
		})
	}
	out := &tracklistOutput{}
	out.Body.Episode = info
	out.Body.Tracks = tracks
	return out, nil
}

// --- authenticated handlers . ݁₊ ✶. ݁ ---

func (h *Handler) listTracks(ctx context.Context, input *episodePathInput) (*tracksOnlyOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if _, err := h.channels.Get(ctx, input.ChannelID, stationID); err != nil {
		return nil, huma.Error404NotFound("channel not found")
	}
	tracks, err := h.store.ListTracks(ctx, input.EpisodeID)
	if err != nil {
		slog.Error("failed to list tracks", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &tracksOnlyOutput{}
	out.Body.Tracks = tracks
	return out, nil
}

func (h *Handler) setTracks(ctx context.Context, input *setTracksInput) (*tracksOnlyOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if _, err := h.channels.Get(ctx, input.ChannelID, stationID); err != nil {
		return nil, huma.Error404NotFound("channel not found")
	}
	tracks, err := h.store.SetTracks(ctx, input.EpisodeID, input.Body.Tracks)
	if err != nil {
		slog.Error("failed to save tracks", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	info, webhookURL, err := h.store.EpisodeWithStation(ctx, input.EpisodeID)
	if err == nil && webhookURL != nil {
		go ForwardToWebhook(context.Background(), *webhookURL, WebhookPayload{
			EpisodeID:    info.ID,
			EpisodeTitle: info.Title,
			StartTime:    info.StartTime,
			EndTime:      info.EndTime,
			Tracks:       tracks,
		})
	}
	out := &tracksOnlyOutput{}
	out.Body.Tracks = tracks
	return out, nil
}

func (h *Handler) createSubmissionLink(ctx context.Context, input *episodePathInput) (*submissionLinkOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if _, err := h.channels.Get(ctx, input.ChannelID, stationID); err != nil {
		return nil, huma.Error404NotFound("channel not found")
	}
	token, err := h.store.CreateToken(ctx, input.EpisodeID)
	if err != nil {
		slog.Error("failed to create submission link", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	out := &submissionLinkOutput{}
	out.Body.URL = fmt.Sprintf("%s/tracklist/%s", h.frontURL, token)
	return out, nil
}

// --- webhook receiver ⊹ ˖ ---

func (h *Handler) receiveWebhook(ctx context.Context, input *webhookInput) (*tracksOnlyOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("forbidden")
	}
	if _, err := h.channels.Get(ctx, input.Body.ChannelID, stationID); err != nil {
		return nil, huma.Error404NotFound("channel not found")
	}
	episodeID, err := h.store.FindEpisodeByTime(ctx, input.Body.ChannelID, input.Body.StartTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("no episode found at that time")
		}
		slog.Error("failed to find episode", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	tracks, err := h.store.SetTracks(ctx, episodeID, input.Body.Tracks)
	if err != nil {
		slog.Error("failed to save tracks from webhook", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}
	info, webhookURL, err := h.store.EpisodeWithStation(ctx, episodeID)
	if err == nil && webhookURL != nil {
		go ForwardToWebhook(context.Background(), *webhookURL, WebhookPayload{
			EpisodeID:    info.ID,
			EpisodeTitle: info.Title,
			StartTime:    info.StartTime,
			EndTime:      info.EndTime,
			Tracks:       tracks,
		})
	}
	out := &tracksOnlyOutput{}
	out.Body.Tracks = tracks
	return out, nil
}
