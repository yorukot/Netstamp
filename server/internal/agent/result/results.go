package result

import (
	"context"
	"errors"
	"log/slog"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
	"github.com/yorukot/netstamp/internal/agent/retry"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type Submitter struct {
	client       httpclient.RuntimeClient
	queue        *agentworker.ResultQueue
	config       agentconfig.Config
	flushEvery   time.Duration
	batchSize    int
	shutdownWait time.Duration
	log          *slog.Logger
}

func New(client httpclient.RuntimeClient, queue *agentworker.ResultQueue, localConfig agentconfig.Config, log *slog.Logger) *Submitter {
	return &Submitter{
		client:       client,
		queue:        queue,
		config:       localConfig,
		flushEvery:   localConfig.ResultFlushInterval.Value,
		batchSize:    localConfig.ResultBatchSize.Value,
		shutdownWait: localConfig.ShutdownTimeout.Value,
		log:          log,
	}
}

func (s *Submitter) Run(ctx context.Context) error {
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
					if errors.Is(err, httpclient.ErrAuthFailed) {
						return err
					}
					s.log.Warn("result batch submit failed", "error", err, "batch_size", len(batch))
				}
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}
			if err := s.flush(ctx, batch); err != nil {
				if errors.Is(err, httpclient.ErrAuthFailed) {
					return err
				}
				s.log.Warn("result batch submit failed", "error", err, "batch_size", len(batch))
			}
			batch = batch[:0]
		}
	}
}

func (s *Submitter) flushBestEffort(batch []agentworker.ResultEnvelope) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownWait)
	defer cancel()
	if err := s.flush(ctx, batch); err != nil {
		s.log.Warn("best-effort result flush failed", "error", err, "batch_size", len(batch))
	}
}

func (s *Submitter) flush(ctx context.Context, batch []agentworker.ResultEnvelope) error {
	if len(batch) == 0 {
		return nil
	}

	input := httpclient.SubmitResultsInput{Results: groupResults(batch)}
	var lastErr error
	backoff := s.config.InitialBackoff.Value

	for attempt := 1; attempt <= s.config.MaxAttempts.Value; attempt++ {
		_, err := s.client.SubmitResults(ctx, input)
		if err == nil {
			return nil
		}
		if errors.Is(err, httpclient.ErrAuthFailed) || errors.Is(err, httpclient.ErrPermanentRuntimeAPI) {
			return err
		}
		lastErr = err
		if attempt == s.config.MaxAttempts.Value {
			break
		}

		s.log.Debug("retrying result submission", "attempt", attempt, "backoff", backoff, "error", err)
		if err := retry.WaitForDuration(ctx, backoff); err != nil {
			return err
		}
		backoff *= 2
		if backoff > s.config.MaxBackoff.Value {
			backoff = s.config.MaxBackoff.Value
		}
	}

	return lastErr
}

func groupResults(batch []agentworker.ResultEnvelope) []httpclient.RuntimeResultGroup {
	orderedKeys := make([]resultGroupKey, 0)
	groups := make(map[resultGroupKey][]httpclient.PingResultBody)

	for _, result := range batch {
		key := resultGroupKey{checkID: result.CheckID, checkType: result.Type}
		if _, ok := groups[key]; !ok {
			orderedKeys = append(orderedKeys, key)
		}
		groups[key] = append(groups[key], pingResultBody(result.Ping))
	}

	output := make([]httpclient.RuntimeResultGroup, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		output = append(output, httpclient.RuntimeResultGroup{
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

func pingResultBody(result domainping.Result) httpclient.PingResultBody {
	return httpclient.PingResultBody{
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
