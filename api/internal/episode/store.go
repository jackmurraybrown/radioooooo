package episode

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


const cols = `
	e.id::text, e.channel_id::text, e.show_id::text, e.title, e.description, e.image_ref,
	e.color, e.start_time, e.end_time, e.type, e.source_adapter, e.source_ref,
	e.original_start, e.ical_uid, e.ical_feed_id::text,
	e.auto_filled, e.repeat_of::text, e.created_at, e.updated_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type CreateParams struct {
	ChannelID     string
	StationID     string // used to verify station owns the channel
	Title         string
	Description   string
	StartTime     time.Time
	EndTime       time.Time
	Type          string
	SourceAdapter string
	SourceRef     string
}

type UpdateParams struct {
	Title         string
	Description   string
	StartTime     time.Time
	EndTime       time.Time
	Type          string
	SourceAdapter string
	SourceRef     string
}

// ⋆˙⟡ returns pgx.ErrNoRows if the channel doesn't belong to stationID.
func (s *Store) Create(ctx context.Context, p CreateParams) (Episode, error) {
	rows, err := s.db.Query(ctx, `
		insert into episodes (channel_id, title, description, start_time, end_time, type, source_adapter, source_ref)
		select $1::uuid, $2, $3, $4, $5, $6, $7, $8
		from channels where id = $1::uuid and station_id = $9::uuid
		returning`+cols,
		p.ChannelID, p.Title, p.Description, p.StartTime, p.EndTime, p.Type, p.SourceAdapter, p.SourceRef, p.StationID,
	)
	if err != nil {
		return Episode{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Episode])
}

func (s *Store) List(ctx context.Context, channelID string) ([]Episode, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from episodes e
		where e.channel_id = $1::uuid
		order by e.start_time asc
	`, channelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Episode])
}

func (s *Store) Get(ctx context.Context, id, channelID string) (Episode, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from episodes e
		where e.id = $1::uuid and e.channel_id = $2::uuid
	`, id, channelID)
	if err != nil {
		return Episode{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Episode])
}

// . ݁₊ ✶ returns pgx.ErrNoRows if not found or wrong station.
func (s *Store) Update(ctx context.Context, id, channelID, stationID string, p UpdateParams) (Episode, error) {
	rows, err := s.db.Query(ctx, `
		update episodes e
		set title=$3, description=$4, start_time=$5, end_time=$6,
		    type=$7, source_adapter=$8, source_ref=$9, updated_at=now()
		from channels c
		where e.id=$1::uuid and e.channel_id=$2::uuid
		  and e.channel_id=c.id and c.station_id=$10::uuid
		returning`+cols,
		id, channelID, p.Title, p.Description, p.StartTime, p.EndTime, p.Type, p.SourceAdapter, p.SourceRef, stationID,
	)
	if err != nil {
		return Episode{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Episode])
}

// ⊹ ࣪ ˖ what's on air right now? pgx.ErrNoRows if nothing.
func (s *Store) GetCurrent(ctx context.Context, channelID string) (Episode, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from episodes e
		where e.channel_id = $1::uuid
		  and e.start_time <= now()
		  and e.end_time > now()
		limit 1
	`, channelID)
	if err != nil {
		return Episode{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Episode])
}

// ✮⋆‧° next episode coming up. pgx.ErrNoRows if nothing.
func (s *Store) GetNext(ctx context.Context, channelID string) (Episode, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from episodes e
		where e.channel_id = $1::uuid
		  and e.start_time > now()
		order by e.start_time asc
		limit 1
	`, channelID)
	if err != nil {
		return Episode{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Episode])
}

func (s *Store) Delete(ctx context.Context, id, channelID, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from episodes e
		using channels c
		where e.id=$1::uuid and e.channel_id=$2::uuid
		  and e.channel_id=c.id and c.station_id=$3::uuid
	`, id, channelID, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
