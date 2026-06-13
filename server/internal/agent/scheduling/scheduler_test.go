package scheduling

import (
	"io"
	"log/slog"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func TestSchedulerRecordsSkippedFullWorkerQueue(t *testing.T) {
	metrics := &fakeSchedulerMetrics{skipped: make(map[string]int)}
	queue := make(chan RunRequest)
	scheduler := NewScheduler(nil, queue, slog.New(slog.NewTextHandler(io.Discard, nil)), metrics)
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	scheduler.dispatchOrSkip(TaskState{
		AssignmentID: "assignment-1",
		Check: domaincheck.Check{
			ID:              "check-1",
			Type:            domaincheck.TypePing,
			IntervalSeconds: 30,
		},
	}, now, now)

	if metrics.scheduled != 1 {
		t.Fatalf("expected one scheduled run, got %d", metrics.scheduled)
	}
	if metrics.skipped["worker_queue_full"] != 1 {
		t.Fatalf("expected worker queue skip to be recorded once, got %#v", metrics.skipped)
	}
}

type fakeSchedulerMetrics struct {
	scheduled int
	skipped   map[string]int
}

func (m *fakeSchedulerMetrics) IncScheduledRun() {
	m.scheduled++
}

func (m *fakeSchedulerMetrics) IncSkippedRun(reason string) {
	m.skipped[reason]++
}
