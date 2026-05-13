package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/agent/observability"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestWorkerSkipsUnsupportedCheckType(t *testing.T) {
	t.Parallel()

	counters := &observability.RuntimeCounters{}
	queue := NewResultQueue(1, counters, slog.New(slog.NewTextHandler(io.Discard, nil)))
	pool := NewWorkerPool(1, nil, queue, fakePingExecutor{}, counters, slog.New(slog.NewTextHandler(io.Discard, nil)))

	pool.runOne(t.Context(), 1, scheduling.RunRequest{
		CheckID:     "check-a",
		CheckType:   domaincheck.Type("http"),
		ScheduledAt: time.Now().UTC(),
	})

	if got := counters.CompletedRuns.Load(); got != 0 {
		t.Fatalf("expected unsupported checks not to increment completed runs, got %d", got)
	}
}

func TestWorkerDebugLogsPingResult(t *testing.T) {
	t.Parallel()

	var logBuffer bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	counters := &observability.RuntimeCounters{}
	queue := NewResultQueue(1, counters, slog.New(slog.NewTextHandler(io.Discard, nil)))
	result := domainping.Result{
		StartedAt:     time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC),
		FinishedAt:    time.Date(2026, 5, 13, 12, 0, 1, 0, time.UTC),
		DurationMs:    1000,
		Status:        domainping.StatusSuccessful,
		SentCount:     4,
		ReceivedCount: 4,
		LossPercent:   0,
	}
	pool := NewWorkerPool(1, nil, queue, fakePingExecutor{result: result}, counters, log)

	pool.runOne(t.Context(), 7, scheduling.RunRequest{
		AssignmentID: "assignment-a",
		CheckID:      "check-a",
		CheckType:    domaincheck.TypePing,
		ScheduledAt:  time.Date(2026, 5, 13, 11, 59, 0, 0, time.UTC),
		PingConfig:   &domainping.Config{},
	})

	var record map[string]any
	if err := json.Unmarshal(logBuffer.Bytes(), &record); err != nil {
		t.Fatalf("decode debug log: %v", err)
	}
	if got := record["msg"]; got != "ping occurrence completed" {
		t.Fatalf("expected ping completion debug message, got %v", got)
	}
	if got := record["level"]; got != "DEBUG" {
		t.Fatalf("expected debug level, got %v", got)
	}
	if got := record["assignment_id"]; got != "assignment-a" {
		t.Fatalf("expected assignment id in log, got %v", got)
	}
	if got := record["check_id"]; got != "check-a" {
		t.Fatalf("expected check id in log, got %v", got)
	}
	if got := record["status"]; got != string(domainping.StatusSuccessful) {
		t.Fatalf("expected ping status in log, got %v", got)
	}
	if got := counters.CompletedRuns.Load(); got != 1 {
		t.Fatalf("expected ping run to increment completed runs, got %d", got)
	}
}

type fakePingExecutor struct {
	result domainping.Result
}

func (e fakePingExecutor) Execute(_ context.Context, req scheduling.RunRequest) ResultEnvelope {
	return ResultEnvelope{
		CheckID: req.CheckID,
		Type:    domaincheck.TypePing,
		Ping:    e.result,
	}
}
