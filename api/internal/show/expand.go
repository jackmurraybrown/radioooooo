package show

// ⋆˙⟡ rrule expansion — generates episode rows from a show's recurrence rule

import (
	"context"
	"fmt"
	"time"

	"github.com/teambition/rrule-go"
)

const DefaultHorizonDays = 90

// ExpandShow generates episodes for a show up to the horizon.
// called on show creation and by the daily expansion job.
func ExpandShow(ctx context.Context, store *Store, s Show, loc *time.Location, horizonDays int) (int, error) {
	rule, err := rrule.StrToRRule(s.RecurrenceRule)
	if err != nil {
		return 0, fmt.Errorf("parse rrule: %w", err)
	}
	rule.DTStart(rule.GetDTStart().In(loc))

	if horizonDays <= 0 {
		horizonDays = DefaultHorizonDays
	}
	horizon := time.Now().AddDate(0, 0, horizonDays)

	if !rule.GetUntil().IsZero() && rule.GetUntil().Before(horizon) {
		horizon = rule.GetUntil()
	}

	// ⊹ ࣪ ˖ find where we left off
	latest, err := store.LatestEpisodeStart(ctx, s.ID)
	if err != nil {
		return 0, fmt.Errorf("latest episode: %w", err)
	}

	expandFrom := time.Now()
	if latest != nil && latest.After(expandFrom) {
		expandFrom = *latest
	}

	// . ݁₊ ✶ generate occurrences
	occurrences := rule.Between(expandFrom, horizon, false)
	if len(occurrences) == 0 {
		return 0, nil
	}

	return store.InsertEpisodes(ctx, s.ID, s.ChannelID, s, occurrences)
}
