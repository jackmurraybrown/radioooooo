package gapfill

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// returns channels with gap fill enabled
func (s *Store) ListEnabledChannels(ctx context.Context) ([]EnabledChannel, error) {
	rows, err := s.db.Query(ctx, `
		select c.id::text, c.station_id::text, c.repeat_prefill_days,
		       st.timezone
		from channels c
		join stations st on st.id = c.station_id
		where c.gap_fill_enabled = true
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[EnabledChannel])
}

type EnabledChannel struct {
	ID               string `db:"id"`
	StationID        string `db:"station_id"`
	RepeatPrefillDays int   `db:"repeat_prefill_days"`
	Timezone         string `db:"timezone"`
}

func (s *Store) ListRules(ctx context.Context, channelID string) ([]Rule, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, channel_id::text, priority, time_from, time_to,
		       type, source_ref, created_at
		from gap_fill_rules
		where channel_id = $1::uuid
		order by priority asc
	`, channelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Rule])
}

// finds gaps in the schedule between now and the horizon
func (s *Store) FindGaps(ctx context.Context, channelID string, horizon time.Time) ([]Gap, error) {
	rows, err := s.db.Query(ctx, `
		select start_time, end_time from episodes
		where channel_id = $1::uuid
		  and end_time > now()
		  and start_time < $2
		  and auto_filled = false
		order by start_time asc
	`, channelID, horizon)
	if err != nil {
		return nil, err
	}

	type ep struct {
		StartTime time.Time `db:"start_time"`
		EndTime   time.Time `db:"end_time"`
	}
	episodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[ep])
	if err != nil {
		return nil, err
	}

	var gaps []Gap
	cursor := time.Now()

	for _, e := range episodes {
		if e.StartTime.After(cursor) {
			gaps = append(gaps, Gap{Start: cursor, End: e.StartTime})
		}
		if e.EndTime.After(cursor) {
			cursor = e.EndTime
		}
	}

	if cursor.Before(horizon) {
		gaps = append(gaps, Gap{Start: cursor, End: horizon})
	}

	return gaps, nil
}

func (s *Store) ClearAutoFilled(ctx context.Context, channelID string) (int, error) {
	result, err := s.db.Exec(ctx, `
		delete from episodes
		where channel_id = $1::uuid and auto_filled = true and start_time > now()
	`, channelID)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// inserts an auto-filled episode. repeatOf is the original episode ID for show-repeats.
func (s *Store) InsertAutoFilled(ctx context.Context, channelID, title, sourceAdapter, sourceRef, epType string, start, end time.Time, repeatOf *string) error {
	_, err := s.db.Exec(ctx, `
		insert into episodes (channel_id, title, type, source_adapter, source_ref, start_time, end_time, auto_filled, repeat_of)
		values ($1::uuid, $2, $3, $4, $5, $6, $7, true, $8)
	`, channelID, title, epType, sourceAdapter, sourceRef, start, end, repeatOf)
	return err
}

// ⊹ ࣪ ˖ finds a past episode from a show eligible for repeat
func (s *Store) FindRepeatEpisode(ctx context.Context, showID, channelID string, avoidDays *int) (*RepeatCandidate, error) {
	excludeAfter := time.Now().AddDate(0, 0, -365)
	if avoidDays != nil && *avoidDays > 0 {
		excludeAfter = time.Now().AddDate(0, 0, -*avoidDays)
	}

	rows, err := s.db.Query(ctx, `
		select e.id::text, e.title, e.type, e.source_adapter, e.source_ref,
		       extract(epoch from (e.end_time - e.start_time))::int as duration_seconds
		from episodes e
		join shows s on s.id = e.show_id
		where e.show_id = $1::uuid
		  and s.allow_repeat = true
		  and e.end_time < now()
		  and e.type in ('recorded', 'playlist')
		  and e.id not in (
		      select ra.episode_id from repeat_airings ra
		      where ra.channel_id = $2::uuid and ra.aired_at > $3
		  )
		order by e.end_time desc
		limit 1
	`, showID, channelID, excludeAfter)
	if err != nil {
		return nil, err
	}
	candidates, err := pgx.CollectRows(rows, pgx.RowToStructByName[RepeatCandidate])
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	return &candidates[0], nil
}

type RepeatCandidate struct {
	ID              string `db:"id"`
	Title           string `db:"title"`
	Type            string `db:"type"`
	SourceAdapter   string `db:"source_adapter"`
	SourceRef       string `db:"source_ref"`
	DurationSeconds int    `db:"duration_seconds"`
}
