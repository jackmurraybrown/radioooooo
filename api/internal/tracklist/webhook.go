package tracklist

// ⋆˙⟡ ⋆.˚ forward tracklist to station webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type WebhookPayload struct {
	EpisodeID    string    `json:"episodeId"`
	EpisodeTitle string    `json:"episodeTitle"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Tracks       []Track   `json:"tracks"`
}

// ⊹ ˖ fire-and-forget POST to station webhook url
func ForwardToWebhook(ctx context.Context, webhookURL string, payload WebhookPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("tracklist webhook marshal failed", "error", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("tracklist webhook request failed", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("tracklist webhook delivery failed", "url", webhookURL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("tracklist webhook returned error", "url", webhookURL, "status", resp.StatusCode)
	}
}
