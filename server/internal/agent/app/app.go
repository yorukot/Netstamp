package app

import (
	"context"
	"log/slog"
	"os"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/executor"
	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
	"github.com/yorukot/netstamp/internal/agent/result"
	agentruntime "github.com/yorukot/netstamp/internal/agent/runtime"
	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
)

type App struct {
	runtime *agentruntime.Service
}

const WorkerQueueCapacityFactor = 2

func New(config agentconfig.Config, log *slog.Logger) *App {
	// Client is the HTTP client used to communicate with the controller
	client := httpclient.New(config)

	// ResultQueue is the queue where worker results are submitted
	resultQueue := agentworker.NewResultQueue(config.ResultQueueSize.Value, log)
	// ResultSubmitter is responsible for submitting results back to the controller
	resultSubmitter := result.New(*client, resultQueue, config, log)

	// WorkerQueue is the channel where run requests are submitted to workers
	workerQueue := make(chan scheduling.RunRequest, config.MaxWorkers.Value*2)
	// PingExecutor is responsible for executing ping checks
	pingExecutor := executor.NewICMPPingExecutor()
	// WorkerPool is responsible for managing worker execution
	workerPool := agentworker.NewWorkerPool(config.MaxWorkers.Value, workerQueue, resultQueue, pingExecutor, log)

	// AssignmentStore is responsible for storing and managing assignments
	assignmentStore := scheduling.NewAssignmentStore(config.ProbeID, config.AssignmentTTL.Value, log)
	// Scheduler is responsible for scheduling run requests to workers
	scheduler := scheduling.NewScheduler(assignmentStore, workerQueue, log)

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
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.runtime.Run(ctx)
}

func NewLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
