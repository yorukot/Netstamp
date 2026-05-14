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
	switch req.CheckType {
	case domaincheck.TypePing:
		if req.PingConfig == nil {
			p.log.Warn("skipped ping occurrence without config", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.CheckID)
			return
		}
		result := p.ping.Execute(ctx, req)
		p.logPingResult(ctx, workerID, req, result.Ping)
		p.results.Enqueue(result)
	default:
		p.log.Warn("skipped unsupported check type", "worker_id", workerID, "assignment_id", req.AssignmentID, "check_id", req.CheckID, "check_type", req.CheckType)
	}
}

func (p *WorkerPool) logPingResult(ctx context.Context, workerID int, req scheduling.RunRequest, result domainping.Result) {
	attrs := []slog.Attr{
		slog.Int("worker_id", workerID),
		slog.String("assignment_id", req.AssignmentID),
		slog.String("check_id", req.CheckID),
		slog.String("check_type", string(req.CheckType)),
		slog.Time("scheduled_at", req.ScheduledAt),
		slog.Time("started_at", result.StartedAt),
		slog.Time("finished_at", result.FinishedAt),
		slog.Int("duration_ms", int(result.DurationMs)),
		slog.String("status", string(result.Status)),
		slog.Int("sent_count", int(result.SentCount)),
		slog.Int("received_count", int(result.ReceivedCount)),
		slog.Float64("loss_percent", result.LossPercent),
	}
	if result.RttMinMs != nil {
		attrs = append(attrs, slog.Float64("rtt_min_ms", *result.RttMinMs))
	}
	if result.RttAvgMs != nil {
		attrs = append(attrs, slog.Float64("rtt_avg_ms", *result.RttAvgMs))
	}
	if result.RttMedianMs != nil {
		attrs = append(attrs, slog.Float64("rtt_median_ms", *result.RttMedianMs))
	}
	if result.RttMaxMs != nil {
		attrs = append(attrs, slog.Float64("rtt_max_ms", *result.RttMaxMs))
	}
	if result.RttStddevMs != nil {
		attrs = append(attrs, slog.Float64("rtt_stddev_ms", *result.RttStddevMs))
	}
	if result.ErrorCode != nil {
		attrs = append(attrs, slog.String("error_code", *result.ErrorCode))
	}

	p.log.LogAttrs(ctx, slog.LevelDebug, "ping occurrence completed", attrs...)
}
