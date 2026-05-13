package scheduling

import (
	"testing"
	"time"
)

func TestComputePhaseIsDeterministicAndBounded(t *testing.T) {
	t.Parallel()

	interval := 30 * time.Second
	first := ComputePhase("probe-a", "assignment-a", interval)
	second := ComputePhase("probe-a", "assignment-a", interval)
	if first != second {
		t.Fatalf("expected deterministic phase, got %s and %s", first, second)
	}
	if first < 0 || first >= interval {
		t.Fatalf("expected phase within interval, got %s for %s", first, interval)
	}
	if got := ComputePhase("probe-a", "assignment-a", time.Second); got != 0 {
		t.Fatalf("expected one second interval to have zero phase, got %s", got)
	}
}

func TestComputeNextFutureDueDoesNotBackfill(t *testing.T) {
	t.Parallel()

	previous := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	now := previous.Add(10 * time.Second)
	got := ComputeNextFutureDue(previous, now, time.Second)
	want := now.Add(time.Second)
	if !got.Equal(want) {
		t.Fatalf("expected next future due %s, got %s", want, got)
	}
}
