package media

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	m.id::text, m.station_id::text, m.title, m.artist, m.duration, m.artwork_ref,
	m.file_format, m.file_size_bytes,
	m.source_adapter, m.source_ref, m.local_ref, m.download_status, m.download_error,
	m.downloaded_at, m.loudness_lufs, m.true_peak_db, m.created_at, m.updated_at`

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
		source_adapter, source_ref, local_ref, download_status, download_error,
		downloaded_at, loudness_lufs, true_peak_db, created_at, updated_at`,
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
		m.source_adapter, m.source_ref, m.local_ref, m.download_status, m.download_error,
		m.downloaded_at, m.loudness_lufs, m.true_peak_db, m.created_at, m.updated_at`,
		id, stationID, p.Title, p.Artist, p.ArtworkRef, p.FileFormat,
	)
	if err != nil {
		return Media{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Media])
}

// UpdateStatus sets the download_status (and optionally duration) for a media item.
// ✮ ⋆ ˚｡𖦹 used by the ingest pipeline and seed tooling.
func (s *Store) UpdateStatus(ctx context.Context, id, stationID, status string, duration *int) error {
	_, err := s.db.Exec(ctx, `
		update media set download_status=$3, duration=coalesce($4, duration), updated_at=now()
		where id=$1::uuid and station_id=$2::uuid
	`, id, stationID, status, duration)
	return err
}

// sets local_ref and marks as ready after upload
func (s *Store) SetLocalRef(ctx context.Context, id, localRef string) error {
	_, err := s.db.Exec(ctx, `
		update media set local_ref=$2, download_status='ready', updated_at=now()
		where id=$1::uuid
	`, id, localRef)
	return err
}

// ⋆˙⟡ sets loudness + duration after ffmpeg analysis.
func (s *Store) UpdateLoudness(ctx context.Context, id string, lufs, truePeak float64, duration *int) error {
	_, err := s.db.Exec(ctx, `
		update media set loudness_lufs=$2, true_peak_db=$3, duration=coalesce($4, duration), updated_at=now()
		where id=$1::uuid
	`, id, lufs, truePeak, duration)
	return err
}

// ⊹ ࣪ ˖ finds media ready for loudness analysis.
func (s *Store) ListUnanalysed(ctx context.Context, limit int) ([]Media, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from media m
		where m.download_status = 'ready' and m.loudness_lufs is null
		order by m.created_at asc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Media])
}

// ✮⋆‧° looks up a media item by source_ref (s3 key, url, etc.)
func (s *Store) GetBySourceRef(ctx context.Context, ref string) (Media, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from media m
		where m.source_ref = $1
		limit 1
	`, ref)
	if err != nil {
		return Media{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Media])
}

// (˵>ᗜ<˵) looks up loudness by file path for the broadcast controller.
// returns nil, nil if no matching media or no analysis yet.
func (s *Store) GetLoudnessByPath(ctx context.Context, path string) (*float64, error) {
	var lufs *float64
	err := s.db.QueryRow(ctx, `
		select loudness_lufs from media
		where local_ref = $1 and loudness_lufs is not null
		limit 1
	`, path).Scan(&lufs)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return lufs, nil
}

// delete returns pgx.ErrNoRows if the item is not found or belongs to a different station.
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
