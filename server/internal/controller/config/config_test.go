package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testDatabaseConnectionString = "postgres://netstamp:netstamp@localhost:5432/netstamp?sslmode=disable"

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Env != "local" {
		t.Fatalf("expected local env, got %q", cfg.Env)
	}
	if cfg.DemoMode {
		t.Fatal("expected demo mode to be disabled by default")
	}
	if cfg.ServiceName != "controller" {
		t.Fatalf("expected default service name, got %q", cfg.ServiceName)
	}
	if cfg.Version != "0.1.0" {
		t.Fatalf("expected default app version, got %q", cfg.Version)
	}
	if cfg.APIVersion != "v1" {
		t.Fatalf("expected default API version, got %q", cfg.APIVersion)
	}
	if cfg.LogPseudonymKey != "local-development-log-pseudonym-key-change-before-production" {
		t.Fatalf("expected default log pseudonym key, got %q", cfg.LogPseudonymKey)
	}
	if cfg.SettingsSecretKey != "local-development-system-settings-encryption-key-change-before-production" {
		t.Fatalf("expected default system settings secret key, got %q", cfg.SettingsSecretKey)
	}
	if cfg.HTTP.BackendBaseURL != "" {
		t.Fatalf("expected empty backend base URL, got %q", cfg.HTTP.BackendBaseURL)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("expected default HTTP addr, got %q", cfg.HTTP.Addr)
	}
	if cfg.HTTP.WebDir != "" {
		t.Fatalf("expected empty web dir, got %q", cfg.HTTP.WebDir)
	}
	if cfg.HTTP.TrustedProxies != "" {
		t.Fatalf("expected empty trusted proxies, got %q", cfg.HTTP.TrustedProxies)
	}
	if cfg.HTTP.RequestTimeout != 10*time.Second {
		t.Fatalf("expected default request timeout, got %s", cfg.HTTP.RequestTimeout)
	}
	if cfg.Database.Host != "localhost" {
		t.Fatalf("expected database host, got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected database port, got %d", cfg.Database.Port)
	}
	if cfg.Database.ConnectionString() != testDatabaseConnectionString {
		t.Fatalf("expected connection string, got %q", cfg.Database.ConnectionString())
	}
	if !cfg.Auth.RegistrationEnabled {
		t.Fatal("expected registration to be enabled by default")
	}
	if cfg.Auth.SudoTTL != 5*time.Minute {
		t.Fatalf("expected five-minute sudo TTL, got %s", cfg.Auth.SudoTTL)
	}
	if cfg.Auth.OIDCEnabled {
		t.Fatal("expected OIDC to be disabled by default")
	}
	if cfg.Auth.GoogleEnabled || cfg.Auth.GitHubEnabled {
		t.Fatal("expected Google and GitHub authentication to be disabled by default")
	}
	if cfg.Auth.ExternalFlowTTL != 10*time.Minute {
		t.Fatalf("expected ten-minute external auth flow TTL, got %s", cfg.Auth.ExternalFlowTTL)
	}
	if cfg.Alerting.SMTP.Configured() {
		t.Fatal("expected SMTP to be unconfigured by default")
	}
	if !cfg.AssignmentRefresh.WorkerEnabled {
		t.Fatal("expected assignment refresh worker to be enabled by default")
	}
	if cfg.AssignmentRefresh.WorkerInterval != 5*time.Second {
		t.Fatalf("expected default assignment refresh worker interval, got %s", cfg.AssignmentRefresh.WorkerInterval)
	}
	if cfg.AssignmentRefresh.WorkerBatchSize != 25 {
		t.Fatalf("expected default assignment refresh worker batch size, got %d", cfg.AssignmentRefresh.WorkerBatchSize)
	}
	if cfg.AssignmentRefresh.WorkerStaleTimeout != time.Minute {
		t.Fatalf("expected default assignment refresh stale timeout, got %s", cfg.AssignmentRefresh.WorkerStaleTimeout)
	}
	if cfg.Alerting.SMTP.Port != 587 {
		t.Fatalf("expected default SMTP port, got %d", cfg.Alerting.SMTP.Port)
	}
	if cfg.Alerting.SMTP.TLSMode != "starttls" {
		t.Fatalf("expected default SMTP TLS mode, got %q", cfg.Alerting.SMTP.TLSMode)
	}
	if cfg.Alerting.SMTP.Timeout != 10*time.Second {
		t.Fatalf("expected default SMTP timeout, got %s", cfg.Alerting.SMTP.Timeout)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv(keyAppEnv, "production")
	t.Setenv(keyDemoMode, "true")
	t.Setenv(keyServiceName, "netstamp-worker")
	t.Setenv(keyAppVersion, "0.2.0")
	t.Setenv(keyAPIVersion, "v2")
	t.Setenv(keyLogPseudonymKey, "production-log-pseudonym-key")
	t.Setenv(keySystemSettingsEncryptionKey, "production-system-settings-key")
	t.Setenv(keyBackendBaseURL, "https://app.netstamp.dev")
	t.Setenv(keyHTTPAddr, ":8181")
	t.Setenv(keyWebDir, "/app/web")
	t.Setenv(keyHTTPTrustedProxies, "10.0.0.0/8,127.0.0.1")
	t.Setenv(keyRequestTimeout, "250ms")
	t.Setenv(keyDatabaseHost, "db.internal")
	t.Setenv(keyDatabasePort, "15432")
	t.Setenv(keyDatabaseUser, "netstamp_user")
	t.Setenv(keyDatabasePassword, "secret")
	t.Setenv(keyDatabaseName, "netstamp_prod")
	t.Setenv(keyDatabaseSSLMode, "require")
	t.Setenv(keyDBMaxConns, "12")
	t.Setenv(keyAuthRegistrationEnabled, "false")
	t.Setenv(keyAuthSudoTTL, "4m")
	t.Setenv(keyAuthOIDCEnabled, "true")
	t.Setenv(keyAuthOIDCIssuerURL, "https://identity.example.com")
	t.Setenv(keyAuthOIDCClientID, "netstamp")
	t.Setenv(keyAuthOIDCClientSecret, "oidc-secret")
	t.Setenv(keyAuthOIDCDisplayName, "Company SSO")
	t.Setenv(keyAuthOIDCJITEnabled, "true")
	t.Setenv(keyAuthExternalFlowTTL, "8m")
	t.Setenv(keyAuthGoogleEnabled, "true")
	t.Setenv(keyAuthGoogleClientID, "google-client")
	t.Setenv(keyAuthGoogleClientSecret, "google-secret")
	t.Setenv(keyAuthGoogleDisplayName, "Google Workspace")
	t.Setenv(keyAuthGoogleJITEnabled, "true")
	t.Setenv(keyAuthGoogleAllowedHostedDomains, "example.com, example.org")
	t.Setenv(keyAuthGitHubEnabled, "true")
	t.Setenv(keyAuthGitHubClientID, "github-client")
	t.Setenv(keyAuthGitHubClientSecret, "github-secret")
	t.Setenv(keyAuthGitHubDisplayName, "GitHub Enterprise Team")
	t.Setenv(keyAuthGitHubJITEnabled, "true")
	t.Setenv(keyAuthGitHubAllowSignup, "false")
	t.Setenv(keyOTLPTracesEndpoint, "http://victoria-traces:10428/insert/opentelemetry/v1/traces")
	t.Setenv(keyAssignmentRefreshWorkerEnabled, "false")
	t.Setenv(keyAssignmentRefreshWorkerInterval, "7s")
	t.Setenv(keyAssignmentRefreshWorkerBatchSize, "9")
	t.Setenv(keyAssignmentRefreshWorkerStaleTimeout, "2m")
	t.Setenv(keySMTPHost, "smtp.example.com")
	t.Setenv(keySMTPPort, "465")
	t.Setenv(keySMTPUsername, "netstamp")
	t.Setenv(keySMTPPassword, "secret")
	t.Setenv(keySMTPFrom, "alerts@example.com")
	t.Setenv(keySMTPTLSMode, "implicit")
	t.Setenv(keySMTPTimeout, "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.Env)
	}
	if !cfg.DemoMode {
		t.Fatal("expected demo mode to be enabled from environment")
	}
	if cfg.ServiceName != "netstamp-worker" {
		t.Fatalf("expected service override, got %q", cfg.ServiceName)
	}
	if cfg.Version != "0.2.0" {
		t.Fatalf("expected app version override, got %q", cfg.Version)
	}
	if cfg.APIVersion != "v2" {
		t.Fatalf("expected API version override, got %q", cfg.APIVersion)
	}
	if cfg.LogPseudonymKey != "production-log-pseudonym-key" {
		t.Fatalf("expected log pseudonym key override, got %q", cfg.LogPseudonymKey)
	}
	if cfg.SettingsSecretKey != "production-system-settings-key" {
		t.Fatalf("expected system settings secret key override, got %q", cfg.SettingsSecretKey)
	}
	if cfg.HTTP.BackendBaseURL != "https://app.netstamp.dev" {
		t.Fatalf("expected backend base URL override, got %q", cfg.HTTP.BackendBaseURL)
	}
	if cfg.HTTP.Addr != ":8181" {
		t.Fatalf("expected HTTP addr override, got %q", cfg.HTTP.Addr)
	}
	if cfg.HTTP.WebDir != "/app/web" {
		t.Fatalf("expected web dir override, got %q", cfg.HTTP.WebDir)
	}
	if cfg.HTTP.TrustedProxies != "10.0.0.0/8,127.0.0.1" {
		t.Fatalf("expected trusted proxies override, got %q", cfg.HTTP.TrustedProxies)
	}
	trustedProxies, err := cfg.HTTP.TrustedProxyPrefixes()
	if err != nil {
		t.Fatalf("parse trusted proxies: %v", err)
	}
	if len(trustedProxies) != 2 || trustedProxies[0].String() != "10.0.0.0/8" || trustedProxies[1].String() != "127.0.0.1/32" {
		t.Fatalf("unexpected trusted proxy prefixes: %#v", trustedProxies)
	}
	if cfg.HTTP.RequestTimeout != 250*time.Millisecond {
		t.Fatalf("expected request timeout override, got %s", cfg.HTTP.RequestTimeout)
	}
	if cfg.Database.Host != "db.internal" {
		t.Fatalf("expected database host override, got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 15432 {
		t.Fatalf("expected database port override, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "netstamp_user" {
		t.Fatalf("expected database user override, got %q", cfg.Database.User)
	}
	if cfg.Database.Name != "netstamp_prod" {
		t.Fatalf("expected database name override, got %q", cfg.Database.Name)
	}
	if cfg.Database.SSLMode != "require" {
		t.Fatalf("expected database sslmode override, got %q", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxConns != 12 {
		t.Fatalf("expected DB max conns override, got %d", cfg.Database.MaxConns)
	}
	if cfg.Auth.RegistrationEnabled {
		t.Fatal("expected registration to be disabled from environment")
	}
	if cfg.Auth.SudoTTL != 4*time.Minute {
		t.Fatalf("expected sudo TTL override, got %s", cfg.Auth.SudoTTL)
	}
	if !cfg.Auth.OIDCEnabled || !cfg.Auth.OIDCJITEnabled {
		t.Fatal("expected OIDC and JIT provisioning to be enabled from environment")
	}
	if cfg.Auth.OIDCIssuerURL != "https://identity.example.com" || cfg.Auth.OIDCClientID != "netstamp" || cfg.Auth.OIDCDisplayName != "Company SSO" {
		t.Fatalf("unexpected OIDC configuration: %#v", cfg.Auth)
	}
	if cfg.Auth.ExternalFlowTTL != 8*time.Minute || !cfg.Auth.GoogleEnabled || !cfg.Auth.GoogleJITEnabled || cfg.Auth.GoogleHostedDomains != "example.com, example.org" {
		t.Fatalf("unexpected Google configuration: %#v", cfg.Auth)
	}
	if !cfg.Auth.GitHubEnabled || !cfg.Auth.GitHubJITEnabled || cfg.Auth.GitHubAllowSignup || cfg.Auth.GitHubClientID != "github-client" {
		t.Fatalf("unexpected GitHub configuration: %#v", cfg.Auth)
	}
	if cfg.Tracing.OTLPTracesEndpoint != "http://victoria-traces:10428/insert/opentelemetry/v1/traces" {
		t.Fatalf("expected OTLP traces endpoint override, got %q", cfg.Tracing.OTLPTracesEndpoint)
	}
	if cfg.AssignmentRefresh.WorkerEnabled {
		t.Fatal("expected assignment refresh worker to be disabled from environment")
	}
	if cfg.AssignmentRefresh.WorkerInterval != 7*time.Second {
		t.Fatalf("expected assignment refresh worker interval override, got %s", cfg.AssignmentRefresh.WorkerInterval)
	}
	if cfg.AssignmentRefresh.WorkerBatchSize != 9 {
		t.Fatalf("expected assignment refresh worker batch size override, got %d", cfg.AssignmentRefresh.WorkerBatchSize)
	}
	if cfg.AssignmentRefresh.WorkerStaleTimeout != 2*time.Minute {
		t.Fatalf("expected assignment refresh worker stale timeout override, got %s", cfg.AssignmentRefresh.WorkerStaleTimeout)
	}
	if !cfg.Alerting.SMTP.Configured() {
		t.Fatal("expected SMTP to be configured from environment")
	}
	if cfg.Alerting.SMTP.Host != "smtp.example.com" {
		t.Fatalf("expected SMTP host override, got %q", cfg.Alerting.SMTP.Host)
	}
	if cfg.Alerting.SMTP.Port != 465 {
		t.Fatalf("expected SMTP port override, got %d", cfg.Alerting.SMTP.Port)
	}
	if cfg.Alerting.SMTP.Username != "netstamp" {
		t.Fatalf("expected SMTP username override, got %q", cfg.Alerting.SMTP.Username)
	}
	if cfg.Alerting.SMTP.Password != "secret" {
		t.Fatal("expected SMTP password override")
	}
	if cfg.Alerting.SMTP.From != "alerts@example.com" {
		t.Fatalf("expected SMTP from override, got %q", cfg.Alerting.SMTP.From)
	}
	if cfg.Alerting.SMTP.TLSMode != "implicit" {
		t.Fatalf("expected SMTP TLS mode override, got %q", cfg.Alerting.SMTP.TLSMode)
	}
	if cfg.Alerting.SMTP.Timeout != 3*time.Second {
		t.Fatalf("expected SMTP timeout override, got %s", cfg.Alerting.SMTP.Timeout)
	}
}

func TestLoadFromDotEnv(t *testing.T) {
	clearConfigEnv(t)

	dir := t.TempDir()
	t.Chdir(dir)

	err := os.WriteFile(filepath.Join(dir, ".env"), []byte(strings.Join([]string{
		"APP_ENV=staging",
		"SERVICE_NAME=netstamp-staging",
		"HTTP_ADDR=:8282",
		"REQUEST_TIMEOUT=2s",
		"DATABASE_HOST=db.staging.internal",
		"DATABASE_NAME=netstamp_staging",
		"",
	}, "\n")), 0o600)
	if err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Env != "staging" {
		t.Fatalf("expected staging env, got %q", cfg.Env)
	}
	if cfg.ServiceName != "netstamp-staging" {
		t.Fatalf("expected service from .env, got %q", cfg.ServiceName)
	}
	if cfg.HTTP.Addr != ":8282" {
		t.Fatalf("expected HTTP addr from .env, got %q", cfg.HTTP.Addr)
	}
	if cfg.HTTP.RequestTimeout != 2*time.Second {
		t.Fatalf("expected request timeout from .env, got %s", cfg.HTTP.RequestTimeout)
	}
	if cfg.Database.Host != "db.staging.internal" {
		t.Fatalf("expected database host from .env, got %q", cfg.Database.Host)
	}
	if cfg.Database.Name != "netstamp_staging" {
		t.Fatalf("expected database name from .env, got %q", cfg.Database.Name)
	}
}

func TestLoadReturnsValidationErrors(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv(keyRequestTimeout, "not-a-duration")
	t.Setenv(keyBackendBaseURL, "https://app.netstamp.dev/api")
	t.Setenv(keyDatabaseHost, " ")
	t.Setenv(keyDBMaxConns, "-1")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	message := err.Error()
	for _, want := range []string{
		"'REQUEST_TIMEOUT' time: invalid duration",
		"BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials",
		"DATABASE_HOST must not be empty",
		"DB_MAX_CONNS must not be negative",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error to contain %q, got %q", want, message)
		}
	}
}

func TestValidateReturnsErrorsForInvalidValues(t *testing.T) {
	cfg := validConfig()
	cfg.Env = " "
	cfg.ServiceName = ""
	cfg.Version = "\t"
	cfg.LogLevel = "verbose"
	cfg.LogPseudonymKey = ""
	cfg.SettingsSecretKey = ""
	cfg.ShutdownTimeout = 0
	cfg.HTTP.BackendBaseURL = "https://app.netstamp.dev/api"
	cfg.HTTP.Addr = "localhost"
	cfg.HTTP.RequestTimeout = -time.Second
	cfg.HTTP.ReadHeaderTimeout = 0
	cfg.HTTP.ReadTimeout = 0
	cfg.HTTP.WriteTimeout = 0
	cfg.HTTP.IdleTimeout = 0
	cfg.HTTP.TrustedProxies = "10.0.0.0/8,,bad"
	cfg.Database.Host = ""
	cfg.Database.Port = 0
	cfg.Database.User = ""
	cfg.Database.Name = ""
	cfg.Database.SSLMode = "invalid"
	cfg.Database.MaxConns = 0
	cfg.Database.MinConns = 1
	cfg.Database.MaxConnLifetime = 0
	cfg.Database.MaxConnIdleTime = -time.Second
	cfg.Tracing.OTLPTracesEndpoint = "victoria-traces:10428"
	cfg.AssignmentRefresh.WorkerInterval = 0
	cfg.AssignmentRefresh.WorkerBatchSize = 0
	cfg.AssignmentRefresh.WorkerStaleTimeout = -time.Second
	cfg.Alerting.SMTP.Port = 0
	cfg.Alerting.SMTP.Timeout = 0
	cfg.Alerting.SMTP.TLSMode = "ssl"
	cfg.Alerting.SMTP.Username = "netstamp"
	cfg.Alerting.SMTP.From = "alerts"

	err := errors.Join(validate(cfg)...)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	message := err.Error()
	for _, want := range []string{
		"APP_ENV must not be empty",
		"SERVICE_NAME must not be empty",
		"APP_VERSION must not be empty",
		"LOG_LEVEL must be one of debug, info, warn, error, dpanic, panic, or fatal",
		"LOG_PSEUDONYM_KEY must not be empty",
		"SYSTEM_SETTINGS_ENCRYPTION_KEY must not be empty",
		"SHUTDOWN_TIMEOUT must be greater than 0",
		"BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials",
		"HTTP_ADDR must be a host:port address",
		"REQUEST_TIMEOUT must be greater than 0",
		"HTTP_READ_HEADER_TIMEOUT must be greater than 0",
		"HTTP_READ_TIMEOUT must be greater than 0",
		"HTTP_WRITE_TIMEOUT must be greater than 0",
		"HTTP_IDLE_TIMEOUT must be greater than 0",
		"HTTP_TRUSTED_PROXIES must not contain empty entries",
		"DATABASE_HOST must not be empty",
		"DATABASE_USER must not be empty",
		"DATABASE_NAME must not be empty",
		"DATABASE_PORT must be between 1 and 65535",
		"DATABASE_SSLMODE must be one of disable, allow, prefer, require, verify-ca, or verify-full",
		"DB_MAX_CONNS must be greater than 0",
		"DB_MIN_CONNS must not be greater than DB_MAX_CONNS",
		"DB_MAX_CONN_LIFETIME must be greater than 0",
		"DB_MAX_CONN_IDLE_TIME must be greater than 0",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT must be a valid HTTP URL",
		"ASSIGNMENT_REFRESH_WORKER_INTERVAL must be greater than 0",
		"ASSIGNMENT_REFRESH_WORKER_STALE_TIMEOUT must be greater than 0",
		"ASSIGNMENT_REFRESH_WORKER_BATCH_SIZE must be greater than 0",
		"SMTP_PORT must be between 1 and 65535",
		"SMTP_TIMEOUT must be greater than 0",
		"SMTP_TLS_MODE must be one of starttls, implicit, or none",
		"SMTP_HOST must not be empty",
		"SMTP_FROM must be a valid email address",
		"SMTP_USERNAME and SMTP_PASSWORD must be set together",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error to contain %q, got %q", want, message)
		}
	}
}

func TestValidateOptionalHTTPOrigin(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError string
	}{
		{name: "empty", value: ""},
		{name: "http origin", value: "http://localhost:8080"},
		{name: "https origin", value: "https://app.netstamp.dev"},
		{name: "trailing slash", value: "https://app.netstamp.dev/"},
		{name: "missing scheme", value: "app.netstamp.dev", wantError: "BACKEND_BASE_URL must be a valid HTTP origin"},
		{name: "unsupported scheme", value: "ftp://app.netstamp.dev", wantError: "BACKEND_BASE_URL must use http or https"},
		{name: "path", value: "https://app.netstamp.dev/api", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "query", value: "https://app.netstamp.dev?preview=true", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "fragment", value: "https://app.netstamp.dev#api", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "credentials", value: "https://user:pass@app.netstamp.dev", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateOptionalHTTPOrigin(keyBackendBaseURL, tt.value)
			err := errors.Join(errs...)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}

func TestLoadReturnsUnknownDotEnvKeyErrors(t *testing.T) {
	clearConfigEnv(t)

	dir := t.TempDir()
	t.Chdir(dir)

	err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UNKNOWN_SETTING=true\n"), 0o600)
	if err != nil {
		t.Fatalf("write .env: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "has invalid keys: unknown_setting") {
		t.Fatalf("expected unknown key error, got %q", err.Error())
	}
}

func validConfig() Config {
	return Config{
		Env:               "local",
		ServiceName:       "controller",
		Version:           "0.1.0",
		APIVersion:        "v1",
		LogLevel:          "info",
		LogPseudonymKey:   "local-development-log-pseudonym-key-change-before-production",
		SettingsSecretKey: "local-development-system-settings-encryption-key-change-before-production",
		ShutdownTimeout:   10 * time.Second,
		HTTP: HTTPConfig{
			Addr:              ":8080",
			RequestTimeout:    10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "netstamp",
			Password:        "netstamp",
			Name:            "netstamp",
			SSLMode:         "disable",
			MaxConns:        10,
			MinConns:        0,
			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
		},
		Auth: AuthConfig{
			SessionHashKey:       "local-development-session-hash-key-change-before-production",
			SessionIdleTTL:       24 * time.Hour,
			SessionAbsoluteTTL:   7 * 24 * time.Hour,
			SessionTouchInterval: 5 * time.Minute,
			RegistrationEnabled:  true,
			Argon2idMemoryKiB:    64 * 1024,
			Argon2idIterations:   3,
			Argon2idParallelism:  4,
		},
		Tracing: TracingConfig{},
		AssignmentRefresh: AssignmentRefreshConfig{
			WorkerEnabled:      true,
			WorkerInterval:     5 * time.Second,
			WorkerBatchSize:    25,
			WorkerStaleTimeout: time.Minute,
		},
		Alerting: AlertingConfig{
			EvaluationEnabled:              true,
			NotificationWorkerEnabled:      true,
			NotificationWorkerInterval:     5 * time.Second,
			NotificationWorkerBatchSize:    25,
			NotificationWorkerStaleTimeout: time.Minute,
			NotificationHTTPTimeout:        10 * time.Second,
			SMTP: SMTPConfig{
				Port:    587,
				TLSMode: "starttls",
				Timeout: 10 * time.Second,
			},
		},
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for key := range defaultSettings {
		t.Setenv(key, "")
	}
}
