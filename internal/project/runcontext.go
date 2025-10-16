package project

import (
	"database/sql"
	"log/slog"

	"github.com/ChaseHampton/cargoworker/internal/stats"
	"github.com/google/uuid"
)

type RunContext struct {
	RunId       uuid.UUID
	OutDir      string
	ToolVersion string
	IRSchema    string
	Limits      Limits
	Logger      *slog.Logger
	DB          *sql.DB
	Events      chan<- Event
	Stats       *stats.Stats
	Closers     []func() error
}

type Limits struct {
	Concurrency int
	MemMB       int
}
