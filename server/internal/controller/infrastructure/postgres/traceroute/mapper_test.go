package pgtraceroute

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
)

func TestMapRunRowsSkipsNoHopSentinel(t *testing.T) {
	startedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	result := mapRunRows([]sqlc.ListTracerouteRunRowsRow{{
		StartedAt:  startedAt,
		FinishedAt: startedAt.Add(time.Second),
		Status:     sqlc.TracerouteStatusSuccessful,
		HopIndex:   0,
	}}, 10)

	if len(result.Runs) != 1 {
		t.Fatalf("expected one run, got %#v", result.Runs)
	}
	if len(result.Runs[0].Hops) != 0 {
		t.Fatalf("expected sentinel row to be skipped, got %#v", result.Runs[0].Hops)
	}
}

func TestMapTopologyRowsSkipsNoHopSentinel(t *testing.T) {
	startedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	result := mapTopologyRows([]sqlc.ListTracerouteTopologyRowsRow{{
		StartedAt:     startedAt,
		ProbePublicID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		ProbeName:     "edge-1",
		CheckPublicID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		CheckName:     "trace example",
		CheckTarget:   "example.net",
		HopIndex:      0,
	}})

	if len(result.Runs) != 1 {
		t.Fatalf("expected one topology run, got %#v", result.Runs)
	}
	if len(result.Runs[0].Hops) != 0 {
		t.Fatalf("expected sentinel row to be skipped, got %#v", result.Runs[0].Hops)
	}
}
