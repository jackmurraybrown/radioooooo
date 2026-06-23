package gapfill

// gap-fill job — scans for schedule gaps and fills them
// re-derives from scratch each run

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
)

type FillArgs struct{}

func (FillArgs) Kind() string { return "gap_fill" }

type FillWorker struct {
	river.WorkerDefaults[FillArgs]
	store *Store
}

func NewFillWorker(store *Store) *FillWorker {
	return &FillWorker{store: store}
}

func (w *FillWorker) Work(ctx context.Context, job *river.Job[FillArgs]) error {
	channels, err := w.store.ListEnabledChannels(ctx)
	if err != nil {
		return fmt.Errorf("gapfill: list channels: %w", err)
	}

	for _, ch := range channels {
		if err := w.fillChannel(ctx, ch); err != nil {
			slog.Error("gapfill: channel failed", "channel", ch.ID, "error", err)
		}
	}

	return nil
}

func (w *FillWorker) fillChannel(ctx context.Context, ch EnabledChannel) error {
	loc, err := time.LoadLocation(ch.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone %s: %w", ch.Timezone, err)
	}

	// . ݁₊ ✶ clear and rebuild
	cleared, err := w.store.ClearAutoFilled(ctx, ch.ID)
	if err != nil {
		return fmt.Errorf("clear: %w", err)
	}
	if cleared > 0 {
		slog.Info("gapfill: cleared stale auto-fills", "channel", ch.ID, "count", cleared)
	}

	horizon := time.Now().AddDate(0, 0, ch.RepeatPrefillDays)

	gaps, err := w.store.FindGaps(ctx, ch.ID, horizon)
	if err != nil {
		return fmt.Errorf("find gaps: %w", err)
	}
	if len(gaps) == 0 {
		return nil
	}

	rules, err := w.store.ListRules(ctx, ch.ID)
	if err != nil {
		return fmt.Errorf("list rules: %w", err)
	}
	if len(rules) == 0 {
		return nil
	}

	filled := 0
	for _, gap := range gaps {
		rule := matchRule(rules, gap, loc)
		if rule == nil {
			continue
		}

		switch rule.Type {
		case "playlist":
			err := w.store.InsertAutoFilled(ctx, ch.ID, "auto: playlist", "playlist", rule.SourceRef, "playlist", gap.Start, gap.End, nil)
			if err != nil {
				slog.Warn("gapfill: playlist insert failed", "error", err)
				continue
			}
			filled++

		case "show-repeat":
			candidate, err := w.store.FindRepeatEpisode(ctx, rule.SourceRef, ch.ID, nil)
			if err != nil || candidate == nil {
				continue
			}
			end := gap.Start.Add(time.Duration(candidate.DurationSeconds) * time.Second)
			if end.After(gap.End) {
				end = gap.End
			}
			err = w.store.InsertAutoFilled(ctx, ch.ID, candidate.Title+" (repeat)", candidate.SourceAdapter, candidate.SourceRef, candidate.Type, gap.Start, end, &candidate.ID)
			if err != nil {
				slog.Warn("gapfill: repeat insert failed", "error", err)
				continue
			}
			filled++
		}
	}

	if filled > 0 {
		slog.Info("gapfill: episodes created", "channel", ch.ID, "count", filled)
	}

	return nil
}

// ✮⋆‧° finds the first rule whose time slot covers the gap start
func matchRule(rules []Rule, gap Gap, loc *time.Location) *Rule {
	gapTime := gap.Start.In(loc).Format("15:04")

	for i := range rules {
		r := &rules[i]

		// no time constraints — matches everything
		if r.TimeFrom == nil && r.TimeTo == nil {
			return r
		}

		from := "00:00"
		to := "24:00"
		if r.TimeFrom != nil {
			from = *r.TimeFrom
		}
		if r.TimeTo != nil {
			to = *r.TimeTo
		}

		if from <= to {
			if gapTime >= from && gapTime < to {
				return r
			}
		} else {
			// ⋆˙⟡ wraps past midnight (e.g. 22:00-06:00)
			if gapTime >= from || gapTime < to {
				return r
			}
		}
	}

	return nil
}
