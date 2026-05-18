package scheduling

import (
	"io"
	"log/slog"
	"testing"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func TestReconcileAcceptsAssignmentWithoutEmbeddedProbe(t *testing.T) {
	store := NewAssignmentStore("probe-1", time.Minute, discardLogger())
	pulledAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)

	summary := store.Reconcile([]domainassignment.Assignment{{
		ID:              "assignment-1",
		CheckVersion:    "check-version-1",
		SelectorVersion: "selector-version-1",
		Check: &domaincheck.Check{
			ID:              "check-1",
			Type:            domaincheck.TypePing,
			Target:          "127.0.0.1",
			IntervalSeconds: 30,
			PingConfig:      pingConfig(),
		},
	}}, pulledAt)

	if summary.Added != 1 || summary.Active != 1 || summary.Unsupported != 0 {
		t.Fatalf("unexpected summary: %#v", summary)
	}

	tasks := store.ActiveTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected one active task, got %d", len(tasks))
	}
	if tasks[0].Check.ID != "check-1" {
		t.Fatalf("expected check id check-1, got %q", tasks[0].Check.ID)
	}

	request := tasks[0].RunRequest(pulledAt.Add(time.Second), pulledAt.Add(2*time.Second))
	if request.Check.ID != "check-1" {
		t.Fatalf("expected request check id check-1, got %q", request.Check.ID)
	}
}

func TestReconcileSkipsAssignmentWithoutCheck(t *testing.T) {
	store := NewAssignmentStore("probe-1", time.Minute, discardLogger())

	summary := store.Reconcile([]domainassignment.Assignment{{
		ID: "assignment-1",
	}}, time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC))

	if summary.Unsupported != 1 || summary.Active != 0 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if tasks := store.ActiveTasks(); len(tasks) != 0 {
		t.Fatalf("expected no active tasks, got %d", len(tasks))
	}
}

func TestReconcileSkipsTracerouteUntilExecutorExists(t *testing.T) {
	store := NewAssignmentStore("probe-1", time.Minute, discardLogger())
	config := domaintraceroute.DefaultConfig()

	summary := store.Reconcile([]domainassignment.Assignment{{
		ID: "assignment-1",
		Check: &domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			Target:           "example.com",
			IntervalSeconds:  30,
			TracerouteConfig: &config,
		},
	}}, time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC))

	if summary.Unsupported != 1 || summary.Active != 0 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if tasks := store.ActiveTasks(); len(tasks) != 0 {
		t.Fatalf("expected no active tasks, got %d", len(tasks))
	}
}

func pingConfig() *domainping.Config {
	config := domainping.DefaultConfig()
	return &config
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
