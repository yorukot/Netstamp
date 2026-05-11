package scheduler

import (
	"context"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func TestSchedulerRunsNewAssignmentsImmediatelyThenWaitsForInterval(t *testing.T) {
	scheduler := New()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	assignment := domaincheck.Assignment{ID: "assignment-1", IntervalSeconds: 30}

	scheduler.Replace([]domaincheck.Assignment{assignment}, now)
	due, err := scheduler.Due(context.Background(), now)
	if err != nil {
		t.Fatalf("due: %v", err)
	}
	if len(due) != 1 || due[0].ID != assignment.ID {
		t.Fatalf("expected assignment due immediately, got %#v", due)
	}

	due, err = scheduler.Due(context.Background(), now)
	if err != nil {
		t.Fatalf("due while running: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected running assignment not to duplicate, got %#v", due)
	}

	scheduler.Complete([]domaincheck.Assignment{assignment}, now)
	due, err = scheduler.Due(context.Background(), now.Add(29*time.Second))
	if err != nil {
		t.Fatalf("due before interval: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected assignment to wait for interval, got %#v", due)
	}

	due, err = scheduler.Due(context.Background(), now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("due after interval: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("expected assignment due after interval, got %#v", due)
	}
}

func TestSchedulerReplaceRemovesMissingAssignments(t *testing.T) {
	scheduler := New()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	scheduler.Replace([]domaincheck.Assignment{{ID: "assignment-1"}}, now)
	scheduler.Replace(nil, now)

	due, err := scheduler.Due(context.Background(), now)
	if err != nil {
		t.Fatalf("due: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected no assignments after replace, got %#v", due)
	}
}
