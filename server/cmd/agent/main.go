package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	agentcli "github.com/yorukot/netstamp/internal/agent/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return agentcli.Execute(ctx, agentcli.Options{
		Args:   os.Args[1:],
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
