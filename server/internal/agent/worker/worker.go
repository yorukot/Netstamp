package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type ResultEnvelope struct {
	CheckID string
	Type    domaincheck.Type
	Ping    domainping.Result
}

type PingExecutor interface {
	Execute(ctx context.Context, req scheduling.RunRequest) ResultEnvelope
}

type WorkerPool struct {
	maxWorkers int
	queue      <-chan scheduling.RunRequest
	results    *ResultQueue
	ping       PingExecutor
	log        *slog.Logger
}

func NewWorkerPool(maxWorkers int, queue <-chan scheduling.RunRequest, results *ResultQueue, ping PingExecutor, log *slog.Logger) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		queue:      queue,
		results:    results,
		ping:       ping,
		log:        log,
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
		result := p.ping.Execute(ctx, req)
		p.results.Enqueue(result)
	default:
		p.log.Warn("skipped unsupported check type", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.Check.ID, "check_type", req.Check.Type)
	}
}
