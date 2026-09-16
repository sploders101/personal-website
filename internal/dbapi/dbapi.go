package dbapi

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	queries "github.com/sploders101/personal-website/internal/dbapi/gen"
)

type Db struct {
	db *sql.DB
}

func NewDb(ctx context.Context, driver string, url string) (Db, error) {
	db, err := sql.Open(driver, url)
	if err != nil {
		return Db{}, err
	}
	if err := MigrateDb(ctx, db); err != nil {
		return Db{}, err
	}
	return Db{db}, nil
}

func (db Db) Begin(ctx context.Context) (Tx, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return Tx{}, err
	}
	return Tx{tx}, nil
}

func (db Db) Query() *queries.Queries {
	return queries.New(db.db)
}

func (db Db) Close() {
	if err := db.db.Close(); err != nil {
		slog.Error("Unable to close database pool", "error", err)
	}
}

type Tx struct {
	Tx *sql.Tx
}

func (tx Tx) Commit() error {
	return tx.Tx.Commit()
}

// Small helper for rolling back on an unbounded context without linter errors.
// Unbounded context is used here because the rollback is often intentionally
// triggered by a context cancellation, which if used, would cause the rollback
// to not actually finish.
//
// Errors are simply logged because they are rarely actionable
func (tx Tx) Rollback() {
	//nolint:noctx
	if err := tx.Tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		slog.Error("Failed to roll back transaction", "error", err)
	}
}

func (db Tx) Query() *queries.Queries {
	return queries.New(db.Tx)
}
