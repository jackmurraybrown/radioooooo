package analytics

// ⋆˙⟡ listener stats collector — polls the listener source every 30s,
// resolves countries, and records aggregates. IPs never hit the database.

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	"radioooooo/internal/channel"
)

type CollectorArgs struct{}

func (CollectorArgs) Kind() string { return "listener_stats_collector" }

type CollectorWorker struct {
	river.WorkerDefaults[CollectorArgs]
	source   ListenerSource
	geo      *GeoResolver
	store    *Store
	channels *channel.Store
}

func NewCollectorWorker(source ListenerSource, geo *GeoResolver, store *Store, channels *channel.Store) *CollectorWorker {
	return &CollectorWorker{
		source:   source,
		geo:      geo,
		store:    store,
		channels: channels,
	}
}

func (w *CollectorWorker) Work(ctx context.Context, job *river.Job[CollectorArgs]) error {
	snapshots, err := w.source.Poll(ctx)
	if err != nil {
		slog.Error("analytics: poll failed", "error", err)
		return err
	}

	// ⊹ ࣪ ˖ group by mount + country — IPs are used for geo only, then dropped
	type key struct {
		mount   string
		country string
	}
	counts := make(map[key]int)

	for _, snap := range snapshots {
		country := "XX"
		if w.geo != nil && snap.IP != "" {
			country = w.geo.Country(snap.IP)
		}
		counts[key{mount: snap.Mount, country: country}]++
	}

	// . ݁₊ ✶ resolve mount to channel and current episode, then record
	for k, count := range counts {
		ch, err := w.channels.GetByMount(ctx, k.mount)
		if err != nil {
			slog.Warn("analytics: unknown mount", "mount", k.mount, "error", err)
			continue
		}

		var episodeID *string
		// episode lookup is best-effort — nil if nothing is on air
		if ep, err := w.channels.GetCurrentEpisodeID(ctx, ch.ID); err == nil {
			episodeID = &ep
		}

		if err := w.store.Record(ctx, ch.ID, episodeID, k.country, count); err != nil {
			slog.Error("analytics: record failed", "channel", ch.ID, "error", err)
		}
	}

	if len(snapshots) > 0 {
		slog.Info("analytics: collected", "listeners", len(snapshots), "mounts", len(counts))
	}

	return nil
}
