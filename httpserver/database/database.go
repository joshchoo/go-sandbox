package database

import (
	"context"
	"database/sql"
	_ "github.com/tursodatabase/go-libsql"
)

func InitSQLiteDB(ctx context.Context, dbFile string) (*sql.DB, error) {
	sqlDB, err := sql.Open("libsql", dbFile)
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	// Enable WAL.
	_, err = sqlDB.QueryContext(ctx, "PRAGMA JOURNAL_MODE=WAL")
	if err != nil {
		return nil, err
	}
	// Perf optimization.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA SYNCHRONOUS=NORMAL")
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}
