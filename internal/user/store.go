package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenDuration = 30 * 24 * time.Hour

const cols = `id::text, station_id::text, email, created_at, updated_at`

// --- store ˚₊✧ ---

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Create inserts a new user with a bcrypt-hashed password.
func (s *Store) Create(ctx context.Context, stationID, email, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	rows, err := s.db.Query(ctx, `
		insert into users (station_id, email, password_hash)
		values ($1::uuid, $2, $3)
		returning `+cols,
		stationID, email, string(hash),
	)
	if err != nil {
		return User{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
}

// GetByEmail returns the user and their stored password hash for login verification.
// returns pgx.ErrNoRows if no user exists with that email.
func (s *Store) GetByEmail(ctx context.Context, email string) (User, string, error) {
	var hash string
	var u User
	rows, err := s.db.Query(ctx, `
		select `+cols+`, password_hash
		from users
		where email = $1
	`, email)
	if err != nil {
		return User{}, "", err
	}

	type row struct {
		User
		PasswordHash string `db:"password_hash"`
	}
	r, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[row])
	if err != nil {
		return User{}, "", err
	}
	u = r.User
	hash = r.PasswordHash
	return u, hash, nil
}

// CheckPassword returns true if the plain-text password matches the stored hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CreateRefreshToken generates a random token, stores its hash, and returns the plain token.
func (s *Store) CreateRefreshToken(ctx context.Context, userID string) (string, error) {
	plain, hash, err := generateToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, `
		insert into refresh_tokens (user_id, token_hash, expires_at)
		values ($1::uuid, $2, $3)
	`, userID, hash, time.Now().Add(refreshTokenDuration))
	return plain, err
}

// RotateRefreshToken validates the old token, deletes it, issues a new one, and returns the user.
// returns pgx.ErrNoRows if the token is invalid or expired.
func (s *Store) RotateRefreshToken(ctx context.Context, plain string) (User, string, error) {
	hash := hashToken(plain)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return User{}, "", err
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		delete from refresh_tokens
		where token_hash = $1 and expires_at > now()
		returning user_id::text
	`, hash).Scan(&userID)
	if err != nil {
		return User{}, "", err
	}

	rows, err := tx.Query(ctx, `select `+cols+` from users where id = $1::uuid`, userID)
	if err != nil {
		return User{}, "", err
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return User{}, "", err
	}

	newPlain, newHash, err := generateToken()
	if err != nil {
		return User{}, "", err
	}
	_, err = tx.Exec(ctx, `
		insert into refresh_tokens (user_id, token_hash, expires_at)
		values ($1::uuid, $2, $3)
	`, userID, newHash, time.Now().Add(refreshTokenDuration))
	if err != nil {
		return User{}, "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return User{}, "", err
	}
	return u, newPlain, nil
}

// DeleteRefreshToken removes a refresh token by plain value — used on logout.
func (s *Store) DeleteRefreshToken(ctx context.Context, plain string) error {
	result, err := s.db.Exec(ctx, `
		delete from refresh_tokens where token_hash = $1
	`, hashToken(plain))
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func generateToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = fmt.Sprintf("%x", b)
	hash = hashToken(plain)
	return
}

func hashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return fmt.Sprintf("%x", h)
}
