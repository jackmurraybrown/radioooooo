package broadcast

// . ݁₊ ✶ broadcast manager — spawns one controller per channel
// replaces the single hardcoded BROADCAST_CHANNEL_ID

import (
	"context"
	"log/slog"
	"sync"

	"radioooooo/internal/channel"
	"radioooooo/internal/episode"
	"radioooooo/internal/media"
	"radioooooo/internal/playlist"
)

type Manager struct {
	liq       *Client
	channels  *channel.Store
	episodes  *episode.Store
	media     *media.Store
	playlists *playlist.Store

	mu          sync.Mutex
	controllers map[string]*Controller
	cancels     map[string]context.CancelFunc
}

func NewManager(liq *Client, channels *channel.Store, episodes *episode.Store, media *media.Store, playlists *playlist.Store) *Manager {
	return &Manager{
		liq:         liq,
		channels:    channels,
		episodes:    episodes,
		media:       media,
		playlists:   playlists,
		controllers: make(map[string]*Controller),
		cancels:     make(map[string]context.CancelFunc),
	}
}

// ⋆˙⟡ starts a controller for every channel in the database
func (m *Manager) Start(ctx context.Context) error {
	chs, err := m.channels.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, ch := range chs {
		m.startChannel(ctx, ch)
	}

	slog.Info("broadcast: manager started", "channels", len(chs))
	return nil
}

func (m *Manager) startChannel(ctx context.Context, ch channel.Channel) {
	// ⊹ ࣪ ˖ queue ID derived from slug — matches the liquidsoap source name
	queueID := "queue_" + ch.Slug
	ctrl := NewController(m.liq, m.episodes, m.media, m.playlists, ch.ID, queueID)

	childCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.controllers[ch.ID] = ctrl
	m.cancels[ch.ID] = cancel
	m.mu.Unlock()

	go func() {
		slog.Info("broadcast: channel started", "channel", ch.ID, "slug", ch.Slug, "mount", ch.Mount)
		if err := ctrl.Run(childCtx); err != nil && err != context.Canceled {
			slog.Error("broadcast: channel exited", "channel", ch.ID, "error", err)
		}
	}()
}

// ✮ ⋆ ˚｡𖦹 stops all controllers
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, cancel := range m.cancels {
		cancel()
		delete(m.cancels, id)
		delete(m.controllers, id)
	}

	slog.Info("broadcast: manager stopped")
}
