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
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type Submitter struct {
	client       httpclient.RuntimeClient
	queue        *agentworker.ResultQueue
	config       agentconfig.Config
	flushEvery   time.Duration
	batchSize    int
	shutdownWait time.Duration
	log          *slog.Logger
	metrics      Metrics
}

type Metrics interface {
	IncSubmitFailure()
	IncSubmitRetry()
	ObserveSubmitBatchSize(size int)
}

func New(client httpclient.RuntimeClient, queue *agentworker.ResultQueue, localConfig agentconfig.Config, log *slog.Logger, metrics Metrics) *Submitter {
	return &Submitter{
		client:       client,
		queue:        queue,
		config:       localConfig,
		flushEvery:   localConfig.ResultFlushInterval.Value,
		batchSize:    localConfig.ResultBatchSize.Value,
		shutdownWait: localConfig.ShutdownTimeout.Value,
		log:          log,
		metrics:      metrics,
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
			s.flushBestEffort(ctx, batch)
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

func (s *Submitter) flushBestEffort(parent context.Context, batch []agentworker.ResultEnvelope) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), s.shutdownWait)
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
			if s.metrics != nil {
				s.metrics.ObserveSubmitBatchSize(len(batch))
			}
			return nil
		}
		if errors.Is(err, httpclient.ErrAuthFailed) || errors.Is(err, httpclient.ErrPermanentRuntimeAPI) {
			if s.metrics != nil {
				s.metrics.IncSubmitFailure()
			}
			return err
		}
		lastErr = err
		if attempt == s.config.MaxAttempts.Value {
			break
		}

		s.log.Debug("retrying result submission", "attempt", attempt, "backoff", backoff, "error", err)
		if s.metrics != nil {
			s.metrics.IncSubmitRetry()
		}
		if err := retry.WaitForDuration(ctx, backoff); err != nil {
			return err
		}
		backoff *= 2
		if backoff > s.config.MaxBackoff.Value {
			backoff = s.config.MaxBackoff.Value
		}
	}

	if s.metrics != nil {
		s.metrics.IncSubmitFailure()
	}
	return lastErr
}

func groupResults(batch []agentworker.ResultEnvelope) []httpclient.RuntimeResultGroup {
	orderedKeys := make([]resultGroupKey, 0)
	groups := make(map[resultGroupKey]runtimeResultGroupValues)

	for _, result := range batch {
		key := resultGroupKey{checkID: result.CheckID, checkType: result.Type}
		if _, ok := groups[key]; !ok {
			orderedKeys = append(orderedKeys, key)
		}
		values := groups[key]
		switch result.Type {
		case domaincheck.TypePing:
			values.ping = append(values.ping, pingResultBody(result.Ping))
		case domaincheck.TypeTCP:
			values.tcp = append(values.tcp, tcpResultBody(result.TCP))
		case domaincheck.TypeTraceroute:
			values.traceroute = append(values.traceroute, tracerouteResultBody(result.Traceroute))
		case domaincheck.TypeHTTP:
			values.http = append(values.http, httpResultBody(result.HTTP))
		}
		groups[key] = values
	}

	output := make([]httpclient.RuntimeResultGroup, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		values := groups[key]
		output = append(output, httpclient.RuntimeResultGroup{
			CheckID:    key.checkID,
			Type:       key.checkType,
			Ping:       values.ping,
			TCP:        values.tcp,
			Traceroute: values.traceroute,
			HTTP:       values.http,
		})
	}

	return output
}

type runtimeResultGroupValues struct {
	ping       []httpclient.PingResultBody
	tcp        []httpclient.TCPResultBody
	traceroute []httpclient.TracerouteResultBody
	http       []httpclient.HTTPResultBody
}

type resultGroupKey struct {
	checkID   string
	checkType domaincheck.Type
}

func pingResultBody(result domainping.Result) httpclient.PingResultBody {
	body := httpclient.PingResultBody(result)
	body.RttSamplesMs = append([]float64(nil), result.RttSamplesMs...)
	return body
}

func tcpResultBody(result domaintcp.Result) httpclient.TCPResultBody {
	return httpclient.TCPResultBody(result)
}

func httpResultBody(result domainhttp.Result) httpclient.HTTPResultBody {
	return httpclient.HTTPResultBody{
		StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, DurationMs: result.DurationMs,
		Status: result.Status, DNSDurationMs: result.DNSDurationMs, ConnectDurationMs: result.ConnectDurationMs,
		TLSDurationMs: result.TLSDurationMs, TTFBDurationMs: result.TTFBDurationMs,
		ResolvedIP: result.ResolvedIP, IPFamily: result.IPFamily, StatusCode: result.StatusCode,
		FinalURL: result.FinalURL, RedirectCount: result.RedirectCount, ResponseBytes: result.ResponseBytes,
		ResponseTruncated: result.ResponseTruncated, BodyMatched: result.BodyMatched,
		TLSVersion: result.TLSVersion, TLSCipherSuite: result.TLSCipherSuite,
		CertificateNotBefore: result.CertificateNotBefore, CertificateNotAfter: result.CertificateNotAfter,
		ErrorCode: result.ErrorCode, ErrorMessage: result.ErrorMessage,
	}
}

func tracerouteResultBody(result domaintraceroute.Result) httpclient.TracerouteResultBody {
	return httpclient.TracerouteResultBody{
		StartedAt:          result.StartedAt,
		FinishedAt:         result.FinishedAt,
		DurationMs:         result.DurationMs,
		Status:             result.Status,
		ResolvedIP:         result.ResolvedIP,
		IPFamily:           result.IPFamily,
		DestinationReached: result.DestinationReached,
		HopCount:           result.HopCount,
		Hops:               tracerouteHopBodies(result.Hops),
		ErrorCode:          result.ErrorCode,
		ErrorMessage:       result.ErrorMessage,
	}
}

func tracerouteHopBodies(hops []domaintraceroute.HopResult) []httpclient.TracerouteHopBody {
	values := make([]httpclient.TracerouteHopBody, len(hops))
	for i, hop := range hops {
		body := httpclient.TracerouteHopBody(hop)
		body.RttSamplesMs = append([]float64(nil), hop.RttSamplesMs...)
		values[i] = body
	}
	return values
}
