package app

import (
	"context"
	"log/slog"
	"os"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/executor"
	agentnetwork "github.com/yorukot/netstamp/internal/agent/infrastructure/network"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/runtimeclient"
	"github.com/yorukot/netstamp/internal/agent/observability"
	agentruntime "github.com/yorukot/netstamp/internal/agent/runtime"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
)

type App struct {
	runtime *agentruntime.Service
}

func New(config agentconfig.Config, log *slog.Logger) *App {
	counters := &observability.RuntimeCounters{}
	runtimeConfig := agentruntime.NewRuntimeConfigStore(config.Runtime)
	client := runtimeclient.New(config)
	resultQueue := agentworker.NewResultQueue(config.ResultQueueSize, counters, log)
	assignmentStore := scheduling.NewAssignmentStore(config.ProbeID, config.AssignmentTTL, log)
	workerQueue := make(chan scheduling.RunRequest, agentconfig.WorkerQueueCapacity(config.MaxWorkers))
	scheduler := scheduling.NewScheduler(assignmentStore, workerQueue, counters, log)
	pingExecutor := executor.NewProBingExecutor()
	workerPool := agentworker.NewWorkerPool(config.MaxWorkers, workerQueue, resultQueue, pingExecutor, counters, log)
	resultSubmitter := agentruntime.NewResultSubmitter(client, resultQueue, runtimeConfig, config, counters, log)

	return &App{
		runtime: agentruntime.NewService(agentruntime.ServiceDependencies{
			Client:          client,
			StatusProvider:  agentnetwork.NewHeartbeatStatusProvider(),
			RuntimeConfig:   runtimeConfig,
			Assignments:     assignmentStore,
			Scheduler:       scheduler,
			Workers:         workerPool,
			Results:         resultSubmitter,
			Counters:        counters,
			ShutdownTimeout: config.ShutdownTimeout,
			Log:             log,
		}),
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.runtime.Run(ctx)
}

func NewLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
