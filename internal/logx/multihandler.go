package logx

import (
	"context"
	"log/slog"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		// Each Handle gets a fresh copy since slog.Record is single-use iterator for Attrs.
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nhs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		nhs[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: nhs}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	nhs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		nhs[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: nhs}
}
