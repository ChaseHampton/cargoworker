package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChaseHampton/cargoworker/internal/cli"
	"github.com/ChaseHampton/cargoworker/internal/project"
)

func main() {
	// Context that cancels on Ctrl-C or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bus := project.NewEventBus(1024)
	defer bus.Close()

	bootstrap := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	bootstrap.Info("cargoworker starting", "pid", os.Getpid())

	cmd := cli.NewRootCmd(cli.Deps{
		BootstrapLogger: bootstrap,
		EventChan:       bus.Sink(),
	})

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		bootstrap.Error("execution failed", "error", err)
		os.Exit(1)
	}
}
