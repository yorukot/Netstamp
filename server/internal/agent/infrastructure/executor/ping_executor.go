package executor

import (
	"context"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type ProBingExecutor struct{}

func NewProBingExecutor() *ProBingExecutor {
	return &ProBingExecutor{}
}

func (e *ProBingExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	result := e.execute(ctx, req)
	return agentworker.ResultEnvelope{
		CheckID: req.Check.ID,
		Type:    domaincheck.TypePing,
		Ping:    result,
	}
}

func (e *ProBingExecutor) execute(ctx context.Context, req scheduling.RunRequest) domainping.Result {
	startedAt := req.ScheduledAt.UTC()
	finishedAt := time.Now().UTC()
	if req.Check.PingConfig == nil {
		return errorPingResult(startedAt, finishedAt, "missing_ping_config", "ping config is missing")
	}
}

func errorPingResult(startedAt, finishedAt time.Time, errorCode, errorMessage string) domainping.Result {
	return domainping.Result{
		StartedAt:     startedAt.UTC(),
		FinishedAt:    finishedAt.UTC(),
		DurationMs:    durationMillis(startedAt, finishedAt),
		Status:        domainping.StatusError,
		SentCount:     0,
		ReceivedCount: 0,
		LossPercent:   100,
		RttMinMs:      nil,
		RttAvgMs:      nil,
		RttMedianMs:   nil,
		RttMaxMs:      nil,
		RttStddevMs:   nil,
		RttSamplesMs:  nil,
		ResolvedIP:    nil,
		IPFamily:      nil,
		Raw:           nil,
		ErrorCode:     optionalString(errorCode),
		ErrorMessage:  optionalString(errorMessage),
	}
}

func durationMillis(startedAt, finishedAt time.Time) *int64 {
	
}