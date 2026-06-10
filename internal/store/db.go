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

	if err := db.addColumnIfMissing(ctx, "stories", "content_source", "TEXT DEFAULT 'feed'"); err != nil {
		return err
	}

	if err := db.addColumnIfMissing(ctx, "stories", "article_fetched_at", "TEXT"); err != nil {
		return err
	}

	if err := db.addColumnIfMissing(ctx, "stories", "article_fetch_error", "TEXT"); err != nil {
		return err
	}

	return nil
}

func (db *DB) addColumnIfMissing(ctx context.Context, table string, column string, definition string) error {
	rows, err := db.sql.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return fmt.Errorf("inspect table %q: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue any
		var pk int

		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan table info for %q: %w", table, err)
		}

		if name == column {
			return nil
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate table info for %q: %w", table, err)
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	if _, err := db.sql.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}

	return nil
}

func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}

	return db.sql.Close()
}
