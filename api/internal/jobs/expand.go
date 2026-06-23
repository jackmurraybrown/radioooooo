package jobs

// . ݁₊ ✶ daily show expansion job — expands all shows up to the horizon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"radioooooo/internal/show"
	"radioooooo/internal/station"
)

type ShowExpansionArgs struct{}

func (ShowExpansionArgs) Kind() string { return "show_expansion" }

type ShowExpansionWorker struct {
	river.WorkerDefaults[ShowExpansionArgs]
	store    *show.Store
	stations *station.Store
}

func NewShowExpansionWorker(store *show.Store, stations *station.Store) *ShowExpansionWorker {
	return &ShowExpansionWorker{store: store, stations: stations}
}

func (w *ShowExpansionWorker) Work(ctx context.Context, job *river.Job[ShowExpansionArgs]) error {
	shows, err := w.store.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("expand: list shows: %w", err)
	}

	for _, s := range shows {
		tzName, err := w.stations.TimezoneForChannel(ctx, s.ChannelID)
		if err != nil {
			slog.Error("expand: timezone lookup failed", "show", s.ID, "error", err)
			continue
		}
		loc, err := time.LoadLocation(tzName)
		if err != nil {
			slog.Error("expand: invalid timezone", "show", s.ID, "tz", tzName, "error", err)
			continue
		}

		count, err := show.ExpandShow(ctx, w.store, s, loc)
		if err != nil {
			slog.Error("expand: show failed", "show", s.ID, "title", s.Title, "error", err)
			continue
		}
		if count > 0 {
			slog.Info("expand: episodes created", "show", s.ID, "title", s.Title, "count", count)
		}
	}

	return nil
}
