package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	applabel "github.com/yorukot/netstamp/internal/application/label"
	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	appproject "github.com/yorukot/netstamp/internal/application/project"
	"github.com/yorukot/netstamp/internal/config"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	pglabel "github.com/yorukot/netstamp/internal/infrastructure/postgres/label"
	pgprobe "github.com/yorukot/netstamp/internal/infrastructure/postgres/probe"
	pgproject "github.com/yorukot/netstamp/internal/infrastructure/postgres/project"
	pguser "github.com/yorukot/netstamp/internal/infrastructure/postgres/user"
	"github.com/yorukot/netstamp/internal/infrastructure/security"
	"github.com/yorukot/netstamp/internal/logger"
	"github.com/yorukot/netstamp/internal/observability/tracing"
	httpserver "github.com/yorukot/netstamp/internal/transport/http"
)

type Application struct {
	Config     config.Config
	Log        *zap.Logger
	HTTPServer *http.Server
	DBPool     *pgxpool.Pool
	Tracing    *tracing.Provider
}

func New(ctx context.Context) (*Application, error) {
	// Load configuration from environment variables and .env file
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Creating logger before database connection to ensure we can log any errors that occur during startup
	log, _, err := logger.New(logger.Config{
		Env:     cfg.Env,
		Service: cfg.ServiceName,
		Version: cfg.Version,
		Level:   cfg.LogLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	// Setup
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Warn("otel_error", zap.Error(err))
	}))

	tracingProvider, err := tracing.NewProvider(ctx, tracing.Config{
		Env:                cfg.Env,
		ServiceName:        cfg.ServiceName,
		ServiceVersion:     cfg.Version,
		OTLPTracesEndpoint: cfg.Tracing.OTLPTracesEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("create tracing provider: %w", err)
	}

	// Open database connection pool.
	dbPool, err := postgres.NewPool(ctx, postgres.PoolConfig{
		ConnectionString: cfg.Database.ConnectionString(),
		MaxConns:         cfg.Database.MaxConns,
		MinConns:         cfg.Database.MinConns,
		MaxConnLifetime:  cfg.Database.MaxConnLifetime,
		MaxConnIdleTime:  cfg.Database.MaxConnIdleTime,
	})
	if err != nil {
		return nil, err
	}

	// Initialize application services and handlers
	userRepo := pguser.NewUserRepository(dbPool)
	passwordHasher := security.NewArgon2idPasswordHasher(security.Argon2idConfig{
		MemoryKiB:   cfg.Auth.Argon2idMemoryKiB,
		Iterations:  cfg.Auth.Argon2idIterations,
		Parallelism: cfg.Auth.Argon2idParallelism,
	})
	tokenIssuer := security.NewJWTIssuer(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
	authEvents := logger.NewAuthEventRecorder(log, cfg.LogPseudonymKey)
	projectEvents := logger.NewProjectEventRecorder(log)
	labelEvents := logger.NewLabelEventRecorder(log)
	probeEvents := logger.NewProbeEventRecorder(log)

	authSvc := appauth.NewService(userRepo, passwordHasher, tokenIssuer, authEvents)
	projectRepo := pgproject.NewProjectRepository(dbPool)
	projectSvc := appproject.NewService(projectRepo, projectEvents)
	labelRepo := pglabel.NewLabelRepository(dbPool)
	labelSvc := applabel.NewService(labelRepo, projectRepo, labelEvents)
	probeRepo := pgprobe.NewProbeRepository(dbPool)
	probeSvc := appprobe.NewService(probeRepo, projectRepo, labelRepo, security.NewProbeSecretGenerator(), probeEvents)
	readiness := postgres.NewReadinessCheck(dbPool)

	httpHandler := httpserver.NewRouter(httpserver.Dependencies{
		Log:            log,
		APIVersion:     cfg.APIVersion,
		BackendBaseURL: cfg.HTTP.BackendBaseURL,
		AuthService:    authSvc,
		AuthVerifier:   tokenIssuer,
		LabelService:   labelSvc,
		ProbeService:   probeSvc,
		ProjectService: projectSvc,
		ReadinessCheck: readiness,
		RequestTimeout: cfg.HTTP.RequestTimeout,
	})

	return &Application{
		Config:     cfg,
		Log:        log,
		HTTPServer: httpserver.NewServer(cfg.HTTP, httpHandler),
		DBPool:     dbPool,
		Tracing:    tracingProvider,
	}, nil
}
