package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrationFs embed.FS

func MigrateDb(ctx context.Context, db *sql.DB) error {
	migrations, err := fs.Sub(migrationFs, "migrations/")
	if err != nil {
		return err
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("failed to create postgres session locker")
	}
	provider, err := goose.NewProvider(
		database.DialectPostgres,
		db,
		migrations,
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return err
	}
	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	return nil
}
