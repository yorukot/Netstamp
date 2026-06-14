package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	keyAppEnv                         = "APP_ENV"
	keyDemoMode                       = "DEMO_MODE"
	keyServiceName                    = "SERVICE_NAME"
	keyAppVersion                     = "APP_VERSION"
	keyAPIVersion                     = "API_VERSION"
	keyLogLevel                       = "LOG_LEVEL"
	keyLogPseudonymKey                = "LOG_PSEUDONYM_KEY"
	keyShutdownTimeout                = "SHUTDOWN_TIMEOUT"
	keyBackendBaseURL                 = "BACKEND_BASE_URL"
	keyHTTPAddr                       = "HTTP_ADDR"
	keyWebDir                         = "WEB_DIR"
	keyRequestTimeout                 = "REQUEST_TIMEOUT"
	keyHTTPReadHeaderTimeout          = "HTTP_READ_HEADER_TIMEOUT"
	keyHTTPReadTimeout                = "HTTP_READ_TIMEOUT"
	keyHTTPWriteTimeout               = "HTTP_WRITE_TIMEOUT" //nolint:gosec // This is a timeout env key, not a credential.
	keyHTTPIdleTimeout                = "HTTP_IDLE_TIMEOUT"
	keyHTTPTrustedProxies             = "HTTP_TRUSTED_PROXIES"
	keyDatabaseHost                   = "DATABASE_HOST"
	keyDatabasePort                   = "DATABASE_PORT"
	keyDatabaseUser                   = "DATABASE_USER"
	keyDatabasePassword               = "DATABASE_PASSWORD"
	keyDatabaseName                   = "DATABASE_NAME"
	keyDatabaseSSLMode                = "DATABASE_SSLMODE"
	keyDBMaxConns                     = "DB_MAX_CONNS"
	keyDBMinConns                     = "DB_MIN_CONNS"
	keyDBMaxConnLifetime              = "DB_MAX_CONN_LIFETIME"
	keyDBMaxConnIdleTime              = "DB_MAX_CONN_IDLE_TIME"
	keyAuthJWTSecret                  = "AUTH_JWT_SECRET"       //nolint:gosec // This is the env key name, not the secret value.
	keyAuthAccessTokenTTL             = "AUTH_ACCESS_TOKEN_TTL" //nolint:gosec // This is a token TTL env key, not a credential.
	keyAuthRegistrationEnabled        = "AUTH_REGISTRATION_ENABLED"
	keyAuthArgon2idMemoryKiB          = "AUTH_ARGON2ID_MEMORY_KIB"
	keyAuthArgon2idIter               = "AUTH_ARGON2ID_ITERATIONS"
	keyAuthArgon2idParallel           = "AUTH_ARGON2ID_PARALLELISM"
	keyOTLPTracesEndpoint             = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	keyAlertEvaluationEnabled         = "ALERT_EVALUATION_ENABLED"
	keyNotificationWorkerEnabled      = "NOTIFICATION_WORKER_ENABLED"
	keyNotificationWorkerInterval     = "NOTIFICATION_WORKER_INTERVAL"
	keyNotificationWorkerBatchSize    = "NOTIFICATION_WORKER_BATCH_SIZE"
	keyNotificationWorkerStaleTimeout = "NOTIFICATION_WORKER_STALE_TIMEOUT"
	keyNotificationHTTPTimeout        = "NOTIFICATION_HTTP_TIMEOUT"
)

