package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appalerteval "github.com/yorukot/netstamp/internal/controller/application/alerteval"
	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	apppublicstatus "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/config"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/notify"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pgalert "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/alert"
	pgassignment "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/assignment"
	pgcheck "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/check"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	pgping "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/ping"
	pgprobe "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/probe"
	pgproject "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/project"
	pgpublicstatus "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/publicstatus"
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

type controllerServices struct {
	authService              *appauth.Service
	authVerifier             appauth.TokenVerifier
	userService              *appuser.Service
	alertService             *appalert.Service
	alertEmailSMTPConfigured bool
	assignmentService        *appassignment.Service
	checkService             *appcheck.Service
	labelService             *applabel.Service
	probeService             *appprobe.Service
	probeRuntimeService      *appproberuntime.Service
	projectService           *appproject.Service
	publicStatusService      *apppublicstatus.Service
	resultService            *appresult.Service
	backgroundWorkers        []backgroundWorker
}

func buildObservability(ctx context.Context, cfg config.Config, log *zap.Logger) (*obmetrics.Provider, *tracing.Provider, error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Warn("otel_error", zap.Error(err))
	}))

	metricsProvider, err := obmetrics.NewProvider(obmetrics.Config{
		Env:            cfg.Env,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.Version,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create metrics provider: %w", err)
	}

	tracingProvider, err := tracing.NewProvider(ctx, tracing.Config{
		Env:                cfg.Env,
		ServiceName:        cfg.ServiceName,
		ServiceVersion:     cfg.Version,
		OTLPTracesEndpoint: cfg.Tracing.OTLPTracesEndpoint,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create tracing provider: %w", err)
	}

	return metricsProvider, tracingProvider, nil
}

func buildDBPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	return postgres.NewPool(ctx, postgres.PoolConfig{
		ConnectionString: cfg.Database.ConnectionString(),
		MaxConns:         cfg.Database.MaxConns,
		MinConns:         cfg.Database.MinConns,
		MaxConnLifetime:  cfg.Database.MaxConnLifetime,
		MaxConnIdleTime:  cfg.Database.MaxConnIdleTime,
	})
}

