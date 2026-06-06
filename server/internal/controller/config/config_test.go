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
	if cfg.HTTP.BackendBaseURL != "" {
		t.Fatalf("expected empty backend base URL, got %q", cfg.HTTP.BackendBaseURL)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("expected default HTTP addr, got %q", cfg.HTTP.Addr)
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
}

func TestLoadFromEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv(keyAppEnv, "production")
	t.Setenv(keyServiceName, "netstamp-worker")
	t.Setenv(keyAppVersion, "0.2.0")
	t.Setenv(keyAPIVersion, "v2")
	t.Setenv(keyLogPseudonymKey, "production-log-pseudonym-key")
	t.Setenv(keyBackendBaseURL, "https://api.netstamp.dev")
	t.Setenv(keyHTTPAddr, ":8181")
	t.Setenv(keyHTTPTrustedProxies, "10.0.0.0/8,127.0.0.1")
	t.Setenv(keyRequestTimeout, "250ms")
	t.Setenv(keyDatabaseHost, "db.internal")
	t.Setenv(keyDatabasePort, "15432")
	t.Setenv(keyDatabaseUser, "netstamp_user")
	t.Setenv(keyDatabasePassword, "secret")
	t.Setenv(keyDatabaseName, "netstamp_prod")
	t.Setenv(keyDatabaseSSLMode, "require")
	t.Setenv(keyDBMaxConns, "12")
	t.Setenv(keyOTLPTracesEndpoint, "http://victoria-traces:10428/insert/opentelemetry/v1/traces")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.Env)
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
	if cfg.HTTP.BackendBaseURL != "https://api.netstamp.dev" {
		t.Fatalf("expected backend base URL override, got %q", cfg.HTTP.BackendBaseURL)
	}
	if cfg.HTTP.Addr != ":8181" {
		t.Fatalf("expected HTTP addr override, got %q", cfg.HTTP.Addr)
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
	if cfg.Tracing.OTLPTracesEndpoint != "http://victoria-traces:10428/insert/opentelemetry/v1/traces" {
		t.Fatalf("expected OTLP traces endpoint override, got %q", cfg.Tracing.OTLPTracesEndpoint)
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
	t.Setenv(keyBackendBaseURL, "https://api.netstamp.dev/api")
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
	cfg.ShutdownTimeout = 0
	cfg.HTTP.BackendBaseURL = "https://api.netstamp.dev/api"
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
		{name: "https origin", value: "https://api.netstamp.dev"},
		{name: "trailing slash", value: "https://api.netstamp.dev/"},
		{name: "missing scheme", value: "api.netstamp.dev", wantError: "BACKEND_BASE_URL must be a valid HTTP origin"},
		{name: "unsupported scheme", value: "ftp://api.netstamp.dev", wantError: "BACKEND_BASE_URL must use http or https"},
		{name: "path", value: "https://api.netstamp.dev/api", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "query", value: "https://api.netstamp.dev?preview=true", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "fragment", value: "https://api.netstamp.dev#api", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
		{name: "credentials", value: "https://user:pass@api.netstamp.dev", wantError: "BACKEND_BASE_URL must be an origin without path, query, fragment, or credentials"},
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
		Env:             "local",
		ServiceName:     "controller",
		Version:         "0.1.0",
		APIVersion:      "v1",
		LogLevel:        "info",
		LogPseudonymKey: "local-development-log-pseudonym-key-change-before-production",
		ShutdownTimeout: 10 * time.Second,
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
			JWTSecret:           "local-development-jwt-secret-change-before-production",
			AccessTokenTTL:      12 * time.Hour,
			Argon2idMemoryKiB:   64 * 1024,
			Argon2idIterations:  3,
			Argon2idParallelism: 4,
		},
		Tracing: TracingConfig{},
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for key := range defaultSettings {
		t.Setenv(key, "")
	}
}
