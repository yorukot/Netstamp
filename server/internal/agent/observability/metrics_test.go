package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsHandlerExposesAgentMetricsWithoutSensitiveLabels(t *testing.T) {
	metrics := NewMetrics(MetricsOptions{
		WorkerQueueDepth:    func() float64 { return 2 },
		WorkerQueueCapacity: func() float64 { return 8 },
		ResultQueueDepth:    func() float64 { return 3 },
		ResultQueueCapacity: func() float64 { return 10 },
	})
	metrics.IncScheduledRun()
	metrics.IncSkippedRun("worker_queue_full")
	metrics.IncDroppedResult("result_queue_full")
	metrics.IncSubmitRetry()
	metrics.IncSubmitFailure()
	metrics.ObserveSubmitBatchSize(42)
	metrics.IncActiveWorker()
	metrics.ObserveExecutorDuration("ping", 150*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rr := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	expected := []string{
		"netstamp_probe_worker_queue_depth 2",
		"netstamp_probe_worker_queue_capacity 8",
		"netstamp_probe_result_queue_depth 3",
		"netstamp_probe_result_queue_capacity 10",
		"netstamp_probe_scheduled_runs_total 1",
		`netstamp_probe_skipped_runs_total{reason="worker_queue_full"} 1`,
		`netstamp_probe_dropped_results_total{reason="result_queue_full"} 1`,
		"netstamp_probe_result_submit_retries_total 1",
		"netstamp_probe_result_submit_failures_total 1",
		`netstamp_probe_executor_duration_seconds_count{check_type="ping"} 1`,
	}
	for _, value := range expected {
		if !strings.Contains(body, value) {
			t.Fatalf("expected metrics output to contain %q, got:\n%s", value, body)
		}
	}

	disallowed := []string{"probe_secret", "check_target", "error_message", "resolved_ip"}
	for _, value := range disallowed {
		if strings.Contains(body, value) {
			t.Fatalf("metrics output contains sensitive label %q:\n%s", value, body)
		}
	}
}
