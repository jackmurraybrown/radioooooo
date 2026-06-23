package channel

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db            *pgxpool.Pool
	encryptionKey []byte // 32 bytes for AES-256
}

func NewStore(db *pgxpool.Pool, encryptionKey string) *Store {
	key, _ := hex.DecodeString(encryptionKey)
	return &Store{db: db, encryptionKey: key}
}

func (s *Store) Create(ctx context.Context, stationID, name, slug string) (Channel, error) {
	rows, err := s.db.Query(ctx, `
		insert into channels (station_id, name, slug)
		values ($1::uuid, $2, $3)
		returning id::text, station_id::text, name, slug, mount, harbor_password_hash, created_at, updated_at
	`, stationID, name, slug)
	if err != nil {
		return Channel{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Channel])
}

func (s *Store) List(ctx context.Context, stationID string) ([]Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, harbor_password_hash, created_at, updated_at
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
		select id::text, station_id::text, name, slug, mount, harbor_password_hash, created_at, updated_at
		from channels
		where id = $1::uuid and station_id = $2::uuid
	`, id, stationID)
	if err != nil {
		return Channel{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Channel])
}

// ✮⋆‧° all channels across all stations (for the broadcast manager)
func (s *Store) ListAll(ctx context.Context) ([]Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, harbor_password_hash, created_at, updated_at
		from channels order by created_at asc
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Channel])
}

// ⋆˙⟡ looks up a channel by its icecast mount path
func (s *Store) GetByMount(ctx context.Context, mount string) (Channel, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, name, slug, mount, harbor_password_hash, created_at, updated_at
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

// ✮⋆‧° generates a new harbor password, stores AES-encrypted, returns plaintext.
func (s *Store) SetHarborPassword(ctx context.Context, id, stationID string) (string, error) {
	plain, err := generateHarborPassword()
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	encrypted, err := s.encrypt(plain)
	if err != nil {
		return "", fmt.Errorf("encrypt password: %w", err)
	}
	result, err := s.db.Exec(ctx, `
		update channels set harbor_password_hash = $3, updated_at = now()
		where id = $1::uuid and station_id = $2::uuid
	`, id, stationID, encrypted)
	if err != nil {
		return "", err
	}
	if result.RowsAffected() == 0 {
		return "", pgx.ErrNoRows
	}
	return plain, nil
}

// ⋆˙⟡ decrypts and returns the harbor password for display in the dashboard.
func (s *Store) GetHarborPassword(ctx context.Context, id, stationID string) (string, error) {
	var encrypted *string
	err := s.db.QueryRow(ctx, `
		select harbor_password_hash from channels
		where id = $1::uuid and station_id = $2::uuid
	`, id, stationID).Scan(&encrypted)
	if err != nil {
		return "", err
	}
	if encrypted == nil {
		return "", nil
	}
	return s.decrypt(*encrypted)
}

// ⊹ ࣪ ˖ validates a harbor connection — liquidsoap auth callback.
func (s *Store) VerifyHarborPassword(ctx context.Context, mount, password string) (bool, error) {
	var encrypted *string
	err := s.db.QueryRow(ctx, `
		select harbor_password_hash from channels where mount = $1
	`, mount).Scan(&encrypted)
	if err != nil || encrypted == nil {
		return false, err
	}
	plain, err := s.decrypt(*encrypted)
	if err != nil {
		return false, err
	}
	return plain == password, nil
}

func generateHarborPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// . ݁₊ ✶ AES-256-GCM encrypt/decrypt
func (s *Store) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *Store) decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
