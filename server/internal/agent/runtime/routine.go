package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
	"github.com/yorukot/netstamp/internal/agent/retry"
)

func (s *Service) heartbeatLoop(ctx context.Context) error {
	for {
		status := agentStatus()
		if _, err := s.Client.Heartbeat(ctx, status); err != nil {
			if errors.Is(err, httpclient.ErrAuthFailed) {
				return err
			}
			s.Log.Warn("probe heartbeat failed", "error", err)
		}

		// Wait for next heartbeat interval or cancel if contextx is dead
		if err := retry.WaitForDuration(ctx, s.Config.HeartbeatInterval.Value); err != nil {
			return err
		}
	}
}


func (s *Service) assignmentLoop(ctx context.Context) error {
	for {
		if err := s.pullAssignments(ctx); err != nil {
			if errors.Is(err, httpclient.ErrAuthFailed) {
				return err
			}
			s.Log.Warn("probe assignment pull failed", "error", err)
		}

		if err := retry.WaitForDuration(ctx, s.Config.AssignmentPollInterval.Value); err != nil {
			return err
		}
	}
}

func (s *Service) pullAssignments(ctx context.Context) error {
	output, err := s.Client.ListAssignments(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	// reconcile assignments with the server
	summary := s.Assignments.Reconcile(output.Assignments, now)

	// wake the scheduler to process any new assignments
	s.Scheduler.Wake()
	s.Log.Info("probe assignments reconciled", "active", summary.Active, "added", summary.Added, "updated", summary.Updated, "removed", summary.Removed, "unsupported", summary.Unsupported, "server_time", output.ServerTime)

	return nil
}
