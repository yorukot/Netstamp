package httpquery

import (
	"testing"
	"time"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestSelectReadPlanUsesRollupForMixedAgeWindow(t *testing.T) {
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)

	plan := SelectReadPlan(120, now.Add(-rawRetentionWindow-time.Millisecond), now, 300)

	want := domainhttp.SeriesReadPlan{
		Mode:        domainhttp.SeriesReadModeRollup,
		Source:      domainhttp.SeriesSourceAggregate,
		Resolution:  domainhttp.SeriesResolutionOneMinute,
		TotalPoints: 120,
	}
	if plan != want {
		t.Fatalf("unexpected HTTP read plan: got %#v want %#v", plan, want)
	}
}
