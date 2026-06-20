package broadcast

// ✮ ⋆ ˚｡𖦹 broadcast controller — polls schedule, pushes sources to liquidsoap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"radioooooo/internal/episode"
)

type Controller struct {
	liq       *Client
	episodes  *episode.Store
	channelID string
	queueID   string

	mu      sync.Mutex
	current *episode.Episode
}

func NewController(liq *Client, eps *episode.Store, channelID, queueID string) *Controller {
	return &Controller{
		liq:       liq,
		episodes:  eps,
		channelID: channelID,
		queueID:   queueID,
	}
}

// ⊹ ࣪ ˖ polls every second, wall-clock aligned, until ctx is cancelled.
func (c *Controller) Run(ctx context.Context) error {
	c.tick(ctx) // ⋆˙⟡ immediate on startup

	for {
		now := time.Now()
		next := now.Truncate(time.Second).Add(time.Second)

		select {
		case <-time.After(next.Sub(now)):
			c.tick(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Controller) tick(ctx context.Context) {
	ep, err := c.episodes.GetCurrent(ctx, c.channelID)

	if errors.Is(err, pgx.ErrNoRows) {
		c.mu.Lock()
		playing := c.current
		c.current = nil
		c.mu.Unlock()

		if playing != nil {
			c.skip()
			slog.Info("broadcast: schedule gap, falling back to silence", "was", playing.ID)
		}
		return
	}

	if err != nil {
		slog.Error("broadcast: schedule query failed", "error", err)
		return
	}

	c.mu.Lock()
	same := c.current != nil && c.current.ID == ep.ID
	c.mu.Unlock()

	if same {
		return
	}

	// . ݁₊ ✶ push new source before skipping old — this avoids a gap in the stream
	if ep.Type != episode.TypeLive {
		uri, err := resolveSource(ep)
		if err != nil {
			slog.Error("broadcast: resolve source failed", "episode", ep.ID, "error", err)
			return
		}
		if _, err := c.liq.Push(c.queueID, uri); err != nil {
			slog.Error("broadcast: push failed", "episode", ep.ID, "error", err)
			return
		}
	}

	c.mu.Lock()
	hadCurrent := c.current != nil
	c.current = &ep
	c.mu.Unlock()

	if hadCurrent {
		c.skip()
	}

	slog.Info("broadcast: episode started", "episode", ep.ID, "type", ep.Type, "ends", ep.EndTime)
}

func (c *Controller) skip() {
	if _, err := c.liq.Command(fmt.Sprintf("%s.skip", c.queueID)); err != nil {
		slog.Error("broadcast: skip failed", "error", err)
	}
}

func resolveSource(ep episode.Episode) (string, error) {
	switch ep.SourceAdapter {
	case "local":
		return ep.SourceRef, nil
	case "external", "http":
		return ep.SourceRef, nil
	default:
		return "", fmt.Errorf("unsupported adapter: %s", ep.SourceAdapter)
	}
}
