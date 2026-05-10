package app

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
	"github.com/yorukot/netstamp/internal/probe/controlplane"
	"github.com/yorukot/netstamp/internal/probe/executor"
	dnsexecutor "github.com/yorukot/netstamp/internal/probe/executor/dns"
	httpexecutor "github.com/yorukot/netstamp/internal/probe/executor/http"
	pingexecutor "github.com/yorukot/netstamp/internal/probe/executor/ping"
	tcpexecutor "github.com/yorukot/netstamp/internal/probe/executor/tcp"
	"github.com/yorukot/netstamp/internal/probe/reporter"
	"github.com/yorukot/netstamp/internal/probe/runner"
	"github.com/yorukot/netstamp/internal/probe/scheduler"
)

type Application struct {
	controlplane *controlplane.Client
	scheduler    *scheduler.Scheduler
	runner       *runner.Runner
	reporter     *reporter.Reporter
}

func New() *Application {
	registry := executor.NewRegistry()
	registry.Register(probecontrol.CheckTypePing, pingexecutor.New())
	registry.Register(probecontrol.CheckTypeTCP, tcpexecutor.New())
	registry.Register(probecontrol.CheckTypeHTTP, httpexecutor.New())
	registry.Register(probecontrol.CheckTypeDNS, dnsexecutor.New())

	return &Application{
		controlplane: controlplane.NewClient(),
		scheduler:    scheduler.New(),
		runner:       runner.New(registry),
		reporter:     reporter.New(),
	}
}

func (a *Application) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = a.controlplane
	_ = a.scheduler
	_ = a.runner
	_ = a.reporter
	return nil
}
