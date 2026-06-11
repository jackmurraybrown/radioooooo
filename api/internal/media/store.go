package media

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	m.id::text, m.station_id::text, m.title, m.artist, m.duration, m.artwork_ref,
	m.file_format, m.file_size_bytes,
	m.source_adapter, m.source_ref, m.download_status, m.download_error,
	m.downloaded_at, m.created_at, m.updated_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// --- params ✦ ✧ ✦ ---

type CreateParams struct {
	StationID     string
	Title         string
	Artist        *string
	ArtworkRef    *string
	FileFormat    *string
	FileSizeBytes *int64
	SourceAdapter string
	SourceRef     string
}

type UpdateParams struct {
	Title      string
	Artist     *string
	ArtworkRef *string
	FileFormat *string
}

// --- store ˚₊✧ ---

func (s *Store) Create(ctx context.Context, p CreateParams) (Media, error) {
	rows, err := s.db.Query(ctx, `
		insert into media (
			station_id, title, artist, artwork_ref,
			file_format, file_size_bytes, source_adapter, source_ref
		) values (
			$1::uuid, $2, $3, $4, $5, $6, $7, $8
		)
		returning`+` id::text, station_id::text, title, artist, duration, artwork_ref,
		file_format, file_size_bytes,
		source_adapter, source_ref, download_status, download_error,
		downloaded_at, created_at, updated_at`,
		p.StationID, p.Title, p.Artist, p.ArtworkRef,
		p.FileFormat, p.FileSizeBytes, p.SourceAdapter, p.SourceRef,
	)
	if err != nil {
		return Media{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Media])
}

func (s *Store) List(ctx context.Context, stationID string) ([]Media, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from media m
		where m.station_id = $1::uuid
		order by m.created_at desc
	`, stationID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Media])
}

func (s *Store) Get(ctx context.Context, id, stationID string) (Media, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from media m
		where m.id = $1::uuid and m.station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return Media{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Media])
}

// Update returns pgx.ErrNoRows if the item is not found or belongs to a different station.
func (s *Store) Update(ctx context.Context, id, stationID string, p UpdateParams) (Media, error) {
	rows, err := s.db.Query(ctx, `
		update media m
		set title=$3, artist=$4, artwork_ref=$5, file_format=$6, updated_at=now()
		where m.id=$1::uuid and m.station_id=$2::uuid
		returning`+` m.id::text, m.station_id::text, m.title, m.artist, m.duration, m.artwork_ref,
		m.file_format, m.file_size_bytes,
		m.source_adapter, m.source_ref, m.download_status, m.download_error,
		m.downloaded_at, m.created_at, m.updated_at`,
		id, stationID, p.Title, p.Artist, p.ArtworkRef, p.FileFormat,
	)
	if err != nil {
		return Media{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Media])
}

// Delete returns pgx.ErrNoRows if the item is not found or belongs to a different station.
func (s *Store) Delete(ctx context.Context, id, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from media where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