var defaultSettings = map[string]any{
	keyAppEnv:                         "local",
	keyDemoMode:                       false,
	keyServiceName:                    "controller",
	keyAppVersion:                     "0.1.0",
	keyAPIVersion:                     "v1",
	keyLogLevel:                       "info",
	keyLogPseudonymKey:                "local-development-log-pseudonym-key-change-before-production",
	keyShutdownTimeout:                10 * time.Second,
	keyBackendBaseURL:                 "",
	keyHTTPAddr:                       ":8080",
	keyWebDir:                         "",
	keyRequestTimeout:                 10 * time.Second,
	keyHTTPReadHeaderTimeout:          5 * time.Second,
	keyHTTPReadTimeout:                15 * time.Second,
	keyHTTPWriteTimeout:               15 * time.Second,
	keyHTTPIdleTimeout:                60 * time.Second,
	keyHTTPTrustedProxies:             "",
	keyDatabaseHost:                   "localhost",
	keyDatabasePort:                   int32(5432),
	keyDatabaseUser:                   "netstamp",
	keyDatabasePassword:               "netstamp",
	keyDatabaseName:                   "netstamp",
	keyDatabaseSSLMode:                "disable",
	keyDBMaxConns:                     int32(10),
	keyDBMinConns:                     int32(0),
	keyDBMaxConnLifetime:              time.Hour,
	keyDBMaxConnIdleTime:              30 * time.Minute,
	keyAuthJWTSecret:                  "local-development-jwt-secret-change-before-production",
	keyAuthAccessTokenTTL:             12 * time.Hour,
	keyAuthRegistrationEnabled:        true,
	keyAuthArgon2idMemoryKiB:          uint32(64 * 1024),
	keyAuthArgon2idIter:               uint32(3),
	keyAuthArgon2idParallel:           uint8(4),
	keyOTLPTracesEndpoint:             "",
	keyAlertEvaluationEnabled:         true,
	keyNotificationWorkerEnabled:      true,
	keyNotificationWorkerInterval:     5 * time.Second,
	keyNotificationWorkerBatchSize:    int32(25),
	keyNotificationWorkerStaleTimeout: time.Minute,
	keyNotificationHTTPTimeout:        10 * time.Second,
}

type Config struct {
	Env             string         `mapstructure:"APP_ENV"`
	DemoMode        bool           `mapstructure:"DEMO_MODE"`
	ServiceName     string         `mapstructure:"SERVICE_NAME"`
	Version         string         `mapstructure:"APP_VERSION"`
	APIVersion      string         `mapstructure:"API_VERSION"`
	LogLevel        string         `mapstructure:"LOG_LEVEL"`
	LogPseudonymKey string         `mapstructure:"LOG_PSEUDONYM_KEY"`
	ShutdownTimeout time.Duration  `mapstructure:"SHUTDOWN_TIMEOUT"`
	HTTP            HTTPConfig     `mapstructure:",squash"`
	Database        DatabaseConfig `mapstructure:",squash"`
	Auth            AuthConfig     `mapstructure:",squash"`
	Tracing         TracingConfig  `mapstructure:",squash"`
	Alerting        AlertingConfig `mapstructure:",squash"`
}

type HTTPConfig struct {
	BackendBaseURL    string        `mapstructure:"BACKEND_BASE_URL"`
	Addr              string        `mapstructure:"HTTP_ADDR"`
	WebDir            string        `mapstructure:"WEB_DIR"`
	RequestTimeout    time.Duration `mapstructure:"REQUEST_TIMEOUT"`
	ReadHeaderTimeout time.Duration `mapstructure:"HTTP_READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `mapstructure:"HTTP_READ_TIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"HTTP_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `mapstructure:"HTTP_IDLE_TIMEOUT"`
	TrustedProxies    string        `mapstructure:"HTTP_TRUSTED_PROXIES"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"DATABASE_HOST"`
	Port            int32         `mapstructure:"DATABASE_PORT"`
	User            string        `mapstructure:"DATABASE_USER"`
	Password        string        `mapstructure:"DATABASE_PASSWORD"`
	Name            string        `mapstructure:"DATABASE_NAME"`
	SSLMode         string        `mapstructure:"DATABASE_SSLMODE"`
	MaxConns        int32         `mapstructure:"DB_MAX_CONNS"`
	MinConns        int32         `mapstructure:"DB_MIN_CONNS"`
	MaxConnLifetime time.Duration `mapstructure:"DB_MAX_CONN_LIFETIME"`
	MaxConnIdleTime time.Duration `mapstructure:"DB_MAX_CONN_IDLE_TIME"`
}

