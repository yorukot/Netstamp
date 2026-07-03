package assignment

import (
	"context"
	"errors"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type WorkerConfig struct {
	Enabled       bool
	Interval      time.Duration
	BatchSize     int32
	StaleTimeout  time.Duration
	RetryBackoffs []time.Duration
}

type Worker struct {
	repo Repository
	cfg  WorkerConfig
	now  func() time.Time
}

func NewWorker(repo Repository, cfg WorkerConfig) *Worker {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 25
	}
	if cfg.StaleTimeout <= 0 {
		cfg.StaleTimeout = time.Minute
	}
	if len(cfg.RetryBackoffs) == 0 {
		cfg.RetryBackoffs = []time.Duration{30 * time.Second, 2 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	}
	return &Worker{repo: repo, cfg: cfg, now: func() time.Time { return time.Now().UTC() }}
}

func (w *Worker) Run(ctx context.Context) error {
	if !w.cfg.Enabled {
		<-ctx.Done()
		return nil
	}
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		if err := w.RunOnce(ctx); err != nil {
			// Keep the worker alive; individual job errors are persisted in the job table.
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	jobs, err := w.repo.ClaimRefreshJobs(ctx, w.cfg.BatchSize, w.now().Add(-w.cfg.StaleTimeout))
	if err != nil {
		return err
	}
	var errs []error
	for _, job := range jobs {
		if err := w.process(ctx, job); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (w *Worker) process(ctx context.Context, job domainassignment.RefreshJob) error {
	err := w.refreshTarget(ctx, job)
	if err == nil {
		return w.repo.MarkRefreshJobSucceeded(ctx, job.ID, w.now())
	}

	if terminal, markErr := w.markTerminal(ctx, job, err); terminal {
		return markErr
	}
	if job.AttemptCount+1 >= job.MaxAttempts {
		return w.repo.MarkRefreshJobFailed(ctx, job.ID, "assignment_refresh", "refresh_failed", "assignment refresh failed")
	}
	return w.repo.MarkRefreshJobRetry(ctx, job.ID, w.nextAttempt(job), "assignment_refresh", "refresh_failed", "assignment refresh failed")
}

func (w *Worker) refreshTarget(ctx context.Context, job domainassignment.RefreshJob) error {
	switch job.TargetType {
	case domainassignment.RefreshTargetProject:
		return w.repo.RefreshProbeCheckAssignmentsForProject(ctx, job.ProjectID)
	case domainassignment.RefreshTargetProbe:
		err := w.repo.RefreshProbeCheckAssignmentsForProbe(ctx, job.ProjectID, job.TargetID)
		if errors.Is(err, domainprobe.ErrProbeNotFound) {
			if cleanupErr := w.repo.DeleteProbeCheckAssignmentsForProbe(ctx, job.ProjectID, job.TargetID); cleanupErr != nil {
				return cleanupErr
			}
		}
		return err
	case domainassignment.RefreshTargetCheck:
		err := w.repo.RefreshProbeCheckAssignmentsForCheck(ctx, job.ProjectID, job.TargetID)
		if errors.Is(err, domaincheck.ErrCheckNotFound) {
			if cleanupErr := w.repo.DeleteProbeCheckAssignmentsForCheck(ctx, job.ProjectID, job.TargetID); cleanupErr != nil {
				return cleanupErr
			}
		}
		return err
	case domainassignment.RefreshTargetLabel:
		return w.repo.RefreshProbeCheckAssignmentsForLabel(ctx, job.ProjectID, job.TargetID)
	default:
		return domainassignment.ErrInvalidInput
	}
}

func (w *Worker) markTerminal(ctx context.Context, job domainassignment.RefreshJob, err error) (bool, error) {
	switch {
	case errors.Is(err, domainassignment.ErrInvalidInput):
		return true, w.repo.MarkRefreshJobDiscarded(ctx, job.ID, "assignment_refresh", "invalid_target", "assignment refresh job target is invalid")
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return true, w.repo.MarkRefreshJobDiscarded(ctx, job.ID, "assignment_refresh", "project_not_found", "assignment refresh project was not found")
	case errors.Is(err, domainprobe.ErrProbeNotFound), errors.Is(err, domaincheck.ErrCheckNotFound), errors.Is(err, domainlabel.ErrLabelNotFound):
		return true, w.repo.MarkRefreshJobDiscarded(ctx, job.ID, "assignment_refresh", "target_not_found", "assignment refresh target was not found")
	default:
		return false, nil
	}
}

func (w *Worker) nextAttempt(job domainassignment.RefreshJob) time.Time {
	index := int(job.AttemptCount)
	if index >= len(w.cfg.RetryBackoffs) {
		index = len(w.cfg.RetryBackoffs) - 1
	}
	return w.now().Add(w.cfg.RetryBackoffs[index])
}
