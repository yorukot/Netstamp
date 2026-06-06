package agentruntime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
	"github.com/yorukot/netstamp/internal/agent/result"
	"github.com/yorukot/netstamp/internal/agent/retry"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
)

type Service struct {
	Client      httpclient.RuntimeClient
	Config      *config.Config
	Assignments *scheduling.AssignmentStore
	Scheduler   *scheduling.Scheduler
	Workers     *agentworker.WorkerPool
	Results     *result.Submitter
	Log         *slog.Logger
}

func (s *Service) Run(ctx context.Context) error {
	// First, we need to authenticate with the runtime API
	if err := s.startHello(ctx); err != nil {
		return err
	}

	// Once authenticated, start the runtime loops with one shared context across the service layer.
	runtimeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	group, groupCtx := errgroup.WithContext(runtimeCtx)
	// initialize the heartbeat loop first, send the heartbeat to the server
	group.Go(func() error { return ignoreCanceled(s.heartbeatLoop(groupCtx)) })
	// report IPv4/IPv6 public reachability separately from heartbeat liveness
	group.Go(func() error { return ignoreCanceled(s.ipFamilyCapabilityLoop(groupCtx)) })
	// pull assignments from the server periodically
	group.Go(func() error { return ignoreCanceled(s.assignmentLoop(groupCtx)) })
	// run the scheduler to assign work to workers
	group.Go(func() error { return ignoreCanceled(s.Scheduler.Run(groupCtx)) })
	// run the worker pool to execute the assigned work and put it to the result queue
	group.Go(func() error { return ignoreCanceled(s.Workers.Run(groupCtx)) })
	// submit results to the server periodically
	group.Go(func() error { return ignoreCanceled(s.Results.Run(groupCtx)) })

	errCh := make(chan error, 1)
	go func() {
		errCh <- group.Wait()
	}()

	select {
	case <-ctx.Done():
		s.Log.Info("probe agent draining", "shutdown_timeout", s.Config.ShutdownTimeout.Value)
		cancel()
		return s.waitForShutdown(errCh)
	case err := <-errCh:
		cancel()
		return err
	}
}

func (s *Service) startHello(ctx context.Context) error {
	backoff := s.Config.InitialBackoff.Value

	for attempt := 1; ; attempt++ {
		output, err := s.Client.Hello(ctx)
		if err == nil {
			if versionErr := EnsureMinimumVersion(Version, output.MinimumSupportedAgentVersion); versionErr != nil {
				s.Log.Error("probe agent version is not supported", "minimum_supported_agent_version", output.MinimumSupportedAgentVersion, "agent_version", AgentString)
				return versionErr
			}
			s.Log.Info("probe runtime hello succeeded", "server_time", output.ServerTime, "minimum_supported_agent_version", output.MinimumSupportedAgentVersion)
			return nil
		}
		if errors.Is(err, httpclient.ErrAuthFailed) {
			s.Log.Error("probe runtime authentication failed during hello", "error", err)
			return err
		}
		if errors.Is(err, httpclient.ErrPermanentRuntimeAPI) {
			s.Log.Error("probe runtime hello failed permanently", "error", err)
			return err
		}

		s.Log.Warn("probe runtime hello failed", "attempt", attempt, "backoff", backoff, "error", err)
		if sleepErr := retry.WaitForDuration(ctx, backoff); sleepErr != nil {
			return sleepErr
		}
		// Double the backoff until it reaches the configured maximum.
		backoff *= 2
		if backoff > s.Config.MaxBackoff.Value {
			backoff = s.Config.MaxBackoff.Value
		}
	}
}

// waitForShutdown waits for the shutdown timeout or an error from the error channel.
func (s *Service) waitForShutdown(errCh <-chan error) error {
	timer := time.NewTimer(s.Config.ShutdownTimeout.Value)
	defer timer.Stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("probe agent shutdown timed out after %s", s.Config.ShutdownTimeout.Value)
	}
}

func ignoreCanceled(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
