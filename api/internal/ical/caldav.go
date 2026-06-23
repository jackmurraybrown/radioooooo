package ical

// ⋆˙⟡ CalDAV sync — bidirectional calendar sync via WebDAV
// works with Nextcloud, Fastmail, Apple Calendar, Radicale
// https://datatracker.ietf.org/doc/html/rfc4791

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-webdav/caldav"
)

func syncCalDAV(ctx context.Context, store *Store, feed Feed) error {
	client, err := caldav.NewClient(
		nil,
		feed.URL,
	)
	if err != nil {
		return fmt.Errorf("caldav client: %w", err)
	}

	// ⊹ ࣪ ˖ if no calendar path specified, find the default one
	calPath := ""
	if feed.CalendarPath != nil && *feed.CalendarPath != "" {
		calPath = *feed.CalendarPath
	} else {
		principal, err := client.FindCurrentUserPrincipal(ctx)
		if err != nil {
			return fmt.Errorf("caldav principal: %w", err)
		}
		homeSet, err := client.FindCalendarHomeSet(ctx, principal)
		if err != nil {
			return fmt.Errorf("caldav home set: %w", err)
		}
		calendars, err := client.FindCalendars(ctx, homeSet)
		if err != nil {
			return fmt.Errorf("caldav find calendars: %w", err)
		}
		if len(calendars) == 0 {
			return fmt.Errorf("caldav: no calendars found")
		}
		calPath = calendars[0].Path
	}

	// . ݁₊ ✶ query events from now to 90 days out
	now := time.Now()
	horizon := now.AddDate(0, 0, 90)

	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name: "VEVENT",
				Props: []caldav.PropFilter{},
				Start: now,
				End:   horizon,
			}},
		},
	}

	objects, err := client.QueryCalendar(ctx, calPath, query)
	if err != nil {
		return fmt.Errorf("caldav query: %w", err)
	}

	var activeUIDs []string

	for _, obj := range objects {
		if obj.Data == nil || len(obj.Data.Events()) == 0 {
			continue
		}

		for _, event := range obj.Data.Events() {
			uid := event.Props.Get("UID")
			if uid == nil || uid.Value == "" {
				continue
			}

			start, err := event.DateTimeStart(nil)
			if err != nil {
				slog.Warn("caldav: skip event, no start", "uid", uid.Value, "error", err)
				continue
			}
			end, err := event.DateTimeEnd(nil)
			if err != nil {
				slog.Warn("caldav: skip event, no end", "uid", uid.Value, "error", err)
				continue
			}

			title := ""
			if summary := event.Props.Get("SUMMARY"); summary != nil {
				title = summary.Value
			}

			description := ""
			if desc := event.Props.Get("DESCRIPTION"); desc != nil {
				description = desc.Value
			}

			var color *string
			for _, prop := range []string{"COLOR", "X-APPLE-CALENDAR-COLOR"} {
				if p := event.Props.Get(prop); p != nil && p.Value != "" {
					v := strings.TrimSpace(p.Value)
					color = &v
					break
				}
			}

			if err := store.UpsertEpisodeByUID(ctx, feed.ID, feed.ChannelID, uid.Value, title, description, color, start, end); err != nil {
				slog.Error("caldav: upsert failed", "uid", uid.Value, "error", err)
				continue
			}

			activeUIDs = append(activeUIDs, uid.Value)
		}
	}

	deleted, err := store.DeleteStaleEpisodes(ctx, feed.ID, activeUIDs)
	if err != nil {
		slog.Error("caldav: stale cleanup failed", "feed", feed.ID, "error", err)
	} else if deleted > 0 {
		slog.Info("caldav: removed stale episodes", "feed", feed.ID, "count", deleted)
	}

	if err := store.MarkSynced(ctx, feed.ID); err != nil {
		slog.Error("caldav: mark synced failed", "feed", feed.ID, "error", err)
	}

	slog.Info("caldav: synced", "feed", feed.ID, "events", len(activeUIDs))
	return nil
}
