package publicstatus

import (
	"testing"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func TestAssignmentMetricsBodyMapsHTTPTotalLatency(t *testing.T) {
	totalMs := 125.5
	failurePercent := 100.0
	metrics := assignmentMetricsBody(domainpublic.Assignment{
		CheckType:      domaincheck.TypeHTTP,
		LatestStatus:   "error",
		LatencyAvgMs:   &totalMs,
		FailurePercent: &failurePercent,
	})

	if metrics == nil || metrics.LatencyAvgMs == nil || *metrics.LatencyAvgMs != totalMs {
		t.Fatalf("expected HTTP total latency, got %#v", metrics)
	}
	if metrics.FailurePercent == nil || *metrics.FailurePercent != failurePercent {
		t.Fatalf("expected HTTP failure percent, got %#v", metrics)
	}
	if metrics.ConnectAvgMs != nil || metrics.LossPercent != nil {
		t.Fatalf("unexpected TCP or ping metrics in HTTP response: %#v", metrics)
	}
}
