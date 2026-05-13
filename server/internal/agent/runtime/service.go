package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/yorukot/netstamp/internal/agent/observability"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
)

type Service struct {
	client          RuntimeClient
	statusProvider  HeartbeatStatusProvider
	runtimeConfig   *RuntimeConfigStore
	assignments     *scheduling.AssignmentStore
	scheduler       *scheduling.Scheduler
	workers         *agentworker.WorkerPool
	results         *ResultSubmitter
	counters        *observability.RuntimeCounters
	shutdownTimeout time.Duration
	log             *slog.Logger
}

type ServiceDependencies struct {
	Client          RuntimeClient
	StatusProvider  HeartbeatStatusProvider
	RuntimeConfig   *RuntimeConfigStore
	Assignments     *scheduling.AssignmentStore
	Scheduler       *scheduling.Scheduler
	Workers         *agentworker.WorkerPool
	Results         *ResultSubmitter
	Counters        *observability.RuntimeCounters
	ShutdownTimeout time.Duration
	Log             *slog.Logger
}

func NewService(deps ServiceDependencies) *Service {
	return &Service{
		client:          deps.Client,
		statusProvider:  deps.StatusProvider,
		runtimeConfig:   deps.RuntimeConfig,
		assignments:     deps.Assignments,
		scheduler:       deps.Scheduler,
		workers:         deps.Workers,
		results:         deps.Results,
		counters:        deps.Counters,
		shutdownTimeout: deps.ShutdownTimeout,
		log:             deps.Log,
	}
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.startHello(ctx); err != nil {
		return err
	}

	runtimeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	group, groupCtx := errgroup.WithContext(runtimeCtx)
	group.Go(func() error { return ignoreCanceled(s.workers.Run(groupCtx)) })
	group.Go(func() error { return ignoreCanceled(s.scheduler.Run(groupCtx)) })
	group.Go(func() error { return ignoreCanceled(s.heartbeatLoop(groupCtx)) })
	group.Go(func() error { return ignoreCanceled(s.assignmentLoop(groupCtx)) })
	group.Go(func() error { return ignoreCanceled(s.results.Run(groupCtx)) })

	errCh := make(chan error, 1)
	go func() {
		errCh <- group.Wait()
	}()

	select {
	case <-ctx.Done():
		s.log.Info("probe agent draining", "shutdown_timeout", s.shutdownTimeout)
		cancel()
		return s.waitForShutdown(errCh)
	case err := <-errCh:
		cancel()
		return err
	}
}

func (s *Service) startHello(ctx context.Context) error {
	runtimeConfig := s.runtimeConfig.Get()
	backoff := runtimeConfig.InitialBackoff

	for attempt := 1; ; attempt++ {
		output, err := s.client.Hello(ctx)
		if err == nil {
			if err := EnsureMinimumVersion(Version, output.MinimumSupportedAgentVersion); err != nil {
				s.log.Error("probe agent version is not supported", "minimum_supported_agent_version", output.MinimumSupportedAgentVersion, "agent_version", AgentString)
				return err
			}
			applied := s.runtimeConfig.ApplyController(output.Config)
			s.log.Info("probe runtime hello succeeded", "server_time", output.ServerTime, "minimum_supported_agent_version", output.MinimumSupportedAgentVersion)
			s.log.Debug("runtime config applied", "heartbeat_interval", applied.HeartbeatInterval, "assignment_poll_interval", applied.AssignmentPollInterval, "initial_backoff", applied.InitialBackoff, "max_backoff", applied.MaxBackoff, "max_attempts", applied.MaxAttempts)
			return nil
		}
		if errors.Is(err, ErrAuthFailed) {
			s.counters.AuthFailures.Add(1)
			s.log.Error("probe runtime authentication failed during hello", "error", err)
			return err
		}
		if errors.Is(err, ErrPermanentRuntimeAPI) {
			s.log.Error("probe runtime hello failed permanently", "error", err)
			return err
		}

		s.log.Warn("probe runtime hello failed", "attempt", attempt, "backoff", backoff, "error", err)
		if sleepErr := sleepContext(ctx, backoff); sleepErr != nil {
			return sleepErr
		}
		backoff *= 2
		if backoff > runtimeConfig.MaxBackoff {
			backoff = runtimeConfig.MaxBackoff
		}
	}
}

func (s *Service) heartbeatLoop(ctx context.Context) error {
	for {
		if err := s.sendHeartbeat(ctx); err != nil {
			if errors.Is(err, ErrAuthFailed) {
				s.counters.AuthFailures.Add(1)
				return err
			}
			s.counters.HeartbeatErrors.Add(1)
			s.log.Warn("probe heartbeat failed", "error", err)
		}

		if err := sleepContext(ctx, s.runtimeConfig.Get().HeartbeatInterval); err != nil {
			return err
		}
	}
}

func (s *Service) sendHeartbeat(ctx context.Context) error {
	_, err := s.client.Heartbeat(ctx, s.statusProvider.Status())
	return err
}

func (s *Service) assignmentLoop(ctx context.Context) error {
	for {
		if err := s.pullAssignments(ctx); err != nil {
			if errors.Is(err, ErrAuthFailed) {
				s.counters.AuthFailures.Add(1)
				return err
			}
			s.counters.AssignmentPullErrors.Add(1)
			s.log.Warn("probe assignment pull failed", "error", err)
		}

		if err := sleepContext(ctx, s.runtimeConfig.Get().AssignmentPollInterval); err != nil {
			return err
		}
	}
}

func (s *Service) pullAssignments(ctx context.Context) error {
	output, err := s.client.ListAssignments(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	s.runtimeConfig.ApplyController(output.Config)
	summary := s.assignments.Reconcile(output.Assignments, now)
	s.scheduler.Wake()
	s.log.Info("probe assignments reconciled", "active", summary.Active, "added", summary.Added, "updated", summary.Updated, "removed", summary.Removed, "unsupported", summary.Unsupported, "server_time", output.ServerTime)

	return nil
}

func (s *Service) waitForShutdown(errCh <-chan error) error {
	timer := time.NewTimer(s.shutdownTimeout)
	defer timer.Stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("probe agent shutdown timed out after %s", s.shutdownTimeout)
	}
}

func ignoreCanceled(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
