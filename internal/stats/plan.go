package stats

import (
	"sync"
	"sync/atomic"
	"time"
)

type Plan struct {
	// timings
	startedAt time.Time
	endedAt   time.Time
	dbInitMS  int64 // set directly (no need for atomic)

	// decision
	language       string // "go"
	languageSource string // "flag" | "auto" | "config"

	// inputs
	root           string
	includeGlobs   []string
	excludeGlobs   []string
	gitIgnoreFound bool

	// counters (hot paths use atomics)
	filesDiscovered  atomic.Int64
	filesSelected    atomic.Int64
	dirsVisited      atomic.Int64
	largestFileBytes atomic.Int64

	// small fields (guarded with mu when needed)
	maxDepth int
	warnCnt  atomic.Int64
	errCnt   atomic.Int64

	// breakdown map needs a mutex
	mu              sync.Mutex
	ignoredByReason map[string]int64
}

func NewPlan(root string, include, exclude []string) *Plan {
	return &Plan{
		startedAt:       time.Now().UTC(),
		root:            root,
		includeGlobs:    append([]string(nil), include...),
		excludeGlobs:    append([]string(nil), exclude...),
		gitIgnoreFound:  false,
		ignoredByReason: make(map[string]int64, 8),
	}
}

func (p *Plan) End() {
	p.endedAt = time.Now().UTC()
}

func (p *Plan) AddDBInit(ms int64)              { p.dbInitMS = ms }
func (p *Plan) SetLanguage(lang string)         { p.language = lang }
func (p *Plan) SetGitIgnoreFound(found bool)    { p.gitIgnoreFound = found }
func (p *Plan) SetLanguageSource(source string) { p.languageSource = source }
func (p *Plan) IncDiscovered(n int64)           { p.filesDiscovered.Add(n) }
func (p *Plan) IncSelected(n int64)             { p.filesSelected.Add(n) }
func (p *Plan) IncDirs(n int64)                 { p.dirsVisited.Add(n) }
func (p *Plan) MaxDepthSeen(depth int) {
	if depth <= 0 {
		return
	}
	p.mu.Lock()
	if depth > p.maxDepth {
		p.maxDepth = depth
	}
	p.mu.Unlock()
}
func (p *Plan) ConsiderLargest(sizeBytes int64) {
	for {
		cur := p.largestFileBytes.Load()
		if sizeBytes <= cur {
			return
		}
		if p.largestFileBytes.CompareAndSwap(cur, sizeBytes) {
			return
		}
	}
}
func (p *Plan) Ignore(reason string, n int64) {
	if n == 0 {
		return
	}
	p.mu.Lock()
	p.ignoredByReason[reason] += n
	p.mu.Unlock()
}
func (p *Plan) Warn()  { p.warnCnt.Add(1) }
func (p *Plan) Error() { p.errCnt.Add(1) }

func (p *Plan) Snapshot() *PlanSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	ignoredCopy := make(map[string]int64, len(p.ignoredByReason))
	for k, v := range p.ignoredByReason {
		ignoredCopy[k] = v
	}

	start := p.startedAt
	end := p.endedAt
	var dur int64
	if !start.IsZero() && !end.IsZero() {
		dur = end.Sub(start).Milliseconds()
	}

	return &PlanSnapshot{
		StartedAt:        start,
		EndedAt:          end,
		DurationMS:       dur,
		DBInitMS:         p.dbInitMS,
		Language:         p.language,
		LanguageSource:   p.languageSource,
		Root:             p.root,
		IncludeGlobs:     append([]string(nil), p.includeGlobs...),
		ExcludeGlobs:     append([]string(nil), p.excludeGlobs...),
		GitIgnoreFound:   p.gitIgnoreFound,
		FilesDiscovered:  p.filesDiscovered.Load(),
		FilesSelected:    p.filesSelected.Load(),
		FilesIgnored:     p.filesDiscovered.Load() - p.filesSelected.Load(),
		IgnoredByReason:  ignoredCopy,
		DirsVisited:      p.dirsVisited.Load(),
		MaxDepth:         p.maxDepth,
		LargestFileBytes: p.largestFileBytes.Load(),
		WarnCount:        p.warnCnt.Load(),
		ErrorCount:       p.errCnt.Load(),
	}
}

type PlanSnapshot struct {
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMS int64     `json:"duration_ms"`
	DBInitMS   int64     `json:"db_init_ms"`

	Language       string `json:"language"`
	LanguageSource string `json:"language_source"`

	Root           string   `json:"root"`
	IncludeGlobs   []string `json:"include_globs"`
	ExcludeGlobs   []string `json:"exclude_globs"`
	GitIgnoreFound bool     `json:"gitignore_found"`

	FilesDiscovered int64            `json:"files_discovered"`
	FilesSelected   int64            `json:"files_selected"`
	FilesIgnored    int64            `json:"files_ignored"`
	IgnoredByReason map[string]int64 `json:"ignored_by_reason"`

	DirsVisited      int64 `json:"dirs_visited"`
	MaxDepth         int   `json:"max_depth"`
	LargestFileBytes int64 `json:"largest_file_bytes"`

	WarnCount  int64 `json:"warn_count"`
	ErrorCount int64 `json:"error_count"`
}
