package project

import (
	"context"
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
	PlanContext *PlanContext
}

type PlanContext struct {
	Files []FileMeta `json:"files"`
}

type Limits struct {
	Concurrency int
	MemMB       int
}

type ctxKey int

const (
	ctxKeyRunContext ctxKey = iota
	ctxKeyInputPath
)

func WithRunContext(ctx context.Context, rc *RunContext) context.Context {
	return context.WithValue(ctx, ctxKeyRunContext, rc)
}

func FromContext(ctx context.Context) *RunContext {
	if v := ctx.Value(ctxKeyRunContext); v != nil {
		if rc, ok := v.(*RunContext); ok {
			return rc
		}
	}
	return nil
}

func WithInputPath(ctx context.Context, in string) context.Context {
	return context.WithValue(ctx, ctxKeyInputPath, in)
}

func InputPathFrom(ctx context.Context) string {
	if v := ctx.Value(ctxKeyInputPath); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
