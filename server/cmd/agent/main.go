package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	agentcli "github.com/yorukot/netstamp/internal/agent/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(agentcli.Execute(ctx, agentcli.Options{
		Args:   os.Args[1:],
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}))
}
