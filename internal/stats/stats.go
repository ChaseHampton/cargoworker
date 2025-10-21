package stats

import (
	"strings"
	"time"
)

type Stats struct {
	Run  RunStats      `json:"run"`
	Plan *PlanSnapshot `json:"plan,omitempty"`
}

type RunStats struct {
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	DurationMS  int64     `json:"duration_ms"`
	RunID       string    `json:"run_id"`
	InputPath   string    `json:"input_path"`
	OutDir      string    `json:"out_dir"`
	ToolVersion string    `json:"tool_version"`
	IRSchema    string    `json:"ir_schema"`
	Warnings    int64     `json:"warnings"`
	Errors      int64     `json:"errors"`
}

type Option func(*RunStats)

func WithStartedAt(t time.Time) Option {
	return func(rs *RunStats) {
		if !t.IsZero() {
			rs.StartedAt = t
		}
	}
}

func WithRunID(id string) Option {
	return func(rs *RunStats) { rs.RunID = strings.TrimSpace(id) }
}

func WithInputPath(p string) Option {
	return func(rs *RunStats) { rs.InputPath = strings.TrimSpace(p) }
}

func WithOutDir(p string) Option { return func(rs *RunStats) { rs.OutDir = strings.TrimSpace(p) } }

func WithToolVersion(v string) Option {
	return func(rs *RunStats) { rs.ToolVersion = strings.TrimSpace(v) }
}

func WithIRSchema(v string) Option { return func(rs *RunStats) { rs.IRSchema = strings.TrimSpace(v) } }

func WithWarnErr(warn, err int64) Option {
	return func(rs *RunStats) {
		if warn > 0 {
			rs.Warnings = warn
		}
		if err > 0 {
			rs.Errors = err
		}
	}
}

// New constructs a Stats with sensible defaults and applies any provided options.
// Defaults:
// - StartedAt: time.Now().UTC()
// - Other fields: zero values until set by options or later updates.
func New(opts ...Option) *Stats {
	rs := RunStats{StartedAt: time.Now().UTC()}
	for _, opt := range opts {
		if opt != nil {
			opt(&rs)
		}
	}
	return &Stats{Run: rs}
}

func (s *Stats) End() {
	if s == nil {
		return
	}
	s.Run.EndedAt = time.Now().UTC()
	if !s.Run.StartedAt.IsZero() {
		s.Run.DurationMS = s.Run.EndedAt.Sub(s.Run.StartedAt).Milliseconds()
	}
}

func (s *Stats) IncWarnings(n int64) {
	if s == nil {
		return
	}
	if n <= 0 {
		n = 1
	}
	s.Run.Warnings += n
}

func (s *Stats) IncErrors(n int64) {
	if s == nil {
		return
	}
	if n <= 0 {
		n = 1
	}
	s.Run.Errors += n
}

func (s *Stats) SetPlanSnapshot(p *Plan) {
	if s == nil || p == nil {
		return
	}
	s.Plan = p.Snapshot()
}

func (s *Stats) SetPlan(ps *PlanSnapshot) {
	if s == nil {
		return
	}
	s.Plan = ps
}
