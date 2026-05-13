package runtime

import (
	"context"
	"errors"
	"log/slog"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/observability"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type ResultSubmitter struct {
	client       RuntimeClient
	queue        *agentworker.ResultQueue
	config       *RuntimeConfigStore
	flushEvery   time.Duration
	batchSize    int
	shutdownWait time.Duration
	counters     *observability.RuntimeCounters
	log          *slog.Logger
}

func NewResultSubmitter(client RuntimeClient, queue *agentworker.ResultQueue, config *RuntimeConfigStore, localConfig agentconfig.Config, counters *observability.RuntimeCounters, log *slog.Logger) *ResultSubmitter {
	return &ResultSubmitter{
		client:       client,
		queue:        queue,
		config:       config,
		flushEvery:   localConfig.ResultFlushInterval,
		batchSize:    localConfig.ResultBatchSize,
		shutdownWait: localConfig.ShutdownTimeout,
		counters:     counters,
		log:          log,
	}
}

func (s *ResultSubmitter) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.flushEvery)
	defer ticker.Stop()

	batch := make([]agentworker.ResultEnvelope, 0, s.batchSize)
	for {
		select {
		case <-ctx.Done():
			batch = append(batch, s.queue.Drain(s.batchSize-len(batch))...)
			s.flushBestEffort(batch)
			return ctx.Err()
		case result := <-s.queue.Channel():
			batch = append(batch, result)
			if len(batch) >= s.batchSize {
				if err := s.flush(ctx, batch); err != nil {
					if errors.Is(err, ErrAuthFailed) {
						return err
					}
					s.counters.ResultSubmitErrors.Add(1)
					s.log.Warn("result batch submit failed", "error", err, "batch_size", len(batch))
				}
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}
			if err := s.flush(ctx, batch); err != nil {
				if errors.Is(err, ErrAuthFailed) {
					return err
				}
				s.counters.ResultSubmitErrors.Add(1)
				s.log.Warn("result batch submit failed", "error", err, "batch_size", len(batch))
			}
			batch = batch[:0]
		}
	}
}

func (s *ResultSubmitter) flushBestEffort(batch []agentworker.ResultEnvelope) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownWait)
	defer cancel()
	if err := s.flush(ctx, batch); err != nil {
		s.log.Warn("best-effort result flush failed", "error", err, "batch_size", len(batch))
	}
}

func (s *ResultSubmitter) flush(ctx context.Context, batch []agentworker.ResultEnvelope) error {
	if len(batch) == 0 {
		return nil
	}

	input := SubmitResultsInput{Results: groupResults(batch)}
	runtimeConfig := s.config.Get()
	var lastErr error
	backoff := runtimeConfig.InitialBackoff

	for attempt := 1; attempt <= runtimeConfig.MaxAttempts; attempt++ {
		_, err := s.client.SubmitResults(ctx, input)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrAuthFailed) || errors.Is(err, ErrPermanentRuntimeAPI) {
			return err
		}
		lastErr = err
		if attempt == runtimeConfig.MaxAttempts {
			break
		}

		s.log.Debug("retrying result submission", "attempt", attempt, "backoff", backoff, "error", err)
		if err := sleepContext(ctx, backoff); err != nil {
			return err
		}
		backoff *= 2
		if backoff > runtimeConfig.MaxBackoff {
			backoff = runtimeConfig.MaxBackoff
		}
	}

	return lastErr
}

func groupResults(batch []agentworker.ResultEnvelope) []RuntimeResultGroup {
	orderedKeys := make([]resultGroupKey, 0)
	groups := make(map[resultGroupKey][]PingResultBody)

	for _, result := range batch {
		key := resultGroupKey{checkID: result.CheckID, checkType: result.Type}
		if _, ok := groups[key]; !ok {
			orderedKeys = append(orderedKeys, key)
		}
		groups[key] = append(groups[key], pingResultBody(result.Ping))
	}

	output := make([]RuntimeResultGroup, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		output = append(output, RuntimeResultGroup{
			CheckID: key.checkID,
			Type:    key.checkType,
			Ping:    groups[key],
		})
	}

	return output
}

type resultGroupKey struct {
	checkID   string
	checkType domaincheck.Type
}

func pingResultBody(result domainping.Result) PingResultBody {
	return PingResultBody{
		StartedAt:     result.StartedAt,
		FinishedAt:    result.FinishedAt,
		DurationMs:    result.DurationMs,
		Status:        result.Status,
		SentCount:     result.SentCount,
		ReceivedCount: result.ReceivedCount,
		LossPercent:   result.LossPercent,
		RttMinMs:      result.RttMinMs,
		RttAvgMs:      result.RttAvgMs,
		RttMedianMs:   result.RttMedianMs,
		RttMaxMs:      result.RttMaxMs,
		RttStddevMs:   result.RttStddevMs,
		RttSamplesMs:  append([]float64(nil), result.RttSamplesMs...),
		ResolvedIP:    result.ResolvedIP,
		IPFamily:      result.IPFamily,
		Raw:           result.Raw,
		ErrorCode:     result.ErrorCode,
		ErrorMessage:  result.ErrorMessage,
	}
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