type AuthConfig struct {
	JWTSecret           string        `mapstructure:"AUTH_JWT_SECRET"`
	AccessTokenTTL      time.Duration `mapstructure:"AUTH_ACCESS_TOKEN_TTL"`
	RegistrationEnabled bool          `mapstructure:"AUTH_REGISTRATION_ENABLED"`
	Argon2idMemoryKiB   uint32        `mapstructure:"AUTH_ARGON2ID_MEMORY_KIB"`
	Argon2idIterations  uint32        `mapstructure:"AUTH_ARGON2ID_ITERATIONS"`
	Argon2idParallelism uint8         `mapstructure:"AUTH_ARGON2ID_PARALLELISM"`
}

type TracingConfig struct {
	OTLPTracesEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"`
}

type AlertingConfig struct {
	EvaluationEnabled              bool          `mapstructure:"ALERT_EVALUATION_ENABLED"`
	NotificationWorkerEnabled      bool          `mapstructure:"NOTIFICATION_WORKER_ENABLED"`
	NotificationWorkerInterval     time.Duration `mapstructure:"NOTIFICATION_WORKER_INTERVAL"`
	NotificationWorkerBatchSize    int32         `mapstructure:"NOTIFICATION_WORKER_BATCH_SIZE"`
	NotificationWorkerStaleTimeout time.Duration `mapstructure:"NOTIFICATION_WORKER_STALE_TIMEOUT"`
	NotificationHTTPTimeout        time.Duration `mapstructure:"NOTIFICATION_HTTP_TIMEOUT"`
}

func (cfg DatabaseConfig) ConnectionString() string {
	databaseURL := url.URL{
		Scheme: "postgres",
		User:   url.User(cfg.User),
		Host:   net.JoinHostPort(cfg.Host, strconv.FormatInt(int64(cfg.Port), 10)),
		Path:   cfg.Name,
	}
	if cfg.Password != "" {
		databaseURL.User = url.UserPassword(cfg.User, cfg.Password)
	}

	query := databaseURL.Query()
	query.Set("sslmode", cfg.SSLMode)
	databaseURL.RawQuery = query.Encode()

	return databaseURL.String()
}

func (cfg HTTPConfig) TrustedProxyPrefixes() ([]netip.Prefix, error) {
	return parseTrustedProxyPrefixes(cfg.TrustedProxies)
}

func Load() (Config, error) {
	settings, err := newSettings()
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	var errs []error
	if err := settings.UnmarshalExact(&cfg); err != nil {
		errs = append(errs, fmt.Errorf("decode config: %w", err))
	}

	errs = append(errs, validate(cfg)...)
	return cfg, errors.Join(errs...)
}

