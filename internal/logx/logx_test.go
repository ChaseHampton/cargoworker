// internal/logx/logx_test.go
package logx

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/google/uuid"
)

func TestNew_JSONAndLevel(t *testing.T) {
	var buf bytes.Buffer
	lh := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	_ = lh // silence unused; this test is here as a scaffold if you route output to buffers later.

	lg, closeFn, err := New(Options{
		ConsoleFormat: FormatJSON,
		ConsoleLevel:  LevelInfo,
		RunID:         uuid.New().String(),
		Component:     "test",
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer closeFn()

	lg.Debug("hidden") // should be filtered
	lg.Info("visible")
}

func TestInvalidOptions(t *testing.T) {
	_, _, err := New(Options{ConsoleFormat: "xml"})
	if err == nil {
		t.Fatalf("expected error for invalid format")
	}
	_, _, err = New(Options{ConsoleFormat: FormatJSON, ConsoleLevel: "loud"})
	if err == nil {
		t.Fatalf("expected error for invalid level")
	}
}
