package cli

import (
	"fmt"
	"strings"

	"github.com/ChaseHampton/cargoworker/internal/plan"
	"github.com/ChaseHampton/cargoworker/internal/project"
	"github.com/ChaseHampton/cargoworker/internal/stats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPlanCmd() *cobra.Command {
	var (
		fLang     string
		fIgnore   []string
		fWithDeps bool
	)

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Discover the project and prepare a plan (stub for now)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rc := project.FromContext(cmd.Context())
			if rc == nil {
				return fmt.Errorf("internal: run context unavailable")
			}
			in := project.InputPathFrom(cmd.Context())
			if in == "" {
				return fmt.Errorf("internal: input path not resolved")
			}

			// Resolve options (flags > env > defaults via Viper)
			lang := strings.TrimSpace(viper.GetString("plan.language"))
			ignore := viper.GetStringSlice("plan.ignore")
			withDeps := viper.GetBool("plan.with_deps")

			rc.Logger.Info("plan start",
				"run_id", rc.RunId, "in", in,
				"language", lang, "ignore", ignore, "with_deps", withDeps)

			planStats := stats.NewPlan(in, []string{}, ignore)
			runner := plan.NewRunner(planStats)
			snap, err := runner.Plan(cmd.Context())
			if err != nil {
				return fmt.Errorf("plan failed: %w", err)
			}
			rc.Logger.Info("plan completed (stub)")
			rc.Stats.SetPlan(snap)

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&fLang, "language", "go", "language to plan (default: go)")
	cmd.Flags().StringSliceVar(&fIgnore, "ignore", nil, "comma- or repeatable list of globs to ignore")
	cmd.Flags().BoolVar(&fWithDeps, "with-deps", false, "include module/package dependencies in planning")

	// Viper bindings (env keys: CARGOWORKER_PLAN_LANGUAGE, _PLAN_IGNORE, _PLAN_WITH_DEPS)
	_ = viper.BindPFlag("plan.language", cmd.Flags().Lookup("language"))
	_ = viper.BindPFlag("plan.ignore", cmd.Flags().Lookup("ignore"))
	_ = viper.BindPFlag("plan.with_deps", cmd.Flags().Lookup("with-deps"))

	// Sensible defaults (so env-only works)
	viper.SetDefault("plan.language", "go")
	viper.SetDefault("plan.ignore", []string{})
	viper.SetDefault("plan.with_deps", false)

	return cmd
}
