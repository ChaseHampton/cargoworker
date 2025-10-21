package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ChaseHampton/cargoworker/internal/logx"
	"github.com/ChaseHampton/cargoworker/internal/project"
	"github.com/ChaseHampton/cargoworker/internal/stats"
)

type Deps struct {
	BootstrapLogger *slog.Logger
	EventChan       chan<- project.Event
}

func NewRootCmd(deps Deps) *cobra.Command {
	var (
		fConsole, fFile, fQuiet       bool
		fConsoleFmt, fConsoleLvl      string
		fFilePath, fFileFmt, fFileLvl string
		fRunID, fOut, fIn             string
	)

	cmd := &cobra.Command{
		Use:           "cargoworker [PATH]",
		Short:         "cargoworker processes code bases and generates documentation",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			inPath := strings.TrimSpace(viper.GetString("in"))
			if inPath == "" && len(args) > 0 {
				inPath = args[0]
			}
			if inPath == "" {
				return fmt.Errorf("missing input path: provide [PATH] or --in or CARGOWORKER_IN")
			}

			if p, err := filepath.Abs(inPath); err == nil {
				inPath = p
			}
			if fi, err := os.Stat(inPath); err != nil || !fi.IsDir() {
				return fmt.Errorf("input path not found or not a directory: %s", inPath)
			}

			opts := logx.Options{
				ConsoleEnabled: viper.GetBool("log.console"),
				ConsoleFormat:  logx.Format(viper.GetString("log.console.format")),
				ConsoleLevel:   logx.Level(viper.GetString("log.console.level")),
				FileEnabled:    viper.GetBool("log.file"),
				FilePath:       viper.GetString("log.file.path"),
				FileFormat:     logx.Format(viper.GetString("log.file.format")),
				FileLevel:      logx.Level(viper.GetString("log.file.level")),
				Source:         true, // or viper.GetBool("log.source") if you expose it
				RunID:          viper.GetString("run-id"),
				Component:      "cargoworker",
				Quiet:          viper.GetBool("quiet"),
			}

			if opts.RunID == "" {
				opts.RunID = uuid.New().String()
			}

			outRoot := viper.GetString("out")
			if outRoot == "" {
				outRoot = "./cargoworkerout"
			}
			runOut := filepath.Join(outRoot, opts.RunID)
			if err := os.MkdirAll(runOut, 0o755); err != nil {
				return err
			}
			if opts.FileEnabled {
				if opts.FilePath == "" {
					if err := os.MkdirAll(filepath.Join(runOut, "logs"), 0o755); err != nil {
						return err
					}
					opts.FilePath = filepath.Join(runOut, "logs", "run.log")
				} else {
					if err := os.MkdirAll(filepath.Dir(opts.FilePath), 0o755); err != nil {
						return err
					}
				}
			}

			final, closer, err := logx.New(opts)
			if err != nil {
				return fmt.Errorf("setup logging: %w", err)
			}

			var runUUID uuid.UUID
			if u, err := uuid.Parse(opts.RunID); err == nil {
				runUUID = u
			} else {
				runUUID = uuid.New()
			}

			st := stats.New(
				stats.WithRunID(opts.RunID),
				stats.WithInputPath(inPath),
				stats.WithOutDir(runOut),
				stats.WithToolVersion("v1"),
				stats.WithIRSchema("v1"),
			)

			rc := &project.RunContext{
				RunId:       runUUID,
				OutDir:      runOut,
				ToolVersion: "v1",
				IRSchema:    "v1",             // fill in later
				Limits:      project.Limits{}, // fill in later
				Logger:      final,
				DB:          nil,
				Events:      deps.EventChan,
				Stats:       st,
				Closers:     []func() error{closer},
			}

			ctx := project.WithRunContext(cmd.Context(), rc)
			ctx = project.WithInputPath(ctx, inPath)
			cmd.SetContext(ctx)

			// A single handoff log
			final.Info("logger initialized",
				"run_id", opts.RunID, "out", runOut,
				"console", opts.ConsoleEnabled, "file", opts.FileEnabled,
				"file_path", opts.FilePath, "level", opts.ConsoleLevel)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			rc := project.FromContext(cmd.Context())
			if rc == nil {
				return nil
			}

			rc.Logger.Info("shutdown complete", "run_id", rc.RunId)

			var cerr error
			for i := len(rc.Closers) - 1; i >= 0; i-- {
				if rc.Closers[i] == nil {
					continue
				}
				if err := rc.Closers[i](); err != nil {
					rc.Logger.Warn("close error", "err", err)
					cerr = errors.Join(cerr, err)
				}
			}
			return cerr
		},
	}

	pf := cmd.PersistentFlags()
	pf.BoolVar(&fConsole, "log.console", true, "enable console logging")
	pf.StringVar(&fConsoleFmt, "log.console.format", "text", "console log format: text|json")
	pf.StringVar(&fConsoleLvl, "log.console.level", "info", "console log level")
	pf.BoolVar(&fFile, "log.file", true, "enable file logging")
	pf.StringVar(&fFilePath, "log.file.path", "", "file log path (default: OUT/RUNID/logs/run.log)")
	pf.StringVar(&fFileFmt, "log.file.format", "json", "file log format: text|json")
	pf.StringVar(&fFileLvl, "log.file.level", "debug", "file log level")
	pf.StringVar(&fRunID, "run-id", "", "override run id")
	pf.StringVar(&fOut, "out", "./out", "output root directory")
	pf.StringVar(&fIn, "in", "", "input path (falls back to positional [PATH])")
	pf.BoolVar(&fQuiet, "quiet", false, "suppress console output below errors")

	// ----- Viper binding
	viper.SetEnvPrefix("CARGOWORKER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// Defaults (kept in Viper so env-only works)
	viper.SetDefault("log.console", true)
	viper.SetDefault("log.console.format", "text")
	viper.SetDefault("log.console.level", "info")
	viper.SetDefault("log.file", true)
	viper.SetDefault("log.file.format", "json")
	viper.SetDefault("log.file.level", "debug")
	viper.SetDefault("out", "./cargoworkerout")
	viper.SetDefault("in", "")
	viper.SetDefault("quiet", false)

	_ = viper.BindPFlag("log.console", pf.Lookup("log.console"))
	_ = viper.BindPFlag("log.console.format", pf.Lookup("log.console.format"))
	_ = viper.BindPFlag("log.console.level", pf.Lookup("log.console.level"))
	_ = viper.BindPFlag("log.file", pf.Lookup("log.file"))
	_ = viper.BindPFlag("log.file.path", pf.Lookup("log.file.path"))
	_ = viper.BindPFlag("log.file.format", pf.Lookup("log.file.format"))
	_ = viper.BindPFlag("log.file.level", pf.Lookup("log.file.level"))
	_ = viper.BindPFlag("run-id", pf.Lookup("run-id"))
	_ = viper.BindPFlag("out", pf.Lookup("out"))
	_ = viper.BindPFlag("in", pf.Lookup("in"))
	_ = viper.BindPFlag("quiet", pf.Lookup("quiet"))

	cmd.AddCommand(NewPlanCmd())
	return cmd
}
