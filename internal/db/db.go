package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Open(ctx context.Context, filePath string) (*sql.DB, error) {
	conn_str := fmt.Sprintf("file:%s?_busy_timeout=5000&_pragma=foreign_keys(1)", filePath)
	db, err := sql.Open("sqlite", conn_str)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA temp_store=MEMORY;",
		"PRAGMA busy_timeout=5000;",
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}

	ver, err := CurrentUserVersion(ctx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read user_version: %w", err)
	}
	if ver == 0 {
		if err := BootstrapPersistentPragmas(ctx, db); err != nil {
			db.Close()
			return nil, fmt.Errorf("bootstrap pragmas: %w", err)
		}
	}

	return db, nil
}

func CurrentUserVersion(ctx context.Context, db *sql.DB) (int, error) {
	var v int
	if err := db.QueryRowContext(ctx, `PRAGMA user_version;`).Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

func BootstrapPersistentPragmas(ctx context.Context, db *sql.DB) error {
	persistent := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA page_size=4096;",
		"PRAGMA auto_vacuum=INCREMENTAL;",
		"VACUUM;",
	}
	for _, p := range persistent {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return fmt.Errorf("persistent pragma %q: %w", p, err)
		}
	}
	return nil
}
