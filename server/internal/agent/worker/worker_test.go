package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func TestWorkerRunsTracerouteExecutor(t *testing.T) {
	queue := NewResultQueue(1, discardWorkerLogger())
	config := domaintraceroute.DefaultConfig()
	executor := &fakeTracerouteExecutor{result: ResultEnvelope{
		CheckID: "check-1",
		Type:    domaincheck.TypeTraceroute,
		Traceroute: domaintraceroute.Result{
			Status: domaintraceroute.StatusSuccessful,
		},
	}}
	pool := NewWorkerPool(1, nil, queue, nil, nil, executor, discardWorkerLogger(), nil)

	pool.runOne(context.Background(), 1, scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:               "check-1",
			Type:             domaincheck.TypeTraceroute,
			TracerouteConfig: &config,
		},
		ScheduledAt: time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
	})

	if executor.called != 1 {
		t.Fatalf("expected traceroute executor to be called once, got %d", executor.called)
	}
	results := queue.Drain(1)
	if len(results) != 1 || results[0].Type != domaincheck.TypeTraceroute || results[0].Traceroute.Status != domaintraceroute.StatusSuccessful {
		t.Fatalf("unexpected queued result: %#v", results)
	}
}

func TestWorkerRunsTCPExecutor(t *testing.T) {
	queue := NewResultQueue(1, discardWorkerLogger())
	config := domaintcp.DefaultConfig()
	executor := &fakeTCPExecutor{result: ResultEnvelope{
		CheckID: "check-1",
		Type:    domaincheck.TypeTCP,
		TCP: domaintcp.Result{
			Status: domaintcp.StatusSuccessful,
		},
	}}
	pool := NewWorkerPool(1, nil, queue, nil, executor, nil, discardWorkerLogger(), nil)

	pool.runOne(context.Background(), 1, scheduling.RunRequest{
		Check: domaincheck.Check{
			ID:        "check-1",
			Type:      domaincheck.TypeTCP,
			TCPConfig: &config,
		},
		ScheduledAt: time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
	})

	if executor.called != 1 {
		t.Fatalf("expected tcp executor to be called once, got %d", executor.called)
	}
	results := queue.Drain(1)
	if len(results) != 1 || results[0].Type != domaincheck.TypeTCP || results[0].TCP.Status != domaintcp.StatusSuccessful {
		t.Fatalf("unexpected queued result: %#v", results)
	}
}

func TestResultQueueRecordsDrops(t *testing.T) {
	queue := NewResultQueue(1, discardWorkerLogger())
	metrics := &fakeResultQueueMetrics{}
	queue.SetMetrics(metrics)

	queue.Enqueue(ResultEnvelope{CheckID: "first", Type: domaincheck.TypePing})
	queue.Enqueue(ResultEnvelope{CheckID: "second", Type: domaincheck.TypePing})

	if metrics.dropped["result_queue_full"] != 1 {
		t.Fatalf("expected result_queue_full drop to be recorded once, got %#v", metrics.dropped)
	}
}

type fakeTCPExecutor struct {
	called int
	result ResultEnvelope
}

func (e *fakeTCPExecutor) Execute(context.Context, scheduling.RunRequest) ResultEnvelope {
	e.called++
	return e.result
}

type fakeTracerouteExecutor struct {
	called int
	result ResultEnvelope
}

func (e *fakeTracerouteExecutor) Execute(context.Context, scheduling.RunRequest) ResultEnvelope {
	e.called++
	return e.result
}

type fakeResultQueueMetrics struct {
	dropped map[string]int
}

func (m *fakeResultQueueMetrics) IncDroppedResult(reason string) {
	if m.dropped == nil {
		m.dropped = make(map[string]int)
	}
	m.dropped[reason]++
}

func discardWorkerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
