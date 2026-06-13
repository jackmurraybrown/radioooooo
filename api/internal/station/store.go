package station

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, name, slug string) (Station, error) {
	rows, err := s.db.Query(ctx, `
		INSERT INTO stations (name, slug)
		VALUES ($1, $2)
		RETURNING id::text, name, slug, logo_url, created_at, updated_at
	`, name, slug)
	if err != nil {
		return Station{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Station])
}

func (s *Store) List(ctx context.Context) ([]Station, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, name, slug, logo_url, created_at, updated_at
		FROM stations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Station])
}

func (s *Store) Get(ctx context.Context, id string) (Station, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, name, slug, logo_url, created_at, updated_at
		FROM stations
		WHERE id = $1::uuid
	`, id)
	if err != nil {
		return Station{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Station])
}

type UpdateParams struct {
	Name    string
	Slug    string
	LogoURL *string
}

func (s *Store) Update(ctx context.Context, id string, p UpdateParams) (Station, error) {
	rows, err := s.db.Query(ctx, `
		UPDATE stations
		SET name = $1, slug = $2, logo_url = $3, updated_at = now()
		WHERE id = $4::uuid
		RETURNING id::text, name, slug, logo_url, created_at, updated_at
	`, p.Name, p.Slug, p.LogoURL, id)
	if err != nil {
		return Station{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Station])
}

func (s *Store) Delete(ctx context.Context, id string) error {
	result, err := s.db.Exec(ctx, `DELETE FROM stations WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreateAPIKey generates a new api key for the station, stores its hash, and returns
// the plain-text key. the key is only available at creation time — it cannot be recovered.
func (s *Store) CreateAPIKey(ctx context.Context, stationID string) (string, error) {
	plain, hashed, err := generateKey()
	if err != nil {
		return "", fmt.Errorf("generating api key: %w", err)
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO api_keys (station_id, key_hash)
		VALUES ($1::uuid, $2)
	`, stationID, hashed)
	return plain, err
}

// VerifyAPIKey checks a plain-text key against stored hashes and returns the
// associated station id. returns pgx.ErrNoRows if the key is not found.
func (s *Store) VerifyAPIKey(ctx context.Context, plain string) (string, error) {
	hashed := hashKey(plain)
	var stationID string
	err := s.db.QueryRow(ctx, `
		SELECT station_id::text FROM api_keys WHERE key_hash = $1
	`, hashed).Scan(&stationID)
	return stationID, err
}

func generateKey() (plain, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	hashed = hashKey(plain)
	return
}

func hashKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
