package tracklist

// ✮⋆‧°—°‧⋆✮ tracklist store — token management + CRUD for episode tracks

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const trackCols = `id::text, episode_id::text, position, title, artist, album, started_at, ended_at, created_at`

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// ⊹ ࣪ ˖ generates a 30-day token for the tracklist edit link
func (s *Store) CreateToken(ctx context.Context, episodeID string) (string, error) {
	plain, hashed, err := generateToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, `
		insert into tracklist_tokens (episode_id, token_hash, expires_at)
		values ($1::uuid, $2, $3)
	`, episodeID, hashed, time.Now().Add(30*24*time.Hour))
	return plain, err
}

// ⋆˙⟡ returns episode ID if token is valid and unexpired (multi-use)
func (s *Store) ValidateToken(ctx context.Context, plain string) (string, error) {
	hashed := hashToken(plain)
	var episodeID string
	err := s.db.QueryRow(ctx, `
		select episode_id::text from tracklist_tokens
		where token_hash = $1 and expires_at > now()
	`, hashed).Scan(&episodeID)
	return episodeID, err
}

func (s *Store) ListTracks(ctx context.Context, episodeID string) ([]Track, error) {
	rows, err := s.db.Query(ctx, `
		select `+trackCols+` from episode_tracks
		where episode_id = $1::uuid
		order by position asc
	`, episodeID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Track])
}

// . ݁₊ ✶ replaces all tracks for an episode in one tx
func (s *Store) SetTracks(ctx context.Context, episodeID string, inputs []TrackInput) ([]Track, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from episode_tracks where episode_id = $1::uuid`, episodeID); err != nil {
		return nil, err
	}

	for i, t := range inputs {
		_, err := tx.Exec(ctx, `
			insert into episode_tracks (episode_id, position, title, artist, album, started_at, ended_at)
			values ($1::uuid, $2, $3, $4, $5, $6, $7)
		`, episodeID, i, t.Title, t.Artist, t.Album, t.StartedAt, t.EndedAt)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.ListTracks(ctx, episodeID)
}

// ˖ˎˊ˗ episode info + station webhook url for forwarding decisions
func (s *Store) EpisodeWithStation(ctx context.Context, episodeID string) (EpisodeInfo, *string, error) {
	var info EpisodeInfo
	var webhookURL *string
	err := s.db.QueryRow(ctx, `
		select e.id::text, e.title, e.start_time, e.end_time,
		       st.id::text, st.tracklist_webhook_url
		from episodes e
		join channels c on c.id = e.channel_id
		join stations st on st.id = c.station_id
		where e.id = $1::uuid
	`, episodeID).Scan(&info.ID, &info.Title, &info.StartTime, &info.EndTime, &info.StationID, &webhookURL)
	return info, webhookURL, err
}

// ٩(^ᗜ^ )و match episode by channel + time overlap (for wails webhook)
func (s *Store) FindEpisodeByTime(ctx context.Context, channelID string, startTime time.Time) (string, error) {
	var episodeID string
	err := s.db.QueryRow(ctx, `
		select id::text from episodes
		where channel_id = $1::uuid
		  and start_time <= $2 and end_time > $2
	`, channelID, startTime).Scan(&episodeID)
	return episodeID, err
}

func generateToken() (plain, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	hashed = hashToken(plain)
	return
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
