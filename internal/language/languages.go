package language

import (
	"context"
	"database/sql"

	"github.com/ChaseHampton/cargoworker/internal/db"
)

func GetSourceExtension(ext string, rdb *sql.DB, ctx context.Context) (*SourceExtension, error) {
	rows, err := rdb.QueryContext(ctx, db.LanguageByExtensionSQL, ext)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lang SourceExtension
	if rows.Next() {
		if err := rows.Scan(&lang.Extension, &lang.LanguageID, &lang.IsText, &lang.IsPrimary, &lang.Notes); err != nil {
			return nil, err
		}
		return &lang, nil
	}
	return nil, sql.ErrNoRows
}

func GetSourceBasename(name string, rdb *sql.DB, ctx context.Context) (*SourceBasename, error) {
	rows, err := rdb.QueryContext(ctx, db.LanguageByBasenameSQL, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lang SourceBasename
	if rows.Next() {
		if err := rows.Scan(&lang.Name, &lang.LanguageID, &lang.IsText, &lang.Notes); err != nil {
			return nil, err
		}
		return &lang, nil
	}
	return nil, sql.ErrNoRows
}