func validate(cfg Config) []error {
	var errs []error

	// Global settings
	errs = append(errs, validateRequiredString(keyAppEnv, cfg.Env)...)
	errs = append(errs, validateRequiredString(keyServiceName, cfg.ServiceName)...)
	errs = append(errs, validateRequiredString(keyAppVersion, cfg.Version)...)
	errs = append(errs, validateAPIVersion(cfg.APIVersion)...)
	errs = append(errs, validateLogLevel(cfg.LogLevel)...)
	errs = append(errs, validateRequiredString(keyLogPseudonymKey, cfg.LogPseudonymKey)...)
	errs = append(errs, validatePositiveDuration(keyShutdownTimeout, cfg.ShutdownTimeout)...)

	// HTTP settings
	errs = append(errs, validateOptionalHTTPOrigin(keyBackendBaseURL, cfg.HTTP.BackendBaseURL)...)
	errs = append(errs, validateListenAddr(keyHTTPAddr, cfg.HTTP.Addr)...)
	errs = append(errs, validatePositiveDuration(keyRequestTimeout, cfg.HTTP.RequestTimeout)...)
	errs = append(errs, validatePositiveDuration(keyHTTPReadHeaderTimeout, cfg.HTTP.ReadHeaderTimeout)...)
	errs = append(errs, validatePositiveDuration(keyHTTPReadTimeout, cfg.HTTP.ReadTimeout)...)
	errs = append(errs, validatePositiveDuration(keyHTTPWriteTimeout, cfg.HTTP.WriteTimeout)...)
	errs = append(errs, validatePositiveDuration(keyHTTPIdleTimeout, cfg.HTTP.IdleTimeout)...)
	errs = append(errs, validateTrustedProxyPrefixes(keyHTTPTrustedProxies, cfg.HTTP.TrustedProxies)...)

	// Database settings
	errs = append(errs, validateRequiredString(keyDatabaseHost, cfg.Database.Host)...)
	errs = append(errs, validateRequiredString(keyDatabaseUser, cfg.Database.User)...)
	errs = append(errs, validateRequiredString(keyDatabaseName, cfg.Database.Name)...)
	errs = append(errs, validateDatabasePort(cfg.Database.Port)...)
	errs = append(errs, validateDatabaseSSLMode(cfg.Database.SSLMode)...)

	if cfg.Database.MaxConns < 0 {
		errs = append(errs, errors.New("DB_MAX_CONNS must not be negative"))
	}
	if cfg.Database.MinConns < 0 {
		errs = append(errs, errors.New("DB_MIN_CONNS must not be negative"))
	}
	if cfg.Database.MaxConns == 0 {
		errs = append(errs, errors.New("DB_MAX_CONNS must be greater than 0"))
	}
	if cfg.Database.MinConns > cfg.Database.MaxConns {
		errs = append(errs, errors.New("DB_MIN_CONNS must not be greater than DB_MAX_CONNS"))
	}
	errs = append(errs, validatePositiveDuration(keyDBMaxConnLifetime, cfg.Database.MaxConnLifetime)...)
	errs = append(errs, validatePositiveDuration(keyDBMaxConnIdleTime, cfg.Database.MaxConnIdleTime)...)

	// Auth settings
	errs = append(errs, validateRequiredString(keyAuthJWTSecret, cfg.Auth.JWTSecret)...)
	errs = append(errs, validatePositiveDuration(keyAuthAccessTokenTTL, cfg.Auth.AccessTokenTTL)...)
	errs = append(errs, validatePositiveUint32(keyAuthArgon2idMemoryKiB, cfg.Auth.Argon2idMemoryKiB)...)
	errs = append(errs, validatePositiveUint32(keyAuthArgon2idIter, cfg.Auth.Argon2idIterations)...)
	errs = append(errs, validatePositiveUint8(keyAuthArgon2idParallel, cfg.Auth.Argon2idParallelism)...)

	// Tracing settings
	errs = append(errs, validateOptionalHTTPURL(keyOTLPTracesEndpoint, cfg.Tracing.OTLPTracesEndpoint)...)

	// Alerting settings
	errs = append(errs, validatePositiveDuration(keyNotificationWorkerInterval, cfg.Alerting.NotificationWorkerInterval)...)
	errs = append(errs, validatePositiveDuration(keyNotificationWorkerStaleTimeout, cfg.Alerting.NotificationWorkerStaleTimeout)...)
	errs = append(errs, validatePositiveDuration(keyNotificationHTTPTimeout, cfg.Alerting.NotificationHTTPTimeout)...)
	if cfg.Alerting.NotificationWorkerBatchSize <= 0 {
		errs = append(errs, errors.New("NOTIFICATION_WORKER_BATCH_SIZE must be greater than 0"))
	}

	return errs
}

func newSettings() (*viper.Viper, error) {
	settings := viper.New()
	settings.SetConfigName(".env")
	settings.SetConfigType("env")
	settings.AddConfigPath(".")
	settings.AddConfigPath("server")
	settings.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	settings.AutomaticEnv()

	var errs []error
	for key, value := range defaultSettings {
		settings.SetDefault(key, value)
		if err := settings.BindEnv(key); err != nil {
			errs = append(errs, fmt.Errorf("bind %s: %w", key, err))
		}
	}

	if err := loadDotEnv(settings); err != nil {
		errs = append(errs, err)
	}

	return settings, errors.Join(errs...)
}

func loadDotEnv(settings *viper.Viper) error {
	if err := settings.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("read .env: %w", err)
	}

	return nil
}
