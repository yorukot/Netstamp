package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type QueueValueFunc func() float64

type MetricsOptions struct {
	WorkerQueueDepth    QueueValueFunc
	WorkerQueueCapacity QueueValueFunc
	ResultQueueDepth    QueueValueFunc
	ResultQueueCapacity QueueValueFunc
}

type Metrics struct {
	registry *prometheus.Registry
	handler  http.Handler

	activeWorkers       prometheus.Gauge
	scheduledRuns       prometheus.Counter
	skippedRuns         *prometheus.CounterVec
	droppedResults      *prometheus.CounterVec
	submitFailures      prometheus.Counter
	submitRetries       prometheus.Counter
	submitBatchSize     prometheus.Histogram
	executorDurationSec *prometheus.HistogramVec
}

func NewMetrics(options MetricsOptions) *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := &Metrics{
		registry: registry,
		handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		}),
		activeWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "netstamp_probe_worker_active",
			Help: "Number of probe agent workers currently executing checks.",
		}),
		scheduledRuns: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "netstamp_probe_scheduled_runs_total",
			Help: "Total probe check run occurrences considered by the scheduler.",
		}),
		skippedRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "netstamp_probe_skipped_runs_total",
			Help: "Total probe check run occurrences skipped by the scheduler.",
		}, []string{"reason"}),
		droppedResults: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "netstamp_probe_dropped_results_total",
			Help: "Total probe check results dropped before submission.",
		}, []string{"reason"}),
		submitFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "netstamp_probe_result_submit_failures_total",
			Help: "Total result batch submission failures after retry handling.",
		}),
		submitRetries: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "netstamp_probe_result_submit_retries_total",
			Help: "Total result batch submission retries.",
		}),
		submitBatchSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "netstamp_probe_result_submit_batch_size",
			Help:    "Number of result envelopes in successfully submitted batches.",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		}),
		executorDurationSec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "netstamp_probe_executor_duration_seconds",
			Help:    "Probe check executor duration by check type.",
			Buckets: prometheus.DefBuckets,
		}, []string{"check_type"}),
	}

	registry.MustRegister(
		m.activeWorkers,
		m.scheduledRuns,
		m.skippedRuns,
		m.droppedResults,
		m.submitFailures,
		m.submitRetries,
		m.submitBatchSize,
		m.executorDurationSec,
		newGaugeFunc("netstamp_probe_worker_queue_depth", "Current worker queue depth.", options.WorkerQueueDepth),
		newGaugeFunc("netstamp_probe_worker_queue_capacity", "Configured worker queue capacity.", options.WorkerQueueCapacity),
		newGaugeFunc("netstamp_probe_result_queue_depth", "Current result queue depth.", options.ResultQueueDepth),
		newGaugeFunc("netstamp_probe_result_queue_capacity", "Configured result queue capacity.", options.ResultQueueCapacity),
	)

	return m
}

func newGaugeFunc(name, help string, value QueueValueFunc) prometheus.GaugeFunc {
	if value == nil {
		value = func() float64 { return 0 }
	}
	return prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, value)
}

func (m *Metrics) Handler() http.Handler {
	if m == nil || m.handler == nil {
		return http.NotFoundHandler()
	}
	return m.handler
}

func (m *Metrics) IncScheduledRun() {
	if m != nil {
		m.scheduledRuns.Inc()
	}
}

func (m *Metrics) IncSkippedRun(reason string) {
	if m != nil {
		m.skippedRuns.WithLabelValues(reason).Inc()
	}
}

func (m *Metrics) IncDroppedResult(reason string) {
	if m != nil {
		m.droppedResults.WithLabelValues(reason).Inc()
	}
}

func (m *Metrics) IncSubmitFailure() {
	if m != nil {
		m.submitFailures.Inc()
	}
}

func (m *Metrics) IncSubmitRetry() {
	if m != nil {
		m.submitRetries.Inc()
	}
}

func (m *Metrics) ObserveSubmitBatchSize(size int) {
	if m != nil {
		m.submitBatchSize.Observe(float64(size))
	}
}

func (m *Metrics) IncActiveWorker() {
	if m != nil {
		m.activeWorkers.Inc()
	}
}

func (m *Metrics) DecActiveWorker() {
	if m != nil {
		m.activeWorkers.Dec()
	}
}

func (m *Metrics) ObserveExecutorDuration(checkType string, duration time.Duration) {
	if m != nil {
		m.executorDurationSec.WithLabelValues(checkType).Observe(duration.Seconds())
	}
}
