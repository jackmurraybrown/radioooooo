package playlist

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cols = `
	p.id::text, p.station_id::text, p.name, p.shuffle, p.loop,
	p.source_adapter, p.source_ref, p.created_at, p.updated_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// --- params ✦ ✧ ✦ ---

type CreateParams struct {
	StationID     string
	Name          string
	Shuffle       bool
	Loop          bool
	SourceAdapter *string
	SourceRef     *string
}

type UpdateParams struct {
	Name    string
	Shuffle bool
	Loop    bool
}

// --- store ˚₊✧ ---

func (s *Store) Create(ctx context.Context, p CreateParams) (Playlist, error) {
	rows, err := s.db.Query(ctx, `
		insert into playlists (station_id, name, shuffle, loop, source_adapter, source_ref)
		values ($1::uuid, $2, $3, $4, $5, $6)
		returning`+` id::text, station_id::text, name, shuffle, loop,
		source_adapter, source_ref, created_at, updated_at`,
		p.StationID, p.Name, p.Shuffle, p.Loop, p.SourceAdapter, p.SourceRef,
	)
	if err != nil {
		return Playlist{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Playlist])
}

func (s *Store) List(ctx context.Context, stationID string) ([]Playlist, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from playlists p
		where p.station_id = $1::uuid
		order by p.created_at desc
	`, stationID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Playlist])
}

func (s *Store) Get(ctx context.Context, id, stationID string) (Playlist, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from playlists p
		where p.id = $1::uuid and p.station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return Playlist{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Playlist])
}

// Update returns pgx.ErrNoRows if the playlist is not found or belongs to a different station.
func (s *Store) Update(ctx context.Context, id, stationID string, p UpdateParams) (Playlist, error) {
	rows, err := s.db.Query(ctx, `
		update playlists p
		set name=$3, shuffle=$4, loop=$5, updated_at=now()
		where p.id=$1::uuid and p.station_id=$2::uuid
		returning`+` p.id::text, p.station_id::text, p.name, p.shuffle, p.loop,
		p.source_adapter, p.source_ref, p.created_at, p.updated_at`,
		id, stationID, p.Name, p.Shuffle, p.Loop,
	)
	if err != nil {
		return Playlist{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Playlist])
}

// Delete returns pgx.ErrNoRows if the playlist is not found or belongs to a different station.
func (s *Store) Delete(ctx context.Context, id, stationID string) error {
	result, err := s.db.Exec(ctx, `
		delete from playlists where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ✮⋆‧° resolves a playlist to file paths + loudness for the broadcast controller.
func (s *Store) ListTracks(ctx context.Context, playlistID string) (Playlist, []PlaylistTrack, error) {
	pl, err := s.GetByID(ctx, playlistID)
	if err != nil {
		return Playlist{}, nil, err
	}

	rows, err := s.db.Query(ctx, `
		select m.local_ref, m.loudness_lufs
		from playlist_items pi
		join media m on m.id = pi.media_id
		where pi.playlist_id = $1::uuid
		  and m.local_ref is not null
		order by pi.position asc
	`, playlistID)
	if err != nil {
		return Playlist{}, nil, err
	}
	tracks, err := pgx.CollectRows(rows, pgx.RowToStructByName[PlaylistTrack])
	if err != nil {
		return Playlist{}, nil, err
	}
	return pl, tracks, nil
}

// ⋆˙⟡ get by id without station scoping (for internal use)
func (s *Store) GetByID(ctx context.Context, id string) (Playlist, error) {
	rows, err := s.db.Query(ctx, `
		select`+cols+`
		from playlists p
		where p.id = $1::uuid
	`, id)
	if err != nil {
		return Playlist{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Playlist])
}

// ListItems returns playlist items joined with media metadata, ordered by position.
func (s *Store) ListItems(ctx context.Context, playlistID string) ([]PlaylistItem, error) {
	rows, err := s.db.Query(ctx, `
		select
			pi.id::text, pi.playlist_id::text, pi.media_id::text, pi.position,
			m.title, m.artist, m.duration, pi.created_at
		from playlist_items pi
		join media m on m.id = pi.media_id
		where pi.playlist_id = $1::uuid
		order by pi.position asc
	`, playlistID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[PlaylistItem])
}

// AddItem appends a media item to the playlist at the next available position.
// returns pgx.ErrNoRows if the media does not belong to the station.
func (s *Store) AddItem(ctx context.Context, playlistID, mediaID, stationID string) (PlaylistItem, error) {
	rows, err := s.db.Query(ctx, `
		insert into playlist_items (playlist_id, media_id, position)
		select $1::uuid, $2::uuid,
		       coalesce((select max(position) from playlist_items where playlist_id = $1::uuid), 0) + 1
		from media
		where id = $2::uuid and station_id = $3::uuid
		returning`+` id::text, playlist_id::text, media_id::text, position,
		(select title from media where id = $2::uuid),
		(select artist from media where id = $2::uuid),
		(select duration from media where id = $2::uuid),
		created_at`,
		playlistID, mediaID, stationID,
	)
	if err != nil {
		return PlaylistItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[PlaylistItem])
}

// RemoveItem deletes a playlist item. returns pgx.ErrNoRows if not found.
func (s *Store) RemoveItem(ctx context.Context, itemID, playlistID string) error {
	result, err := s.db.Exec(ctx, `
		delete from playlist_items where id = $1::uuid and playlist_id = $2::uuid
	`, itemID, playlistID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetItems replaces all items in a playlist with an ordered list of media IDs.
// used for drag-and-drop reordering. runs in a transaction.
func (s *Store) SetItems(ctx context.Context, playlistID, stationID string, mediaIDs []string) ([]PlaylistItem, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// verify playlist belongs to station
	var exists bool
	err = tx.QueryRow(ctx, `
		select exists(select 1 from playlists where id = $1::uuid and station_id = $2::uuid)
	`, playlistID, stationID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}

	if _, err = tx.Exec(ctx, `delete from playlist_items where playlist_id = $1::uuid`, playlistID); err != nil {
		return nil, err
	}

	for i, mediaID := range mediaIDs {
		if _, err = tx.Exec(ctx, `
			insert into playlist_items (playlist_id, media_id, position)
			select $1::uuid, $2::uuid, $3
			from media where id = $2::uuid and station_id = $4::uuid
		`, playlistID, mediaID, i+1, stationID); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.ListItems(ctx, playlistID)
}
