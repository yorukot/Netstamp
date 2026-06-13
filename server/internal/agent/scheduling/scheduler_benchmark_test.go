package scheduling

import (
	"io"
	"log/slog"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func BenchmarkSchedulerDispatch(b *testing.B) {
	queue := make(chan RunRequest, 1)
	scheduler := NewScheduler(nil, queue, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	task := TaskState{
		AssignmentID: "assignment-1",
		Check: domaincheck.Check{
			ID:              "check-1",
			Type:            domaincheck.TypePing,
			IntervalSeconds: 30,
		},
	}

	b.ReportAllocs()
	for b.Loop() {
		scheduler.dispatchOrSkip(task, now, now)
		<-queue
	}
}
