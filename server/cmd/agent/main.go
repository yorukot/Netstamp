package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	agentapp "github.com/yorukot/netstamp/internal/agent/app"
	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := agentconfig.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load probe config: %v\n", err)
		return 1
	}

	log := agentapp.NewLogger(config.LogLevel)
	log.LogAttrs(ctx, slog.LevelInfo, "probe agent starting", config.SafeLogAttrs()...)

	app := agentapp.New(config, log)
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("probe agent stopped with error", "error", err)
		return 1
	}

	log.Info("probe agent stopped")
	return 0
}
