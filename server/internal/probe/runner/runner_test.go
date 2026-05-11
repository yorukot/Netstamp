package runner

import (
	"context"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/probe/executor"
)

func TestRunnerExecutesAssignments(t *testing.T) {
	registry := executor.NewRegistry()
	registry.Register(domaincheck.TypePing, fakeExecutor{})

	results, err := New(registry, 2).Run(context.Background(), []domaincheck.Assignment{{
		ID:      "assignment-1",
		CheckID: "check-1",
		Type:    domaincheck.TypePing,
	}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	if results[0].AssignmentID != "assignment-1" || results[0].Ping.Status != domainping.StatusSuccessful {
		t.Fatalf("unexpected result: %#v", results[0])
	}
}

func TestRunnerReturnsErrorResultForUnsupportedType(t *testing.T) {
	results, err := New(executor.NewRegistry(), 1).Run(context.Background(), []domaincheck.Assignment{{
		ID:   "assignment-1",
		Type: domaincheck.Type("dns"),
	}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	if results[0].Ping.ErrorCode == nil || *results[0].Ping.ErrorCode != "unsupported_check_type" {
		t.Fatalf("unexpected unsupported result: %#v", results[0])
	}
}

type fakeExecutor struct{}

func (fakeExecutor) Execute(_ context.Context, assignment domaincheck.Assignment) (domainprobe.Result, error) {
	now := time.Now().UTC()
	return domainprobe.Result{
		AssignmentID: assignment.ID,
		CheckID:      assignment.CheckID,
		Type:         assignment.Type,
		Ping: domainping.Result{
			StartedAt:     now,
			FinishedAt:    now,
			Status:        domainping.StatusSuccessful,
			SentCount:     1,
			ReceivedCount: 1,
		},
	}, nil
}
