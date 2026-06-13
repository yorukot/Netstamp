package app

import (
	"context"
	"log/slog"
	"os"

	"golang.org/x/sync/errgroup"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/executor"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
	"github.com/yorukot/netstamp/internal/agent/observability"
	"github.com/yorukot/netstamp/internal/agent/result"
	agentruntime "github.com/yorukot/netstamp/internal/agent/runtime"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
)

type App struct {
	runtime     *agentruntime.Service
	metrics     *observability.Metrics
	metricsAddr string
	pprofAddr   string
	log         *slog.Logger
}

const WorkerQueueCapacityFactor = 2

func New(config agentconfig.Config, log *slog.Logger) *App {
	// Client is the HTTP client used to communicate with the controller
	client := httpclient.New(config)

	// ResultQueue is the queue where worker results are submitted
	resultQueue := agentworker.NewResultQueue(config.ResultQueueSize.Value, log)
	// WorkerQueue is the channel where run requests are submitted to workers
	workerQueue := make(chan scheduling.RunRequest, config.MaxWorkers.Value*2)
	metrics := observability.NewMetrics(observability.MetricsOptions{
		WorkerQueueDepth: func() float64 {
			return float64(len(workerQueue))
		},
		WorkerQueueCapacity: func() float64 {
			return float64(cap(workerQueue))
		},
		ResultQueueDepth: func() float64 {
			return float64(resultQueue.Depth())
		},
		ResultQueueCapacity: func() float64 {
			return float64(resultQueue.Capacity())
		},
	})
	resultQueue.SetMetrics(metrics)
	// ResultSubmitter is responsible for submitting results back to the controller
	resultSubmitter := result.New(*client, resultQueue, config, log, metrics)

	// PingExecutor is responsible for executing ping checks
	pingExecutor := executor.NewICMPPingExecutor()
	// TCPExecutor is responsible for executing TCP connect checks
	tcpExecutor := executor.NewTCPExecutor()
	// TracerouteExecutor is responsible for executing traceroute checks
	tracerouteExecutor := executor.NewTracerouteExecutor()
	// WorkerPool is responsible for managing worker execution
	workerPool := agentworker.NewWorkerPool(config.MaxWorkers.Value, workerQueue, resultQueue, pingExecutor, tcpExecutor, tracerouteExecutor, log, metrics)

	// AssignmentStore is responsible for storing and managing assignments
	assignmentStore := scheduling.NewAssignmentStore(config.ProbeID, config.AssignmentTTL.Value, log)
	// Scheduler is responsible for scheduling run requests to workers
	scheduler := scheduling.NewScheduler(assignmentStore, workerQueue, log, metrics)

	return &App{
		runtime: &agentruntime.Service{
			Client:      *client,
			Config:      &config,
			Assignments: assignmentStore,
			Scheduler:   scheduler,
			Workers:     workerPool,
			Results:     resultSubmitter,
			Log:         log,
		},
		metrics:     metrics,
		metricsAddr: config.MetricsAddr.Value,
		pprofAddr:   config.PprofAddr.Value,
		log:         log,
	}
}

func (a *App) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return a.runtime.Run(groupCtx)
	})
	if a.metricsAddr != "" {
		group.Go(func() error {
			return observability.RunHTTPServer(groupCtx, a.metricsAddr, a.metrics.Handler(), "metrics", a.log)
		})
	}
	if a.pprofAddr != "" {
		group.Go(func() error {
			return observability.RunHTTPServer(groupCtx, a.pprofAddr, observability.PprofHandler(), "pprof", a.log)
		})
	}

	return group.Wait()
}

func NewLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
