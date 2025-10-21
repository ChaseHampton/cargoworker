package plan

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChaseHampton/cargoworker/internal/db"
	"github.com/ChaseHampton/cargoworker/internal/project"
	"github.com/ChaseHampton/cargoworker/internal/stats"
	"github.com/sabhiram/go-gitignore"
)

type Runner struct {
	RunPlan *stats.Plan
}

func NewRunner(plan *stats.Plan) *Runner {
	return &Runner{
		RunPlan: plan,
	}
}

func (r *Runner) Plan(ctx context.Context) (*stats.PlanSnapshot, error) {
	rc := project.FromContext(ctx)
	if rc == nil {
		return &stats.PlanSnapshot{}, fmt.Errorf("internal: planRunner: run context unavailable")
	}
	in := project.InputPathFrom(ctx)
	if r.RunPlan == nil {
		tPlan := stats.NewPlan(in, []string{}, []string{})
		r.RunPlan = tPlan
	}

	_ = buildDatabase(ctx, rc)

	ig, err := ignore.CompileIgnoreFileAndLines(filepath.Join(in, ".gitignore"), r.RunPlan.Snapshot().ExcludeGlobs...)
	if err != nil {
		rc.Logger.Info("no .gitignore found or could not be read", "error", err.Error())
	}

	metas := []project.FileMeta{}
	err = fs.WalkDir(os.DirFS(in), ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		r.RunPlan.IncDiscovered(1)
		rel := filepath.ToSlash(path)
		if ig.MatchesPath(rel) {
			if d.IsDir() {
				r.RunPlan.IncDirs(1)
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			r.RunPlan.IncDirs(1)
		}
		fullPath := filepath.Join(in, path)
		r.RunPlan.MaxDepthSeen(depthFrom(in, fullPath))
		metas = append(metas, project.FileMeta{
			Path:  fullPath,
			Depth: depthFrom(in, fullPath),
			IsDir: d.IsDir(),
		})
		r.RunPlan.IncSelected(1)

		return nil
	})
	if err != nil {
		return r.RunPlan.Snapshot(), fmt.Errorf("internal: planRunner: failed to walk input path: %w", err)
	}

	return r.RunPlan.Snapshot(), nil
}

func buildDatabase(ctx context.Context, rc *project.RunContext) error {
	if rc == nil {
		return fmt.Errorf("internal: planRunner: run context unavailable")
	}
	cdb := rc.DB
	if cdb == nil {
		return fmt.Errorf("internal: planRunner: database unavailable")
	}
	err := db.RunMigrations(ctx, cdb)
	if err != nil {
		return fmt.Errorf("internal: planRunner: failed to run migrations: %w", err)
	}

	return nil
}

func depthFrom(root, path string) int {
	rel := strings.TrimPrefix(path, root)
	rel = strings.TrimPrefix(rel, string(os.PathSeparator))
	if rel == "" {
		return 0
	}
	return strings.Count(rel, string(os.PathSeparator))
}
