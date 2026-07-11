package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type ResultEnvelope struct {
	CheckID    string
	Type       domaincheck.Type
	Ping       domainping.Result
	TCP        domaintcp.Result
	Traceroute domaintraceroute.Result
	HTTP       domainhttp.Result
}

type PingExecutor interface {
	Execute(ctx context.Context, req scheduling.RunRequest) ResultEnvelope
}

type TCPExecutor interface {
	Execute(ctx context.Context, req scheduling.RunRequest) ResultEnvelope
}

type TracerouteExecutor interface {
	Execute(ctx context.Context, req scheduling.RunRequest) ResultEnvelope
}

type HTTPExecutor interface {
	Execute(ctx context.Context, req scheduling.RunRequest) ResultEnvelope
}

type WorkerPool struct {
	maxWorkers int
	queue      <-chan scheduling.RunRequest
	results    *ResultQueue
	ping       PingExecutor
	tcp        TCPExecutor
	traceroute TracerouteExecutor
	http       HTTPExecutor
	log        *slog.Logger
	metrics    WorkerMetrics
}

type WorkerMetrics interface {
	IncActiveWorker()
	DecActiveWorker()
	ObserveExecutorDuration(checkType string, duration time.Duration)
}

func NewWorkerPool(maxWorkers int, queue <-chan scheduling.RunRequest, results *ResultQueue, ping PingExecutor, tcp TCPExecutor, traceroute TracerouteExecutor, httpExecutor HTTPExecutor, log *slog.Logger, metrics WorkerMetrics) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		queue:      queue,
		results:    results,
		ping:       ping,
		tcp:        tcp,
		traceroute: traceroute,
		http:       httpExecutor,
		log:        log,
		metrics:    metrics,
	}
}

func (p *WorkerPool) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(p.maxWorkers)
	for i := range p.maxWorkers {
		workerID := i + 1
		go func() {
			defer wg.Done()
			p.runWorker(ctx, workerID)
		}()
	}
	wg.Wait()

	return ctx.Err()
}

func (p *WorkerPool) runWorker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-p.queue:
			if !ok {
				return
			}
			p.runOne(ctx, workerID, req)
		}
	}
}

func (p *WorkerPool) runOne(ctx context.Context, workerID int, req scheduling.RunRequest) {
	switch req.Check.Type {
	case domaincheck.TypePing:
		if req.Check.PingConfig == nil {
			p.log.Warn("skipped ping occurrence without config", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		if p.ping == nil {
			p.log.Warn("skipped ping occurrence without executor", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		done := p.startExecutor(domaincheck.TypePing)
		defer done()
		result := p.ping.Execute(ctx, req)
		p.results.Enqueue(result)
	case domaincheck.TypeTCP:
		if req.Check.TCPConfig == nil {
			p.log.Warn("skipped tcp occurrence without config", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		if p.tcp == nil {
			p.log.Warn("skipped tcp occurrence without executor", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		done := p.startExecutor(domaincheck.TypeTCP)
		defer done()
		result := p.tcp.Execute(ctx, req)
		p.results.Enqueue(result)
	case domaincheck.TypeTraceroute:
		if req.Check.TracerouteConfig == nil {
			p.log.Warn("skipped traceroute occurrence without config", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		if p.traceroute == nil {
			p.log.Warn("skipped traceroute occurrence without executor", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		done := p.startExecutor(domaincheck.TypeTraceroute)
		defer done()
		result := p.traceroute.Execute(ctx, req)
		p.results.Enqueue(result)
	case domaincheck.TypeHTTP:
		if req.Check.HTTPConfig == nil || p.http == nil {
			p.log.Warn("skipped http occurrence without config or executor", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID)
			return
		}
		done := p.startExecutor(domaincheck.TypeHTTP)
		defer done()
		p.results.Enqueue(p.http.Execute(ctx, req))
	default:
		p.log.Warn("skipped unsupported check type", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID, "check_type", req.Check.Type)
	}
}

func (p *WorkerPool) startExecutor(checkType domaincheck.Type) func() {
	if p.metrics == nil {
		return func() {}
	}

	p.metrics.IncActiveWorker()
	startedAt := time.Now()
	return func() {
		p.metrics.DecActiveWorker()
		p.metrics.ObserveExecutorDuration(string(checkType), time.Since(startedAt))
	}
}
