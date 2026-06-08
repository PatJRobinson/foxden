package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", path, err)
	}

	return &DB{sql: db}, nil
}

func (db *DB) Init(ctx context.Context) error {
	if _, err := db.sql.ExecContext(ctx, initialSchema); err != nil {
		return fmt.Errorf("initialize schema: %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}

	return db.sql.Close()
}
