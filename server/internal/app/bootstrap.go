package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	appproject "github.com/yorukot/netstamp/internal/application/project"
	"github.com/yorukot/netstamp/internal/config"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	pgproject "github.com/yorukot/netstamp/internal/infrastructure/postgres/project"
	pguser "github.com/yorukot/netstamp/internal/infrastructure/postgres/user"
	"github.com/yorukot/netstamp/internal/infrastructure/security"
	"github.com/yorukot/netstamp/internal/logger"
	"github.com/yorukot/netstamp/internal/observability/tracing"
	grpcserver "github.com/yorukot/netstamp/internal/transport/grpc"
	httpserver "github.com/yorukot/netstamp/internal/transport/http"
)

type Application struct {
	Config     config.Config
	Log        *zap.Logger
	HTTPServer *http.Server
	GRPCServer *grpc.Server
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

	authSvc := appauth.NewService(userRepo, passwordHasher, tokenIssuer, authEvents)
	projectRepo := pgproject.NewProjectRepository(dbPool)
	projectSvc := appproject.NewService(projectRepo)
	readiness := postgres.NewReadinessCheck(dbPool)

	httpHandler := httpserver.NewRouter(httpserver.Dependencies{
		Log:            log,
		APIVersion:     cfg.APIVersion,
		BackendBaseURL: cfg.HTTP.BackendBaseURL,
		AuthService:    authSvc,
		AuthVerifier:   tokenIssuer,
		ProjectService: projectSvc,
		ReadinessCheck: readiness,
		RequestTimeout: cfg.HTTP.RequestTimeout,
	})

	// GRPC server setup will be done in the Run method, as it requires the application context and dependencies to be fully initialized.
	return &Application{
		Config:     cfg,
		Log:        log,
		HTTPServer: httpserver.NewServer(cfg.HTTP, httpHandler),
		GRPCServer: grpcserver.New(grpcserver.Dependencies{
			Log:         log,
			ServiceName: cfg.ServiceName,
		}),
		DBPool:  dbPool,
		Tracing: tracingProvider,
	}, nil
}
