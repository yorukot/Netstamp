package worker

import (
	"io"
	"log/slog"
	"testing"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func BenchmarkResultQueueBackpressure(b *testing.B) {
	queue := NewResultQueue(1, slog.New(slog.NewTextHandler(io.Discard, nil)))
	result := ResultEnvelope{
		CheckID: "check-1",
		Type:    domaincheck.TypePing,
	}

	b.ReportAllocs()
	for b.Loop() {
		queue.Enqueue(result)
	}
}
