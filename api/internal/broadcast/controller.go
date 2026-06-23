package broadcast

// ✮ ⋆ ˚｡𖦹 broadcast controller — polls schedule, pushes sources to liquidsoap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"radioooooo/internal/episode"
	"radioooooo/internal/media"
	"radioooooo/internal/playlist"
)

const targetLUFS = -18.0

type Controller struct {
	liq       *Client
	episodes  *episode.Store
	media     *media.Store
	playlists *playlist.Store
	channelID string
	queueID   string

	mu      sync.Mutex
	current *episode.Episode
}

func NewController(liq *Client, eps *episode.Store, ms *media.Store, ps *playlist.Store, channelID, queueID string) *Controller {
	return &Controller{
		liq:       liq,
		episodes:  eps,
		media:     ms,
		playlists: ps,
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
			slog.Info("broadcast: schedule gap, falling back", "was", playing.ID)
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
	if ep.Type == episode.TypePlaylist {
		if err := c.pushPlaylist(ctx, ep); err != nil {
			slog.Error("broadcast: playlist push failed", "episode", ep.ID, "error", err)
			return
		}
	} else if ep.Type != episode.TypeLive {
		uri, err := c.resolveSource(ctx, ep)
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

// ⊹ ࣪ ˖ resolves a playlist and pushes all tracks to the queue
func (c *Controller) pushPlaylist(ctx context.Context, ep episode.Episode) error {
	pl, tracks, err := c.playlists.ListTracks(ctx, ep.SourceRef)
	if err != nil {
		return fmt.Errorf("playlist lookup: %w", err)
	}
	if len(tracks) == 0 {
		return fmt.Errorf("playlist %s has no tracks with files", ep.SourceRef)
	}

	if pl.Shuffle {
		rand.Shuffle(len(tracks), func(i, j int) {
			tracks[i], tracks[j] = tracks[j], tracks[i]
		})
	}

	for _, t := range tracks {
		uri := annotateWithGain(t.LocalRef, t.LoudnessLUFS)
		if _, err := c.liq.Push(c.queueID, uri); err != nil {
			return fmt.Errorf("push track: %w", err)
		}
	}

	slog.Info("broadcast: playlist queued", "playlist", ep.SourceRef, "tracks", len(tracks), "shuffle", pl.Shuffle)
	return nil
}

// ✮⋆‧° builds an annotate: uri with liq_amplify if loudness data is available
func annotateWithGain(path string, lufs *float64) string {
	if lufs == nil {
		return path
	}
	gainDB := targetLUFS - *lufs
	gainLin := math.Pow(10, gainDB/20.0)
	return fmt.Sprintf("annotate:liq_amplify=\"%.6f\":%s", gainLin, path)
}

func (c *Controller) resolveSource(ctx context.Context, ep episode.Episode) (string, error) {
	switch ep.SourceAdapter {
	case "local":
		lufs, err := c.media.GetLoudnessByPath(ctx, ep.SourceRef)
		if err != nil {
			slog.Warn("broadcast: loudness lookup failed", "path", ep.SourceRef, "error", err)
			return ep.SourceRef, nil
		}
		return annotateWithGain(ep.SourceRef, lufs), nil
	case "s3":
		// ⋆˙⟡ s3 files are pre-downloaded — resolve via local_ref on the media row
		m, err := c.media.GetBySourceRef(ctx, ep.SourceRef)
		if err != nil {
			return "", fmt.Errorf("s3: media lookup failed for %s: %w", ep.SourceRef, err)
		}
		if m.LocalRef == nil {
			return "", fmt.Errorf("s3: file not downloaded yet for %s", ep.SourceRef)
		}
		return annotateWithGain(*m.LocalRef, m.LoudnessLUFS), nil
	case "external", "http":
		return ep.SourceRef, nil
	default:
		return "", fmt.Errorf("unsupported adapter: %s", ep.SourceAdapter)
	}
}
