package runtime

import (
	"testing"
	"time"

	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestGroupResultsGroupsByCheckAndType(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	batch := []agentworker.ResultEnvelope{
		{CheckID: "check-a", Type: domaincheck.TypePing, Ping: domainping.Result{StartedAt: now, FinishedAt: now, Status: domainping.StatusSuccessful}},
		{CheckID: "check-b", Type: domaincheck.TypePing, Ping: domainping.Result{StartedAt: now, FinishedAt: now, Status: domainping.StatusSuccessful}},
		{CheckID: "check-a", Type: domaincheck.TypePing, Ping: domainping.Result{StartedAt: now.Add(time.Second), FinishedAt: now.Add(time.Second), Status: domainping.StatusSuccessful}},
	}

	groups := groupResults(batch)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].CheckID != "check-a" || len(groups[0].Ping) != 2 {
		t.Fatalf("expected first group to contain two check-a ping results, got %#v", groups[0])
	}
	if groups[1].CheckID != "check-b" || len(groups[1].Ping) != 1 {
		t.Fatalf("expected second group to contain one check-b ping result, got %#v", groups[1])
	}
}