func buildControllerServices(cfg config.Config, log *zap.Logger, dbPool *pgxpool.Pool) controllerServices {
	dbTx := postgres.NewTransactor(dbPool)
	userRepo := pguser.NewUserRepository(dbPool)
	projectRepo := pgproject.NewProjectRepository(dbPool)
	alertRepo := pgalert.NewRepository(dbPool)
	labelRepo := pglabel.NewLabelRepository(dbPool)
	assignmentRepo := pgassignment.NewAssignmentRepository(dbPool)
	probeRepo := pgprobe.NewProbeRepository(dbPool)
	checkRepo := pgcheck.NewCheckRepository(dbPool)
	pingRepo := pgping.NewPingRepository(dbPool)
	tcpRepo := pgtcp.NewTCPRepository(dbPool)
	tracerouteRepo := pgtraceroute.NewTracerouteRepository(dbPool)
	resultRepo := pgresult.NewResultRepository(dbPool)
	publicStatusRepo := pgpublicstatus.NewRepository(dbPool)

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

	smtpConfig := notify.SMTPConfig{
		Host:     cfg.Alerting.SMTP.Host,
		Port:     cfg.Alerting.SMTP.Port,
		Username: cfg.Alerting.SMTP.Username,
		Password: cfg.Alerting.SMTP.Password,
		From:     cfg.Alerting.SMTP.From,
		TLSMode:  cfg.Alerting.SMTP.TLSMode,
		Timeout:  cfg.Alerting.SMTP.Timeout,
	}
	notificationSender := notify.NewSender(cfg.Alerting.NotificationHTTPTimeout, smtpConfig)

	authSvc := appauth.NewService(userRepo, passwordHasher, tokenIssuer, authEvents, dbTx)
	authSvc.ConfigurePasswordReset(userRepo, security.NewPasswordResetTokenManager(), notify.NewPasswordResetMailer(smtpConfig), appauth.PasswordResetConfig{
		TokenTTL: cfg.Auth.PasswordResetTokenTTL,
	})
	userSvc := appuser.NewService(userRepo, passwordHasher, userEvents)
	projectSvc := appproject.NewService(projectRepo, userRepo, projectEvents)
	alertSvc := appalert.NewService(alertRepo, projectRepo, notificationSender)
	assignmentSvc := appassignment.NewService(assignmentRepo, projectRepo, assignmentEvents, dbTx)
	labelSvc := applabel.NewService(labelRepo, projectRepo, labelEvents, assignmentSvc, dbTx)
	checkSvc := appcheck.NewService(checkRepo, projectRepo, labelRepo, assignmentSvc, checkEvents, dbTx)
	probeSvc := appprobe.NewService(probeRepo, projectRepo, labelRepo, assignmentSvc, security.NewProbeSecretGenerator(), probeEvents, dbTx)
	publicStatusSvc := apppublicstatus.NewService(publicStatusRepo, projectRepo, pingRepo, tcpRepo)
	probeRuntimeSvc := appproberuntime.NewServiceWithTCP(probeRepo, pingRepo, tcpRepo, tracerouteRepo, security.NewProbeSecretVerifier(), probeRuntimeEvents)
	alertEvalSvc := appalerteval.NewService(alertRepo, cfg.Alerting.EvaluationEnabled, cfg.HTTP.BackendBaseURL, dbTx)
	probeRuntimeSvc.SetAlertEvaluator(alertEvalSvc)
	resultSvc := appresult.NewService(pingRepo, tcpRepo, tracerouteRepo, resultRepo, projectRepo)
	assignmentWorker := appassignment.NewWorker(assignmentRepo, appassignment.WorkerConfig{
		Enabled:      cfg.AssignmentRefresh.WorkerEnabled,
		Interval:     cfg.AssignmentRefresh.WorkerInterval,
		BatchSize:    cfg.AssignmentRefresh.WorkerBatchSize,
		StaleTimeout: cfg.AssignmentRefresh.WorkerStaleTimeout,
	}, appassignment.NewWorkerRefreshRunner(assignmentSvc))
	notificationWorker := appnotification.NewWorker(alertRepo, notificationSender, appnotification.WorkerConfig{
		Enabled:      cfg.Alerting.NotificationWorkerEnabled,
		Interval:     cfg.Alerting.NotificationWorkerInterval,
		BatchSize:    cfg.Alerting.NotificationWorkerBatchSize,
		StaleTimeout: cfg.Alerting.NotificationWorkerStaleTimeout,
	})

	return controllerServices{
		authService:              authSvc,
		authVerifier:             tokenIssuer,
		userService:              userSvc,
		alertService:             alertSvc,
		alertEmailSMTPConfigured: notificationSender.EmailConfigured(),
		assignmentService:        assignmentSvc,
		checkService:             checkSvc,
		labelService:             labelSvc,
		probeService:             probeSvc,
		probeRuntimeService:      probeRuntimeSvc,
		projectService:           projectSvc,
		publicStatusService:      publicStatusSvc,
		resultService:            resultSvc,
		backgroundWorkers:        []backgroundWorker{assignmentWorker, notificationWorker},
	}
}

func buildHTTPHandler(cfg config.Config, log *zap.Logger, dbPool *pgxpool.Pool, metricsProvider *obmetrics.Provider, services controllerServices) (http.Handler, error) {
	trustedProxies, err := cfg.HTTP.TrustedProxyPrefixes()
	if err != nil {
		return nil, fmt.Errorf("parse trusted proxies: %w", err)
	}

	return httpserver.NewRouter(httpserver.Dependencies{
		Log:                         log,
		APIVersion:                  cfg.APIVersion,
		DemoMode:                    cfg.DemoMode,
		BackendBaseURL:              cfg.HTTP.BackendBaseURL,
		PublicWebBaseURL:            cfg.HTTP.PublicWebBaseURL,
		WebDir:                      cfg.HTTP.WebDir,
		AuthService:                 services.authService,
		AuthVerifier:                services.authVerifier,
		AuthCookieSecure:            cfg.Env != "local",
		AuthRegistrationDisabled:    !cfg.Auth.RegistrationEnabled,
		AuthPasswordResetRateWindow: cfg.Auth.PasswordResetRateWindow,
		AuthPasswordResetIPLimit:    cfg.Auth.PasswordResetIPLimit,
		AuthPasswordResetEmailLimit: cfg.Auth.PasswordResetEmailLimit,
		UserService:                 services.userService,
		AlertService:                services.alertService,
		AlertEmailSMTPConfigured:    services.alertEmailSMTPConfigured,
		AssignmentService:           services.assignmentService,
		CheckService:                services.checkService,
		LabelService:                services.labelService,
		ProbeService:                services.probeService,
		ProbeRuntime:                services.probeRuntimeService,
		ProjectService:              services.projectService,
		PublicStatusService:         services.publicStatusService,
		ResultService:               services.resultService,
		ReadinessCheck:              postgres.NewReadinessCheck(dbPool),
		RequestTimeout:              cfg.HTTP.RequestTimeout,
		MetricsHandler:              metricsProvider.Handler(),
		TrustedProxies:              trustedProxies,
	}), nil
}
