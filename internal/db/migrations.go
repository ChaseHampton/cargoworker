package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/migrations/*.sql
var migFS embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	curr, err := CurrentUserVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("get current user version: %w", err)
	}

	entries, err := migFS.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var list []migration
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		base := e.Name() // e.g., 0002_fts.sql
		vstr := strings.SplitN(base, "_", 2)[0]
		v, err := strconv.Atoi(strings.TrimSuffix(vstr, ".sql"))
		if err != nil {
			continue
		}
		if v <= curr {
			continue
		}

		b, err := migFS.ReadFile(path.Join("sql/migrations", base))
		if err != nil {
			return err
		}
		list = append(list, migration{version: v, name: base, sql: string(b)})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })

	for _, m := range list {
		if err := applyOne(ctx, db, m); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
	}
	return nil
}

func applyOne(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Enforce FKs during migration execution.
	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys=ON;`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version=%d;`, m.version)); err != nil {
		return err
	}
	return tx.Commit()
}
