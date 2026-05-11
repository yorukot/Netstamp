package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/probe/config"
	"github.com/yorukot/netstamp/internal/probe/controlplane"
	"github.com/yorukot/netstamp/internal/probe/executor"
	pingexecutor "github.com/yorukot/netstamp/internal/probe/executor/ping"
	"github.com/yorukot/netstamp/internal/probe/reporter"
	"github.com/yorukot/netstamp/internal/probe/runner"
	"github.com/yorukot/netstamp/internal/probe/scheduler"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultPollInterval      = 30 * time.Second
	executionTick            = time.Second
)

type Application struct {
	cfg          config.Config
	controlplane *controlplane.Client
	scheduler    *scheduler.Scheduler
	runner       *runner.Runner
	reporter     *reporter.Reporter
	log          *log.Logger
}

func New() (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	registry := executor.NewRegistry()
	registry.Register(domaincheck.TypePing, pingexecutor.New())

	client, err := controlplane.NewClient(cfg.ControllerURL, cfg.ProbeID, cfg.ProbeSecret, cfg.HTTPTimeout)
	if err != nil {
		return nil, err
	}

	return &Application{
		cfg:          cfg,
		controlplane: client,
		scheduler:    scheduler.New(),
		runner:       runner.New(registry, cfg.MaxWorkers),
		reporter:     reporter.New(),
		log:          log.New(os.Stderr, "probe: ", log.LstdFlags),
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status := controlplane.StatusInput{
		AgentVersion: a.cfg.AgentVersion,
		Addrs:        localAddresses(),
	}
	hello, err := a.controlplane.Hello(ctx, status)
	if err != nil {
		return fmt.Errorf("start probe runtime session: %w", err)
	}
	if err := a.pollAssignments(ctx); err != nil {
		return fmt.Errorf("poll initial assignments: %w", err)
	}

	heartbeatInterval := intervalFromSeconds(hello.HeartbeatIntervalSeconds, defaultHeartbeatInterval)
	pollInterval := intervalFromSeconds(hello.AssignmentPollIntervalSeconds, defaultPollInterval)
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return a.heartbeatLoop(ctx, status, heartbeatInterval)
	})
	group.Go(func() error {
		return a.pollLoop(ctx, pollInterval)
	})
	group.Go(func() error {
		return a.executionLoop(ctx)
	})

	return group.Wait()
}

func (a *Application) heartbeatLoop(ctx context.Context, status controlplane.StatusInput, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.controlplane.Heartbeat(ctx, status); err != nil && ctx.Err() == nil {
				a.log.Printf("heartbeat failed: %v", err)
			}
		}
	}
}

func (a *Application) pollLoop(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.pollAssignments(ctx); err != nil && ctx.Err() == nil {
				a.log.Printf("poll assignments failed: %v", err)
			}
		}
	}
}

func (a *Application) executionLoop(ctx context.Context) error {
	ticker := time.NewTicker(executionTick)
	defer ticker.Stop()

	var pending []domainprobe.Result
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			results, err := a.runDueAssignments(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				a.log.Printf("run assignments failed: %v", err)
			}
			pending = append(pending, results...)
			pending = a.flushResults(ctx, pending)
		}
	}
}

func (a *Application) pollAssignments(ctx context.Context) error {
	assignments, err := a.controlplane.PollAssignments(ctx)
	if err != nil {
		return err
	}
	a.scheduler.Replace(assignments.Assignments, time.Now().UTC())

	return nil
}

func (a *Application) runDueAssignments(ctx context.Context) ([]domainprobe.Result, error) {
	now := time.Now().UTC()
	assignments, err := a.scheduler.Due(ctx, now)
	if err != nil || len(assignments) == 0 {
		return nil, err
	}

	results, runErr := a.runner.Run(ctx, assignments)
	a.scheduler.Complete(assignments, time.Now().UTC())

	return results, runErr
}

func (a *Application) flushResults(ctx context.Context, pending []domainprobe.Result) []domainprobe.Result {
	if len(pending) == 0 {
		return nil
	}

	batches := a.reporter.Batches(a.cfg.ProbeID, pending)
	sent := 0
	for _, batch := range batches {
		output, err := a.controlplane.SubmitResults(ctx, batch)
		if err != nil {
			if ctx.Err() == nil {
				a.log.Printf("submit results failed: %v", err)
			}
			break
		}
		sent += len(batch.Results)
		if output.ResyncNeeded {
			if err := a.pollAssignments(ctx); err != nil && ctx.Err() == nil {
				a.log.Printf("resync assignments failed: %v", err)
			}
		}
	}
	if sent == len(pending) {
		return nil
	}

	return append([]domainprobe.Result(nil), pending[sent:]...)
}

func intervalFromSeconds(seconds int32, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func localAddresses() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	values := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		values = append(values, ip.String())
	}

	return values
}
