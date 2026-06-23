package ical

// ✮ ⋆ ˚｡𖦹 ical feed sync — polls external calendars, syncs events to episodes
// https://datatracker.ietf.org/doc/html/rfc5545

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/riverqueue/river"
)

type SyncArgs struct{}

func (SyncArgs) Kind() string { return "ical_feed_sync" }

type SyncWorker struct {
	river.WorkerDefaults[SyncArgs]
	store  *Store
	client *http.Client
}

func NewSyncWorker(store *Store) *SyncWorker {
	return &SyncWorker{
		store:  store,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *SyncWorker) Work(ctx context.Context, job *river.Job[SyncArgs]) error {
	feeds, err := w.store.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("ical: list feeds: %w", err)
	}

	for _, feed := range feeds {
		var err error
		switch feed.Type {
		case "caldav":
			err = syncCalDAV(ctx, w.store, feed)
		default:
			err = w.syncIcalFeed(ctx, feed)
		}
		if err != nil {
			slog.Error("ical: sync failed", "feed", feed.ID, "type", feed.Type, "url", feed.URL, "error", err)
		}
	}

	return nil
}

func (w *SyncWorker) syncIcalFeed(ctx context.Context, feed Feed) error {
	req, err := http.NewRequestWithContext(ctx, "GET", feed.URL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	cal, err := ics.ParseCalendar(resp.Body)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	var activeUIDs []string

	for _, event := range cal.Events() {
		uid := event.GetProperty(ics.ComponentPropertyUniqueId)
		if uid == nil || uid.Value == "" {
			continue
		}

		start, err := event.GetStartAt()
		if err != nil {
			slog.Warn("ical: skip event, no start", "uid", uid.Value, "error", err)
			continue
		}
		end, err := event.GetEndAt()
		if err != nil {
			slog.Warn("ical: skip event, no end", "uid", uid.Value, "error", err)
			continue
		}

		title := ""
		if summary := event.GetProperty(ics.ComponentPropertySummary); summary != nil {
			title = summary.Value
		}

		description := ""
		if desc := event.GetProperty(ics.ComponentPropertyDescription); desc != nil {
			description = desc.Value
		}

		color := extractColor(event)

		if err := w.store.UpsertEpisodeByUID(ctx, feed.ID, feed.ChannelID, uid.Value, title, description, color, start, end); err != nil {
			slog.Error("ical: upsert failed", "uid", uid.Value, "error", err)
			continue
		}

		activeUIDs = append(activeUIDs, uid.Value)
	}

	// . ݁₊ ✶ remove episodes that were deleted from the calendar
	deleted, err := w.store.DeleteStaleEpisodes(ctx, feed.ID, activeUIDs)
	if err != nil {
		slog.Error("ical: stale cleanup failed", "feed", feed.ID, "error", err)
	} else if deleted > 0 {
		slog.Info("ical: removed stale episodes", "feed", feed.ID, "count", deleted)
	}

	if err := w.store.MarkSynced(ctx, feed.ID); err != nil {
		slog.Error("ical: mark synced failed", "feed", feed.ID, "error", err)
	}

	slog.Info("ical: synced", "feed", feed.ID, "events", len(activeUIDs))
	return nil
}

// ⋆˙⟡ extracts colour from iCal event — tries RFC 7986 COLOR, then Apple, then Google
func extractColor(event *ics.VEvent) *string {
	for _, prop := range []string{"COLOR", "X-APPLE-CALENDAR-COLOR"} {
		if p := event.GetProperty(ics.ComponentProperty(prop)); p != nil && p.Value != "" {
			v := strings.TrimSpace(p.Value)
			return &v
		}
	}
	return nil
}
