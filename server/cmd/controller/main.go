package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/yorukot/netstamp/internal/controller/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Graceful shutdown on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// New application with the context
	application, err := app.New(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		return 1
	}
	defer func() {
		if syncErr := application.Log.Sync(); syncErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "sync logger: %v\n", syncErr)
		}
	}()

	err = application.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		application.Log.Error("startup failed", zap.Error(err))
		return 1
	}

	return 0
}
