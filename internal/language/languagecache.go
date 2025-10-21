package language

import (
	"context"
	"database/sql"
	"sync"

	"golang.org/x/sync/singleflight"
)

type Language struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Ecosystem string `json:"ecosystem" db:"ecosystem"`
}

type SourceExtension struct {
	Extension  string  `json:"extension" db:"ext"`
	LanguageID string  `json:"language_id" db:"language_id"`
	IsText     bool    `json:"is_text" db:"is_text"`
	IsPrimary  bool    `json:"is_primary" db:"is_primary"`
	Notes      *string `json:"notes,omitempty" db:"notes"`
}

type SourceBasename struct {
	Name       string  `json:"name" db:"name"`
	LanguageID string  `json:"language_id" db:"language_id"`
	IsText     bool    `json:"is_text" db:"is_text"`
	Notes      *string `json:"notes,omitempty" db:"notes"`
}

type LanguageCache struct {
	DB        *sql.DB
	Languages map[string]Language
	LangMu    sync.RWMutex

	SourceExtensions map[string]SourceExtension
	ExtMu            sync.RWMutex

	SourceBasenames map[string]SourceBasename
	BaseMu          sync.RWMutex

	sf singleflight.Group
}

func NewLanguageCache(rdb *sql.DB) *LanguageCache {
	return &LanguageCache{
		DB:               rdb,
		Languages:        make(map[string]Language),
		SourceExtensions: make(map[string]SourceExtension),
		SourceBasenames:  make(map[string]SourceBasename),
	}
}

func (lc *LanguageCache) GetExtensionInfo(ext string, ctx context.Context) (*SourceExtension, error) {
	lc.ExtMu.RLock()
	if langExt, ok := lc.SourceExtensions[ext]; ok {
		lc.ExtMu.RUnlock()
		return &langExt, nil
	}
	lc.ExtMu.RUnlock()

	lookupval, err := GetSourceExtension(ext, lc.DB, ctx)
	if err != nil {
		return nil, err
	}

	lc.ExtMu.Lock()
	lc.SourceExtensions[ext] = *lookupval
	lc.ExtMu.Unlock()

	return lookupval, nil
}

func (lc *LanguageCache) GetBasenameInfo(name string, ctx context.Context) (*SourceBasename, error) {
	lc.BaseMu.RLock()
	if langBase, ok := lc.SourceBasenames[name]; ok {
		lc.BaseMu.RUnlock()
		return &langBase, nil
	}
	lc.BaseMu.RUnlock()
	lookupval, err := GetSourceBasename(name, lc.DB, ctx)
	if err != nil {
		return nil, err
	}

	lc.BaseMu.Lock()
	lc.SourceBasenames[name] = *lookupval
	lc.BaseMu.Unlock()

	return lookupval, nil
}
