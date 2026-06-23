package channel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, stationID, name, slug string) (Channel, error) {
	rows, err := s.db.Query(ctx, `
		insert into channels (station_id, name, slug)
		values ($1::uuid, $2, $3)
		returning id::text, station_id::text, name, slug, mount, created_at, updated_at
	`, stationID, name, slug)
	if err != nil {
		return Channel{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Channel])
}

func (s *Store) List(ctx context.Context, stationID string) ([]Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, created_at, updated_at
		from channels
		where station_id = $1::uuid
		order by created_at asc
	`, stationID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Channel])
}

func (s *Store) Get(ctx context.Context, id, stationID string) (Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, created_at, updated_at
		from channels
		where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return Channel{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Channel])
}

// ⋆˙⟡ looks up a channel by its icecast mount path
func (s *Store) GetByMount(ctx context.Context, mount string) (Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, created_at, updated_at
		from channels where mount = $1
	`, mount)
	if err != nil {
		return Channel{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Channel])
}

// ⊹ ࣪ ˖ returns the current episode ID for a channel, or pgx.ErrNoRows
func (s *Store) GetCurrentEpisodeID(ctx context.Context, channelID string) (string, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		select id::text from episodes
		where channel_id = $1::uuid and start_time <= now() and end_time > now()
		limit 1
	`, channelID).Scan(&id)
	return id, err
}

func (s *Store) Delete(ctx context.Context, id, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from channels where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
