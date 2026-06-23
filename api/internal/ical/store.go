package ical

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	id::text, station_id::text, channel_id::text, type, url,
	username, password, calendar_path,
	last_synced_at, created_at, updated_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type CreateParams struct {
	StationID    string
	ChannelID    string
	Type         string
	URL          string
	Username     *string
	Password     *string
	CalendarPath *string
}

func (s *Store) Create(ctx context.Context, p CreateParams) (Feed, error) {
	rows, err := s.db.Query(ctx, `
		insert into ical_feeds (station_id, channel_id, type, url, username, password, calendar_path)
		select $1::uuid, $2::uuid, $3, $4, $5, $6, $7
		from channels where id = $2::uuid and station_id = $1::uuid
		returning`+cols,
		p.StationID, p.ChannelID, p.Type, p.URL, p.Username, p.Password, p.CalendarPath,
	)
	if err != nil {
		return Feed{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Feed])
}

func (s *Store) List(ctx context.Context, channelID string) ([]Feed, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+` from ical_feeds where channel_id = $1::uuid order by created_at asc
	`, channelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Feed])
}

func (s *Store) Delete(ctx context.Context, id, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from ical_feeds where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ✮⋆‧° returns all feeds (for the sync job)
func (s *Store) ListAll(ctx context.Context) ([]Feed, error) {
	rows, err := s.db.Query(ctx, `select`+cols+` from ical_feeds order by created_at asc`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Feed])
}

func (s *Store) MarkSynced(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `
		update ical_feeds set last_synced_at = now(), updated_at = now() where id = $1::uuid
	`, id)
	return err
}

// ⋆˙⟡ upserts an episode by ical UID — creates or updates based on the calendar event
func (s *Store) UpsertEpisodeByUID(ctx context.Context, feedID, channelID, uid, title, description string, color *string, start, end interface{}) error {
	_, err := s.db.Exec(ctx, `
		insert into episodes (channel_id, ical_feed_id, ical_uid, title, description, color, start_time, end_time, type, source_adapter, source_ref)
		values ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, 'live', 'icecast', 'dj-input')
		on conflict (ical_feed_id, ical_uid) where ical_uid is not null
		do update set
			title = excluded.title,
			description = excluded.description,
			color = excluded.color,
			start_time = excluded.start_time,
			end_time = excluded.end_time,
			updated_at = now()
	`, channelID, feedID, uid, title, description, color, start, end)
	return err
}

// ⊹ ࣪ ˖ removes episodes from a feed that are no longer in the calendar
func (s *Store) DeleteStaleEpisodes(ctx context.Context, feedID string, activeUIDs []string) (int, error) {
	if len(activeUIDs) == 0 {
		result, err := s.db.Exec(ctx, `
			delete from episodes where ical_feed_id = $1::uuid
		`, feedID)
		if err != nil {
			return 0, err
		}
		return int(result.RowsAffected()), nil
	}
	result, err := s.db.Exec(ctx, `
		delete from episodes where ical_feed_id = $1::uuid and ical_uid != all($2)
	`, feedID, activeUIDs)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}
