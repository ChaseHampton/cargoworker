package logx

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func New(opts Options) (*slog.Logger, func() error, error) {
	var hs []slog.Handler
	var closers []func() error

	if opts.ConsoleEnabled {
		lvl, _ := toSlogLevel(opts.ConsoleLevel)
		if opts.Quiet {
			lvl, _ = toSlogLevel(LevelError)
		}
		ho := &slog.HandlerOptions{Level: lvl, AddSource: opts.Source}
		w := os.Stderr
		var h slog.Handler
		if opts.ConsoleFormat == FormatJSON {
			h = slog.NewJSONHandler(w, ho)
		} else {
			h = slog.NewTextHandler(w, ho)
		}
		hs = append(hs, h)
	}

	if opts.FileEnabled {
		if opts.FilePath == "" {
			return nil, noopClose, invalid("file", "missing path")
		}
		if err := os.MkdirAll(filepath.Dir(opts.FilePath), 0o755); err != nil {
			return nil, noopClose, err
		}
		f, err := os.OpenFile(opts.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, noopClose, err
		}
		closers = append(closers, f.Close)

		lvl, _ := toSlogLevel(opts.FileLevel)
		ho := &slog.HandlerOptions{Level: lvl, AddSource: opts.Source}
		var fh slog.Handler
		if opts.FileFormat == FormatJSON {
			fh = slog.NewJSONHandler(f, ho)
		} else {
			fh = slog.NewTextHandler(f, ho)
		}
		hs = append(hs, fh)
	}

	mh := &multiHandler{handlers: hs}
	lg := slog.New(mh).With(
		slog.String("run_id", opts.RunID),
		slog.String("component", opts.Component),
	)

	cleanup := func() error {
		var err error
		for _, c := range closers {
			if e := c(); e != nil && err == nil {
				err = e
			}
		}
		return err
	}
	return lg, cleanup, nil
}

// func buildHandler(opts Options) (slog.Handler, func() error, error) {
// 	level, err := toSlogLevel(opts.Level)
// 	if err != nil {
// 		return nil, noopClose, err
// 	}
// 	switch opts.Format {
// 	case FormatJSON, FormatText:
// 	default:
// 		return nil, noopClose, invalid("format", string(opts.Format))
// 	}

// 	var writers []io.Writer
// 	writers = append(writers, os.Stderr)

// 	var file *os.File
// 	if opts.FilePath != "" {
// 		if err := os.MkdirAll(filepath.Dir(opts.FilePath), 0o755); err != nil {
// 			return nil, noopClose, err
// 		}
// 		f, err := os.OpenFile(opts.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
// 		if err != nil {
// 			return nil, noopClose, err
// 		}
// 		file = f
// 		writers = append(writers, file)
// 	}

// 	out := io.MultiWriter(writers...)
// 	ho := &slog.HandlerOptions{Level: level, AddSource: opts.Source}

// 	var h slog.Handler
// 	if opts.Format == FormatJSON {
// 		h = slog.NewJSONHandler(out, ho)
// 	} else {
// 		h = slog.NewTextHandler(out, ho)
// 	}

// 	closer := func() error {
// 		if file != nil {
// 			return file.Close()
// 		}
// 		return nil
// 	}
// 	return h, closer, nil
// }

func toSlogLevel(lvl Level) (slog.Leveler, error) {
	switch lvl {
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo, "":
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return nil, invalid("level", string(lvl))
	}
}

func invalid(kind, got string) error {
	return errors.New("logx: invalid " + kind + ": " + got)
}

func noopClose() error { return nil }

// KVErr standardizes error field naming. Always attach the raw error.
func KVErr(err error) slog.Attr { return slog.Any("err", err) }

// KVDuration is a typed duration field (not a string).
func KVDuration(d time.Duration) slog.Attr { return slog.Duration("dur", d) }

// KVPath normalizes a path under a chosen key.
func KVPath(key, p string) slog.Attr { return slog.String(key, p) }

// WithStartupBanner emits a one-time info banner about resolved logging opts.
// Call this right after New() if you want a predictable startup line.
func WithStartupBanner(lg *slog.Logger, opts Options) {
	lg.InfoContext(context.Background(), "startup",
		slog.String("console_format", string(opts.ConsoleFormat)),
		slog.String("console_level", string(opts.ConsoleLevel)),
		slog.Bool("console_enabled", opts.ConsoleEnabled),
		slog.String("file_format", string(opts.FileFormat)),
		slog.String("file_level", string(opts.FileLevel)),
		slog.Bool("file_enabled", opts.FileEnabled),
		slog.String("file_path", opts.FilePath),
		slog.Bool("source", opts.Source),
		slog.String("component", opts.Component),
	)
}
