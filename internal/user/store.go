package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

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
