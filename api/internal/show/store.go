package show

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	s.id::text, s.channel_id::text, s.title, s.description, s.image_ref,
	s.recurrence_rule, s.duration_minutes, s.type, s.source_adapter, s.source_ref,
	s.allow_repeat, s.contact_email, s.created_at, s.updated_at`

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
	ContactEmail    *string
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
	ContactEmail    *string
}

// ⋆˙⟡ creates a show. returns pgx.ErrNoRows if the channel doesn't belong to stationID.
func (s *Store) Create(ctx context.Context, p CreateParams) (Show, error) {
	rows, err := s.db.Query(ctx, `
		insert into shows (channel_id, title, description, recurrence_rule, duration_minutes, type, source_adapter, source_ref, allow_repeat, contact_email)
		select $1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10
		from channels where id = $1::uuid and station_id = $11::uuid
		returning`+cols,
		p.ChannelID, p.Title, p.Description, p.RecurrenceRule, p.DurationMinutes,
		p.Type, p.SourceAdapter, p.SourceRef, p.AllowRepeat, p.ContactEmail, p.StationID,
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
		    type=$7, source_adapter=$8, source_ref=$9, allow_repeat=$10, contact_email=$11, updated_at=now()
		from channels c
		where s.id=$1::uuid and s.channel_id=$2::uuid
		  and s.channel_id=c.id and c.station_id=$12::uuid
		returning`+cols,
		id, channelID, p.Title, p.Description, p.RecurrenceRule, p.DurationMinutes,
		p.Type, p.SourceAdapter, p.SourceRef, p.AllowRepeat, p.ContactEmail, stationID,
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

// ✮⋆‧° returns all shows (for the expansion job)
func (s *Store) ListAll(ctx context.Context) ([]Show, error) {
	rows, err := s.db.Query(ctx, `select`+cols+` from shows s order by s.created_at asc`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Show])
}

// ⋆˙⟡ latest original_start for a show's episodes — tells us where expansion left off
func (s *Store) LatestEpisodeStart(ctx context.Context, showID string) (*time.Time, error) {
	var t *time.Time
	err := s.db.QueryRow(ctx, `
		select max(original_start) from episodes where show_id = $1::uuid
	`, showID).Scan(&t)
	return t, err
}

// ⊹ ࣪ ˖ inserts a batch of expanded episodes for a show
func (s *Store) InsertEpisodes(ctx context.Context, showID, channelID string, sh Show, starts []time.Time) (int, error) {
	count := 0
	for _, start := range starts {
		end := start.Add(time.Duration(sh.DurationMinutes) * time.Minute)
		_, err := s.db.Exec(ctx, `
			insert into episodes (channel_id, show_id, title, description, image_ref, start_time, end_time, type, source_adapter, source_ref, original_start, contact_email)
			values ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $6, $11)
		`, channelID, showID, sh.Title, sh.Description, sh.ImageRef, start, end, sh.Type, sh.SourceAdapter, sh.SourceRef, sh.ContactEmail)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
