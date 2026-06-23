package database

import (
	"context"
	"embed"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// opens a connection pool and pings.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// ⊹ ࣪ ˖ river schema. managed by rivermigrate
func MigrateRiver(ctx context.Context, db *pgxpool.Pool) error {
	m, err := rivermigrate.New(riverpgxv5.New(db), nil)
	if err != nil {
		return err
	}
	_, err = m.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}

// ⋆˙⟡ app schema. managed by golang-migrate
func Migrate(dsn string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	migrateDSN := strings.Replace(dsn, "postgres://", "pgx5://", 1)

	m, err := migrate.NewWithSourceInstance("iofs", source, migrateDSN)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
