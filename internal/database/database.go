package database

import (
	"context"
	"embed"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// connect opens a connection pool and verifies connectivity.
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

// migrate runs any pending migrations. safe to call on every startup --
// it is a no-op when the schema is already up to date.
func Migrate(dsn string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// golang-migrate's pgx v5 driver expects pgx5:// not postgres://
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
