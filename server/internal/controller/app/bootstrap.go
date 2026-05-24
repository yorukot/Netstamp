package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/config"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pgassignment "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/assignment"
	pgcheck "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/check"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	pgping "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/ping"
	pgprobe "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/probe"
	pgproject "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/project"
	pgresult "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/result"
	pgtcp "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/tcp"
	pgtraceroute "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/traceroute"
	pguser "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/user"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/security"
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

	metricsProvider, err := obmetrics.NewProvider(obmetrics.Config{
		Env:            cfg.Env,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.Version,
	})
	if err != nil {
		return nil, fmt.Errorf("create metrics provider: %w", err)
	}

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
	userEvents := logger.NewUserEventRecorder(log, cfg.LogPseudonymKey)
	projectEvents := logger.NewProjectEventRecorder(log)
	labelEvents := logger.NewLabelEventRecorder(log)
	checkEvents := logger.NewCheckEventRecorder(log)
	probeEvents := logger.NewProbeEventRecorder(log)
	probeRuntimeEvents := logger.NewProbeRuntimeEventRecorder(log)
	assignmentEvents := logger.NewAssignmentEventRecorder(log)

	authSvc := appauth.NewService(userRepo, passwordHasher, tokenIssuer, authEvents)
	userSvc := appuser.NewService(userRepo, passwordHasher, userEvents)
	projectRepo := pgproject.NewProjectRepository(dbPool)
	projectSvc := appproject.NewService(projectRepo, userRepo, projectEvents)
	labelRepo := pglabel.NewLabelRepository(dbPool)
	assignmentRepo := pgassignment.NewAssignmentRepository(dbPool)
	assignmentSvc := appassignment.NewService(assignmentRepo, projectRepo, assignmentEvents)
	probeRepo := pgprobe.NewProbeRepository(dbPool)
	labelSvc := applabel.NewService(labelRepo, projectRepo, labelEvents, assignmentSvc)
	checkRepo := pgcheck.NewCheckRepository(dbPool)
	checkSvc := appcheck.NewService(checkRepo, projectRepo, labelRepo, assignmentSvc, checkEvents)
	probeSvc := appprobe.NewService(probeRepo, projectRepo, labelRepo, assignmentSvc, security.NewProbeSecretGenerator(), probeEvents)
	pingRepo := pgping.NewPingRepository(dbPool)
	tcpRepo := pgtcp.NewTCPRepository(dbPool)
	tracerouteRepo := pgtraceroute.NewTracerouteRepository(dbPool)
	resultRepo := pgresult.NewResultRepository(dbPool)
	probeRuntimeSvc := appproberuntime.NewServiceWithTCP(probeRepo, pingRepo, tcpRepo, tracerouteRepo, security.NewProbeSecretVerifier(), probeRuntimeEvents)
	resultSvc := appresult.NewService(pingRepo, tracerouteRepo, resultRepo, projectRepo)
	readiness := postgres.NewReadinessCheck(dbPool)

	httpHandler := httpserver.NewRouter(httpserver.Dependencies{
		Log:               log,
		APIVersion:        cfg.APIVersion,
		BackendBaseURL:    cfg.HTTP.BackendBaseURL,
		AuthService:       authSvc,
		AuthVerifier:      tokenIssuer,
		AuthCookieSecure:  cfg.Env != "local",
		UserService:       userSvc,
		AssignmentService: assignmentSvc,
		CheckService:      checkSvc,
		LabelService:      labelSvc,
		ProbeService:      probeSvc,
		ProbeRuntime:      probeRuntimeSvc,
		ProjectService:    projectSvc,
		ResultService:     resultSvc,
		ReadinessCheck:    readiness,
		RequestTimeout:    cfg.HTTP.RequestTimeout,
		MetricsHandler:    metricsProvider.Handler(),
	})

	return &Application{
		Config:     cfg,
		Log:        log,
		HTTPServer: httpserver.NewServer(cfg.HTTP, httpHandler),
		DBPool:     dbPool,
		Metrics:    metricsProvider,
		Tracing:    tracingProvider,
	}, nil
}
