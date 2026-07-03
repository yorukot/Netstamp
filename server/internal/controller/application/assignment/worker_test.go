package assignment

import (
	"context"
	"errors"
	"testing"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const testRefreshJobID = "55555555-5555-5555-5555-555555555555"

func TestWorkerMarksSuccessfulRefreshJobsSucceeded(t *testing.T) {
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	repo := &workerRepository{jobs: []domainassignment.RefreshJob{testRefreshJob(domainassignment.RefreshTargetProject)}}
	worker := newTestWorker(repo, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	if repo.refreshMethod != "project" || repo.projectID != testProjectID {
		t.Fatalf("unexpected refresh call: %#v", repo)
	}
	if repo.method != "succeeded" || repo.jobID != testRefreshJobID || !repo.completedAt.Equal(now) {
		t.Fatalf("unexpected succeeded mark: %#v", repo)
	}
}

func TestWorkerRetriesRefreshFailures(t *testing.T) {
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	job := testRefreshJob(domainassignment.RefreshTargetCheck)
	job.AttemptCount = 1
	job.MaxAttempts = 3
	repo := &workerRepository{jobs: []domainassignment.RefreshJob{job}, refreshErr: errors.New("database unavailable")}
	worker := newTestWorker(repo, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	wantNextAttempt := now.Add(2 * time.Minute)
	if repo.method != "retry" || repo.jobID != testRefreshJobID || !repo.nextAttemptAt.Equal(wantNextAttempt) {
		t.Fatalf("unexpected retry mark: %#v", repo)
	}
	assertWorkerResultDetails(t, repo, "assignment_refresh", "refresh_failed", "assignment refresh failed")
}

func TestWorkerFailsRefreshFailuresAtMaxAttempts(t *testing.T) {
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	job := testRefreshJob(domainassignment.RefreshTargetCheck)
	job.AttemptCount = 2
	job.MaxAttempts = 3
	repo := &workerRepository{jobs: []domainassignment.RefreshJob{job}, refreshErr: errors.New("database unavailable")}
	worker := newTestWorker(repo, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	if repo.method != "failed" || repo.jobID != testRefreshJobID {
		t.Fatalf("unexpected failure mark: %#v", repo)
	}
	assertWorkerResultDetails(t, repo, "assignment_refresh", "refresh_failed", "assignment refresh failed")
}

func TestWorkerDiscardsMissingProbeAfterCleanup(t *testing.T) {
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	repo := &workerRepository{
		jobs:       []domainassignment.RefreshJob{testRefreshJob(domainassignment.RefreshTargetProbe)},
		refreshErr: domainprobe.ErrProbeNotFound,
	}
	worker := newTestWorker(repo, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	if repo.deleteMethod != "probe" || repo.probeID != testProbeID {
		t.Fatalf("expected missing probe cleanup, got %#v", repo)
	}
	if repo.method != "discarded" || repo.jobID != testRefreshJobID {
		t.Fatalf("unexpected discard mark: %#v", repo)
	}
	assertWorkerResultDetails(t, repo, "assignment_refresh", "target_not_found", "assignment refresh target was not found")
}

func TestWorkerRetriesWhenMissingCheckCleanupFails(t *testing.T) {
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	cleanupErr := errors.New("cleanup failed")
	repo := &workerRepository{
		jobs:       []domainassignment.RefreshJob{testRefreshJob(domainassignment.RefreshTargetCheck)},
		refreshErr: domaincheck.ErrCheckNotFound,
		deleteErr:  cleanupErr,
	}
	worker := newTestWorker(repo, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to persist retry and return nil: %v", err)
	}

	if repo.deleteMethod != "check" || repo.checkID != testCheckID {
		t.Fatalf("expected missing check cleanup, got %#v", repo)
	}
	if repo.method != "retry" {
		t.Fatalf("expected cleanup failure to retry, got %#v", repo)
	}
}

func TestDisabledAssignmentWorkerWaitsForContextAndReturnsNil(t *testing.T) {
	worker := NewWorker(nil, WorkerConfig{Enabled: false})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("expected worker to wait for context, returned early with %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected worker to return after context cancellation")
	}
}

func newTestWorker(repo *workerRepository, now time.Time) *Worker {
	worker := NewWorker(repo, WorkerConfig{
		Enabled:       true,
		Interval:      time.Hour,
		BatchSize:     10,
		StaleTimeout:  time.Minute,
		RetryBackoffs: []time.Duration{30 * time.Second, 2 * time.Minute},
	})
	worker.now = func() time.Time { return now }
	return worker
}

func testRefreshJob(targetType domainassignment.RefreshTargetType) domainassignment.RefreshJob {
	targetID := testProjectID
	switch targetType {
	case domainassignment.RefreshTargetProbe:
		targetID = testProbeID
	case domainassignment.RefreshTargetCheck:
		targetID = testCheckID
	case domainassignment.RefreshTargetLabel:
		targetID = testLabelID
	}
	return domainassignment.RefreshJob{
		ID:           testRefreshJobID,
		ProjectID:    testProjectID,
		TargetType:   targetType,
		TargetID:     targetID,
		AttemptCount: 0,
		MaxAttempts:  3,
	}
}

func assertWorkerResultDetails(t *testing.T, repo *workerRepository, kind, code, message string) {
	t.Helper()

	if repo.kind != kind || repo.code != code || repo.message != message {
		t.Fatalf("unexpected result details: kind=%q code=%q message=%q", repo.kind, repo.code, repo.message)
	}
}

type workerRepository struct {
	jobs          []domainassignment.RefreshJob
	claimErr      error
	refreshErr    error
	deleteErr     error
	method        string
	refreshMethod string
	deleteMethod  string
	jobID         string
	projectID     string
	probeID       string
	checkID       string
	labelID       string
	completedAt   time.Time
	nextAttemptAt time.Time
	kind          string
	code          string
	message       string
}

func (r *workerRepository) EnqueueRefreshJob(_ context.Context, _ domainassignment.RefreshTarget, _ int32) error {
	return nil
}

func (r *workerRepository) ClaimRefreshJobs(_ context.Context, _ int32, _ time.Time) ([]domainassignment.RefreshJob, error) {
	return r.jobs, r.claimErr
}

func (r *workerRepository) MarkRefreshJobSucceeded(_ context.Context, id string, at time.Time) error {
	r.method = "succeeded"
	r.jobID = id
	r.completedAt = at
	return nil
}

func (r *workerRepository) MarkRefreshJobRetry(_ context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error {
	r.method = "retry"
	r.jobID = id
	r.nextAttemptAt = nextAttemptAt
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

func (r *workerRepository) MarkRefreshJobFailed(_ context.Context, id, kind, code, message string) error {
	r.method = "failed"
	r.jobID = id
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

func (r *workerRepository) MarkRefreshJobDiscarded(_ context.Context, id, kind, code, message string) error {
	r.method = "discarded"
	r.jobID = id
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

func (r *workerRepository) RefreshProbeCheckAssignmentsForProject(_ context.Context, projectID string) error {
	r.refreshMethod = "project"
	r.projectID = projectID
	return r.refreshErr
}

func (r *workerRepository) RefreshProbeCheckAssignmentsForProbe(_ context.Context, projectID, probeID string) error {
	r.refreshMethod = "probe"
	r.projectID = projectID
	r.probeID = probeID
	return r.refreshErr
}

func (r *workerRepository) RefreshProbeCheckAssignmentsForCheck(_ context.Context, projectID, checkID string) error {
	r.refreshMethod = "check"
	r.projectID = projectID
	r.checkID = checkID
	return r.refreshErr
}

func (r *workerRepository) RefreshProbeCheckAssignmentsForLabel(_ context.Context, projectID, labelID string) error {
	r.refreshMethod = "label"
	r.projectID = projectID
	r.labelID = labelID
	return r.refreshErr
}

func (r *workerRepository) DeleteProbeCheckAssignmentsForProbe(_ context.Context, projectID, probeID string) error {
	r.deleteMethod = "probe"
	r.projectID = projectID
	r.probeID = probeID
	return r.deleteErr
}

func (r *workerRepository) DeleteProbeCheckAssignmentsForCheck(_ context.Context, projectID, checkID string) error {
	r.deleteMethod = "check"
	r.projectID = projectID
	r.checkID = checkID
	return r.deleteErr
}
