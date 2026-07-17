package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appalerteval "github.com/yorukot/netstamp/internal/controller/application/alerteval"
	appapitoken "github.com/yorukot/netstamp/internal/controller/application/apitoken"
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
	pgapitoken "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/apitoken"
	pgassignment "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/assignment"
	pgauthsession "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/authsession"
	pgcheck "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/check"
	pghttpcheck "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/httpcheck"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	pgping "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/ping"
	pgprobe "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/probe"
	pgproject "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/project"
	pgpublicstatus "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/publicstatus"
	pgresult "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/result"
	pgsystem "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/system"
	pgtcp "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/tcp"
	pgtraceroute "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/traceroute"
	pguser "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/user"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/security"
	"github.com/yorukot/netstamp/internal/controller/logger"
	httpserver "github.com/yorukot/netstamp/internal/controller/transport/http"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
	obmetrics "github.com/yorukot/netstamp/internal/platform/observability/metrics"
	"github.com/yorukot/netstamp/internal/platform/observability/tracing"
)

type controllerServices struct {
	authService         *appauth.Service
	authVerifier        appauth.SessionManager
	apiTokenService     *appapitoken.Service
	adminService        *appadmin.Service
	userService         *appuser.Service
	alertService        *appalert.Service
	assignmentService   *appassignment.Service
	checkService        *appcheck.Service
	labelService        *applabel.Service
	probeService        *appprobe.Service
	probeRuntimeService *appproberuntime.Service
	projectService      *appproject.Service
	publicStatusService *apppublicstatus.Service
	resultService       *appresult.Service
	backgroundWorkers   []backgroundWorker
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

type adminSMTPProvider struct {
	service *appadmin.Service
}

func (p adminSMTPProvider) SMTPConfig(ctx context.Context) (notify.SMTPConfig, error) {
	settings, err := p.service.EffectiveSMTP(ctx)
	if err != nil {
		return notify.SMTPConfig{}, err
	}
	return notify.SMTPConfig{
		Host:     settings.Host,
		Port:     settings.Port,
		Username: settings.Username,
		Password: settings.Password,
		From:     settings.From,
		TLSMode:  settings.TLSMode,
		Timeout:  time.Duration(settings.TimeoutSeconds) * time.Second,
	}, nil
}

func buildControllerServices(cfg config.Config, log *zap.Logger, dbPool *pgxpool.Pool) (controllerServices, error) {
	dbTx := postgres.NewTransactor(dbPool)
	userRepo := pguser.NewUserRepository(dbPool)
	authSessionRepo := pgauthsession.NewRepository(dbPool)
	apiTokenRepo := pgapitoken.NewRepository(dbPool)
	systemRepo := pgsystem.NewRepository(dbPool)
	projectRepo := pgproject.NewProjectRepository(dbPool)
	alertRepo := pgalert.NewRepository(dbPool)
	labelRepo := pglabel.NewLabelRepository(dbPool)
	assignmentRepo := pgassignment.NewAssignmentRepository(dbPool)
	probeRepo := pgprobe.NewProbeRepository(dbPool)
	checkRepo := pgcheck.NewCheckRepository(dbPool)
	pingRepo := pgping.NewPingRepository(dbPool)
	tcpRepo := pgtcp.NewTCPRepository(dbPool)
	httpRepo := pghttpcheck.NewRepository(dbPool)
	tracerouteRepo := pgtraceroute.NewTracerouteRepository(dbPool)
	resultRepo := pgresult.NewResultRepository(dbPool)
	publicStatusRepo := pgpublicstatus.NewRepository(dbPool)

	passwordHasher := security.NewArgon2idPasswordHasher(security.Argon2idConfig{
		MemoryKiB:   cfg.Auth.Argon2idMemoryKiB,
		Iterations:  cfg.Auth.Argon2idIterations,
		Parallelism: cfg.Auth.Argon2idParallelism,
	})
	sessionManager := security.NewSessionManager(authSessionRepo, security.SessionConfig{
		HashKey:       cfg.Auth.SessionHashKey,
		IdleTTL:       cfg.Auth.SessionIdleTTL,
		AbsoluteTTL:   cfg.Auth.SessionAbsoluteTTL,
		TouchInterval: cfg.Auth.SessionTouchInterval,
		SudoTTL:       cfg.Auth.SudoTTL,
	})
	apiTokenManager := security.NewAPITokenManager(cfg.Auth.APITokenHashKey)
	authEvents := logger.NewAuthEventRecorder(log, cfg.LogPseudonymKey)
	apiTokenEvents := logger.NewAPITokenEventRecorder(log)
	userEvents := logger.NewUserEventRecorder(log, cfg.LogPseudonymKey)
	projectEvents := logger.NewProjectEventRecorder(log)
	alertEvents := logger.NewAlertEventRecorder(log)
	alertEvalEvents := logger.NewAlertEvalEventRecorder(log)
	labelEvents := logger.NewLabelEventRecorder(log)
	checkEvents := logger.NewCheckEventRecorder(log)
	probeEvents := logger.NewProbeEventRecorder(log)
	probeRuntimeEvents := logger.NewProbeRuntimeEventRecorder(log)
	assignmentEvents := logger.NewAssignmentEventRecorder(log)
	publicStatusEvents := logger.NewPublicStatusEventRecorder(log)

	secretCipher, err := security.NewSecretCipher(cfg.SettingsSecretKey)
	if err != nil {
		return controllerServices{}, fmt.Errorf("create system settings cipher: %w", err)
	}
	adminSvc := appadmin.NewService(systemRepo, secretCipher, appadmin.Defaults{
		RegistrationEnabled: cfg.Auth.RegistrationEnabled,
		BackendBaseURL:      cfg.HTTP.BackendBaseURL,
		PublicWebBaseURL:    cfg.HTTP.PublicWebBaseURL,
		SMTP: appadmin.SMTPSettings{
			Host:           cfg.Alerting.SMTP.Host,
			Port:           cfg.Alerting.SMTP.Port,
			Username:       cfg.Alerting.SMTP.Username,
			Password:       cfg.Alerting.SMTP.Password,
			From:           cfg.Alerting.SMTP.From,
			TLSMode:        cfg.Alerting.SMTP.TLSMode,
			TimeoutSeconds: appadmin.DurationSeconds(cfg.Alerting.SMTP.Timeout),
		},
	}, passwordHasher)
	adminSvc.ConfigureSessions(sessionManager)
	adminSvc.ConfigureAuthenticationMethods(userRepo)
	smtpProvider := adminSMTPProvider{service: adminSvc}
	notificationSender := notify.NewDynamicSender(cfg.Alerting.NotificationHTTPTimeout, smtpProvider)

	authSvc := appauth.NewService(userRepo, passwordHasher, sessionManager, authEvents, dbTx)
	authSvc.ConfigureSystemAdmin(systemRepo)
	authSvc.ConfigurePasswordReset(userRepo, security.NewPasswordResetTokenManager(), notify.NewDynamicPasswordResetMailer(smtpProvider), appauth.PasswordResetConfig{
		TokenTTL: cfg.Auth.PasswordResetTokenTTL,
	})
	authSvc.ConfigureEmailVerification(userRepo, security.NewPasswordResetTokenManager(), notify.NewDynamicPasswordResetMailer(smtpProvider), appauth.EmailVerificationConfig{
		TokenTTL: appauth.DefaultEmailVerificationTokenTTL,
	})
	externalAuthProviders := make([]appauth.ExternalProviderRegistration, 0, 3)
	if cfg.Auth.OIDCEnabled {
		externalAuthProviders = append(externalAuthProviders, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID: identity.AuthenticationMethodOIDC, DisplayName: cfg.Auth.OIDCDisplayName,
				JITEnabled: cfg.Auth.OIDCJITEnabled, SudoCapable: true,
			},
			Client: security.NewOIDCClient(security.OIDCClientConfig{
				IssuerURL: cfg.Auth.OIDCIssuerURL, ClientID: cfg.Auth.OIDCClientID, ClientSecret: cfg.Auth.OIDCClientSecret,
				RedirectURL: strings.TrimRight(cfg.HTTP.BackendBaseURL, "/") + "/api/" + cfg.APIVersion + "/auth/external/oidc/callback",
			}),
		})
	}
	if cfg.Auth.GoogleEnabled {
		externalAuthProviders = append(externalAuthProviders, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID: identity.AuthenticationMethodGoogle, DisplayName: cfg.Auth.GoogleDisplayName,
				JITEnabled: cfg.Auth.GoogleJITEnabled, SudoCapable: true,
			},
			Client: security.NewGoogleOIDCClient(security.GoogleOIDCClientConfig{
				ClientID: cfg.Auth.GoogleClientID, ClientSecret: cfg.Auth.GoogleClientSecret,
				RedirectURL:          strings.TrimRight(cfg.HTTP.BackendBaseURL, "/") + "/api/" + cfg.APIVersion + "/auth/external/google/callback",
				AllowedHostedDomains: strings.Split(cfg.Auth.GoogleHostedDomains, ","),
			}),
		})
	}
	if cfg.Auth.GitHubEnabled {
		externalAuthProviders = append(externalAuthProviders, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID: identity.AuthenticationMethodGitHub, DisplayName: cfg.Auth.GitHubDisplayName,
				JITEnabled: cfg.Auth.GitHubJITEnabled, SudoCapable: true,
			},
			Client: security.NewGitHubOAuthClient(security.GitHubOAuthClientConfig{
				ClientID: cfg.Auth.GitHubClientID, ClientSecret: cfg.Auth.GitHubClientSecret,
				RedirectURL: strings.TrimRight(cfg.HTTP.BackendBaseURL, "/") + "/api/" + cfg.APIVersion + "/auth/external/github/callback",
				AllowSignup: cfg.Auth.GitHubAllowSignup,
			}),
		})
	}
	authSvc.ConfigureExternalAuth(userRepo, security.NewPasswordResetTokenManager(), appauth.ExternalAuthConfig{
		FlowTTL: cfg.Auth.ExternalFlowTTL, AuthTimeSkew: time.Minute,
	}, externalAuthProviders...)

	userSvc := appuser.NewService(userRepo, passwordHasher, userEvents)
	apiTokenSvc := appapitoken.NewService(apiTokenRepo, apiTokenManager, apiTokenEvents)
	authSvc.ConfigureAPITokens(apiTokenSvc)
	adminSvc.ConfigureAPITokens(apiTokenSvc)
	userSvc.ConfigureAPITokens(apiTokenSvc)
	userSvc.ConfigureAuthenticationMethods(userRepo)
	userSvc.ConfigureSystemAdmin(systemRepo)
	userSvc.ConfigureSessions(sessionManager)
	projectSvc := appproject.NewService(projectRepo, userRepo, projectEvents)
	alertSvc := appalert.NewService(alertRepo, projectRepo, alertEvents, notificationSender)
	assignmentSvc := appassignment.NewService(assignmentRepo, projectRepo, assignmentEvents, dbTx)
	labelSvc := applabel.NewService(labelRepo, projectRepo, labelEvents, assignmentSvc, dbTx)
	checkSvc := appcheck.NewService(checkRepo, projectRepo, labelRepo, assignmentSvc, checkEvents, dbTx)
	probeSvc := appprobe.NewService(probeRepo, projectRepo, labelRepo, assignmentSvc, security.NewProbeSecretGenerator(), probeEvents, dbTx)
	publicStatusSvc := apppublicstatus.NewService(publicStatusRepo, projectRepo, publicStatusEvents, pingRepo, tcpRepo)
	publicStatusSvc.ConfigureHTTP(httpRepo)
	probeRuntimeSvc := appproberuntime.NewServiceWithResults(probeRepo, pingRepo, tcpRepo, httpRepo, tracerouteRepo, security.NewProbeSecretVerifier(), probeRuntimeEvents)
	alertEvalSvc := appalerteval.NewServiceWithEvents(alertRepo, cfg.Alerting.EvaluationEnabled, cfg.HTTP.BackendBaseURL, alertEvalEvents, dbTx)
	alertEvalSvc.ConfigureBackendBaseURLProvider(adminSvc)
	probeRuntimeSvc.SetAlertEvaluator(alertEvalSvc)
	resultSvc := appresult.NewServiceWithHTTP(pingRepo, tcpRepo, httpRepo, tracerouteRepo, resultRepo, projectRepo)
	assignmentWorker := appassignment.NewWorker(assignmentRepo, appassignment.WorkerConfig{
		Enabled:      cfg.AssignmentRefresh.WorkerEnabled,
		Interval:     cfg.AssignmentRefresh.WorkerInterval,
		BatchSize:    cfg.AssignmentRefresh.WorkerBatchSize,
		StaleTimeout: cfg.AssignmentRefresh.WorkerStaleTimeout,
		Log:          log,
	}, appassignment.NewWorkerRefreshRunner(assignmentSvc))
	notificationWorker := appnotification.NewWorker(alertRepo, notificationSender, appnotification.WorkerConfig{
		Enabled:      cfg.Alerting.NotificationWorkerEnabled,
		Interval:     cfg.Alerting.NotificationWorkerInterval,
		BatchSize:    cfg.Alerting.NotificationWorkerBatchSize,
		StaleTimeout: cfg.Alerting.NotificationWorkerStaleTimeout,
		Log:          log,
	})

	return controllerServices{
		authService:         authSvc,
		authVerifier:        sessionManager,
		apiTokenService:     apiTokenSvc,
		adminService:        adminSvc,
		userService:         userSvc,
		alertService:        alertSvc,
		assignmentService:   assignmentSvc,
		checkService:        checkSvc,
		labelService:        labelSvc,
		probeService:        probeSvc,
		probeRuntimeService: probeRuntimeSvc,
		projectService:      projectSvc,
		publicStatusService: publicStatusSvc,
		resultService:       resultSvc,
		backgroundWorkers:   []backgroundWorker{assignmentWorker, notificationWorker},
	}, nil
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
		APITokenService:             services.apiTokenService,
		APITokenVerifier:            services.apiTokenService,
		AdminService:                services.adminService,
		AuthCookieName:              authCookieName(cfg),
		AuthCookieSecure:            authCookieSecure(cfg),
		AuthRegistrationDisabled:    !cfg.Auth.RegistrationEnabled,
		AuthPasswordResetRateWindow: cfg.Auth.PasswordResetRateWindow,
		AuthPasswordResetIPLimit:    cfg.Auth.PasswordResetIPLimit,
		AuthPasswordResetEmailLimit: cfg.Auth.PasswordResetEmailLimit,
		UserService:                 services.userService,
		AlertService:                services.alertService,
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

func authCookieSecure(cfg config.Config) bool {
	return cfg.Env != "local"
}

func authCookieName(cfg config.Config) string {
	if authCookieSecure(cfg) {
		return httpmiddleware.ProductionSessionCookieName
	}
	return httpmiddleware.LocalSessionCookieName
}
