package show

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	s.id::text, s.channel_id::text, s.title, s.description, s.image_ref,
	s.recurrence_rule, s.duration_minutes, s.type, s.source_adapter, s.source_ref,
	s.allow_repeat, s.created_at, s.updated_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type CreateParams struct {
	ChannelID       string
	StationID       string
	Title           string
	Description     string
	RecurrenceRule  string
	DurationMinutes int
	Type            string
	SourceAdapter   string
	SourceRef       string
	AllowRepeat     bool
}

type UpdateParams struct {
	Title           string
	Description     string
	RecurrenceRule  string
	DurationMinutes int
	Type            string
	SourceAdapter   string
	SourceRef       string
	AllowRepeat     bool
}

// ⋆˙⟡ creates a show. returns pgx.ErrNoRows if the channel doesn't belong to stationID.
func (s *Store) Create(ctx context.Context, p CreateParams) (Show, error) {
	rows, err := s.db.Query(ctx, `
		insert into shows (channel_id, title, description, recurrence_rule, duration_minutes, type, source_adapter, source_ref, allow_repeat)
		select $1::uuid, $2, $3, $4, $5, $6, $7, $8, $9
		from channels where id = $1::uuid and station_id = $10::uuid
		returning`+cols,
		p.ChannelID, p.Title, p.Description, p.RecurrenceRule, p.DurationMinutes,
		p.Type, p.SourceAdapter, p.SourceRef, p.AllowRepeat, p.StationID,
	)
	if err != nil {
		return Show{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Show])
}

func (s *Store) List(ctx context.Context, channelID string) ([]Show, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from shows s
		where s.channel_id = $1::uuid
		order by s.created_at asc
	`, channelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Show])
}

func (s *Store) Get(ctx context.Context, id, channelID string) (Show, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from shows s
		where s.id = $1::uuid and s.channel_id = $2::uuid
	`, id, channelID)
	if err != nil {
		return Show{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Show])
}

// ⊹ ࣪ ˖ returns pgx.ErrNoRows if not found or wrong station.
func (s *Store) Update(ctx context.Context, id, channelID, stationID string, p UpdateParams) (Show, error) {
	rows, err := s.db.Query(ctx, `
		update shows s
		set title=$3, description=$4, recurrence_rule=$5, duration_minutes=$6,
		    type=$7, source_adapter=$8, source_ref=$9, allow_repeat=$10, updated_at=now()
		from channels c
		where s.id=$1::uuid and s.channel_id=$2::uuid
		  and s.channel_id=c.id and c.station_id=$11::uuid
		returning`+` s.id::text, s.channel_id::text, s.title, s.description, s.image_ref,
		s.recurrence_rule, s.duration_minutes, s.type, s.source_adapter, s.source_ref,
		s.allow_repeat, s.created_at, s.updated_at`,
		id, channelID, p.Title, p.Description, p.RecurrenceRule, p.DurationMinutes,
		p.Type, p.SourceAdapter, p.SourceRef, p.AllowRepeat, stationID,
	)
	if err != nil {
		return Show{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Show])
}

func (s *Store) Delete(ctx context.Context, id, channelID, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from shows s
		using channels c
		where s.id=$1::uuid and s.channel_id=$2::uuid
		  and s.channel_id=c.id and c.station_id=$3::uuid
	`, id, channelID, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
