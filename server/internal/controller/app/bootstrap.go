package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/yorukot/netstamp/internal/controller/config"
	"github.com/yorukot/netstamp/internal/controller/logger"
	httpserver "github.com/yorukot/netstamp/internal/controller/transport/http"
	obmetrics "github.com/yorukot/netstamp/internal/platform/observability/metrics"
	"github.com/yorukot/netstamp/internal/platform/observability/tracing"
)

type Application struct {
	Config     config.Config
	Log        *zap.Logger
	HTTPServer *http.Server
	DBPool     *pgxpool.Pool
	Metrics    *obmetrics.Provider
	Tracing    *tracing.Provider
	Workers    []backgroundWorker
}

type backgroundWorker interface {
	Run(context.Context) error
}

func New(ctx context.Context) (*Application, error) {
	// Load configuration from environment variables and .env file
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := buildLogger(cfg)
	if err != nil {
		return nil, err
	}

	metricsProvider, tracingProvider, err := buildObservability(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	dbPool, err := buildDBPool(ctx, cfg)
	if err != nil {
		return nil, err
	}

	services := buildControllerServices(cfg, log, dbPool)
	httpHandler, err := buildHTTPHandler(cfg, log, dbPool, metricsProvider, services)
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:     cfg,
		Log:        log,
		HTTPServer: httpserver.NewServer(cfg.HTTP, httpHandler),
		DBPool:     dbPool,
		Metrics:    metricsProvider,
		Tracing:    tracingProvider,
		Workers:    services.backgroundWorkers,
	}, nil
}

func buildLogger(cfg config.Config) (*zap.Logger, error) {
	// Creating logger before database connection ensures startup failures are visible.
	log, _, err := logger.New(logger.Config{
		Env:     cfg.Env,
		Service: cfg.ServiceName,
		Version: cfg.Version,
		Level:   cfg.LogLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	return log, nil
}
