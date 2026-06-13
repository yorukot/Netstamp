package result

import (
	"testing"
	"time"

	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func BenchmarkGroupResults(b *testing.B) {
	startedAt := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	batch := make([]agentworker.ResultEnvelope, 0, 100)
	for range 100 {
		batch = append(batch, agentworker.ResultEnvelope{
			CheckID: "check-1",
			Type:    domaincheck.TypePing,
			Ping: domainping.Result{
				StartedAt:     startedAt,
				FinishedAt:    startedAt.Add(time.Second),
				DurationMs:    1000,
				Status:        domainping.StatusSuccessful,
				SentCount:     3,
				ReceivedCount: 3,
				LossPercent:   0,
				RttSamplesMs:  []float64{1, 2, 3},
			},
		})
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = groupResults(batch)
	}
}
