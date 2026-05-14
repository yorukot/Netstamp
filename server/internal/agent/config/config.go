package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/netstamp/internal/domain/probe"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	DefaultHTTPTimeout         = 10 * time.Second
	DefaultMaxWorkers          = 256
	DefaultResultQueueSize     = 10000
	DefaultResultBatchSize     = 100
	DefaultResultFlushInterval = 5 * time.Second
	DefaultAssignmentTTL       = 10 * time.Minute
	DefaultShutdownTimeout     = 10 * time.Second
	DefaultAPIVersion          = "v1"
	DefaultLogLevel            = slog.LevelInfo
)

type Config struct {
	ControllerURL string
	APIVersion    string
	ProbeID       string
	ProbeSecret   string

	// ConfigValue tracks whether a value came from the environment.
	// Undefined values may later be supplied by the controller.
	HTTPTimeout            ConfigValue[time.Duration]
	MaxWorkers             ConfigValue[int]
	ResultQueueSize        ConfigValue[int]
	ResultBatchSize        ConfigValue[int]
	ResultFlushInterval    ConfigValue[time.Duration]
	AssignmentTTL          ConfigValue[time.Duration]
	ShutdownTimeout        ConfigValue[time.Duration]
	HeartbeatInterval      ConfigValue[time.Duration]
	AssignmentPollInterval ConfigValue[time.Duration]
	InitialBackoff         ConfigValue[time.Duration]
	MaxBackoff             ConfigValue[time.Duration]
	MaxAttempts            ConfigValue[int]
	LogLevel               slog.Level
}

type ConfigValue[T any] struct {
	Value   T
	Defined bool
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ControllerURL: strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_CONTROLLER_URL")),
		APIVersion:    DefaultAPIVersion,
		ProbeID:       strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_ID")),
		ProbeSecret:   strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_SECRET")),
		LogLevel:      DefaultLogLevel,
	}

	var err error

	if cfg.HTTPTimeout, err = envValue(
		"NETSTAMP_PROBE_HTTP_TIMEOUT",
		DefaultHTTPTimeout,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.MaxWorkers, err = envValue(
		"NETSTAMP_PROBE_MAX_WORKERS",
		DefaultMaxWorkers,
		parseInt,
		positiveInt,
	); err != nil {
		return Config{}, err
	}

	if cfg.ResultQueueSize, err = envValue(
		"NETSTAMP_PROBE_RESULT_QUEUE_SIZE",
		DefaultResultQueueSize,
		parseInt,
		positiveInt,
	); err != nil {
		return Config{}, err
	}

	if cfg.ResultBatchSize, err = envValue(
		"NETSTAMP_PROBE_RESULT_BATCH_SIZE",
		DefaultResultBatchSize,
		parseInt,
		positiveInt,
	); err != nil {
		return Config{}, err
	}

	if cfg.ResultFlushInterval, err = envValue(
		"NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL",
		DefaultResultFlushInterval,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.AssignmentTTL, err = envValue(
		"NETSTAMP_PROBE_ASSIGNMENT_TTL",
		DefaultAssignmentTTL,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.ShutdownTimeout, err = envValue(
		"NETSTAMP_PROBE_SHUTDOWN_TIMEOUT",
		DefaultShutdownTimeout,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.HeartbeatInterval, err = envValue(
		"NETSTAMP_PROBE_HEARTBEAT_INTERVAL",
		time.Duration(probe.DefaultRuntimeHeartbeatIntervalSeconds)*time.Second,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.AssignmentPollInterval, err = envValue(
		"NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL",
		time.Duration(probe.DefaultRuntimeAssignmentPollIntervalSeconds)*time.Second,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.InitialBackoff, err = envValue(
		"NETSTAMP_PROBE_INITIAL_BACKOFF",
		time.Duration(probe.DefaultRuntimeInitialBackoffSeconds)*time.Second,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.MaxBackoff, err = envValue(
		"NETSTAMP_PROBE_MAX_BACKOFF",
		time.Duration(probe.DefaultRuntimeMaxBackoffSeconds)*time.Second,
		parseDuration,
		positiveDuration,
	); err != nil {
		return Config{}, err
	}

	if cfg.MaxAttempts, err = envValue(
		"NETSTAMP_PROBE_MAX_ATTEMPTS",
		int(probe.DefaultRuntimeMaxAttempts),
		parseInt,
		positiveInt,
	); err != nil {
		return Config{}, err
	}

	if cfg.LogLevel, err = parseOptionalLogLevel(
		"NETSTAMP_PROBE_LOG_LEVEL",
		DefaultLogLevel,
	); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.ControllerURL == "" {
		return fmt.Errorf("NETSTAMP_PROBE_CONTROLLER_URL is required")
	}

	controllerURL, err := url.Parse(c.ControllerURL)
	if err != nil {
		return fmt.Errorf("NETSTAMP_PROBE_CONTROLLER_URL is invalid: %w", err)
	}
	if controllerURL.Scheme != "http" && controllerURL.Scheme != "https" {
		return fmt.Errorf("NETSTAMP_PROBE_CONTROLLER_URL must use http or https")
	}
	if controllerURL.Host == "" {
		return fmt.Errorf("NETSTAMP_PROBE_CONTROLLER_URL must include a host")
	}

	if _, err := domainprobe.VNProbeID(c.ProbeID); err != nil {
		return fmt.Errorf("NETSTAMP_PROBE_ID is invalid: %w", err)
	}

	if c.ProbeSecret == "" {
		return fmt.Errorf("NETSTAMP_PROBE_SECRET is required")
	}

	if c.HTTPTimeout.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_HTTP_TIMEOUT must be greater than zero")
	}
	if c.MaxWorkers.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_MAX_WORKERS must be greater than zero")
	}
	if c.ResultQueueSize.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_QUEUE_SIZE must be greater than zero")
	}
	if c.ResultBatchSize.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_BATCH_SIZE must be greater than zero")
	}
	if c.ResultFlushInterval.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL must be greater than zero")
	}
	if c.AssignmentTTL.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_ASSIGNMENT_TTL must be greater than zero")
	}
	if c.ShutdownTimeout.Value <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_SHUTDOWN_TIMEOUT must be greater than zero")
	}

	if c.InitialBackoff.Defined && c.MaxBackoff.Defined && c.MaxBackoff.Value < c.InitialBackoff.Value {
		return fmt.Errorf("NETSTAMP_PROBE_MAX_BACKOFF must be greater than or equal to NETSTAMP_PROBE_INITIAL_BACKOFF")
	}

	return nil
}

func envValue[T any](
	name string,
	fallback T,
	parse func(string) (T, error),
	validate func(T) error,
) (ConfigValue[T], error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return ConfigValue[T]{
			Value:   fallback,
			Defined: false,
		}, nil
	}

	value, err := parse(raw)
	if err != nil {
		return ConfigValue[T]{}, fmt.Errorf("%s is invalid: %w", name, err)
	}

	if validate != nil {
		if err := validate(value); err != nil {
			return ConfigValue[T]{}, fmt.Errorf("%s %w", name, err)
		}
	}

	return ConfigValue[T]{
		Value:   value,
		Defined: true,
	}, nil
}

func parseDuration(raw string) (time.Duration, error) {
	return time.ParseDuration(raw)
}

func parseInt(raw string) (int, error) {
	return strconv.Atoi(raw)
}

func positiveDuration(value time.Duration) error {
	if value <= 0 {
		return fmt.Errorf("must be greater than zero")
	}
	return nil
}

func positiveInt(value int) error {
	if value <= 0 {
		return fmt.Errorf("must be greater than zero")
	}
	return nil
}

func parseOptionalLogLevel(name string, fallback slog.Level) (slog.Level, error) {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	if raw == "" {
		return fallback, nil
	}

	switch raw {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%s must be one of debug, info, warn, or error", name)
	}
}

func (c Config) SafeLogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("controller_url", c.ControllerURL),
		slog.String("api_version", c.APIVersion),
		slog.String("probe_id", c.ProbeID),
		slog.Duration("http_timeout", c.HTTPTimeout.Value),
		slog.Int("max_workers", c.MaxWorkers.Value),
		slog.Int("result_queue_size", c.ResultQueueSize.Value),
		slog.Int("result_batch_size", c.ResultBatchSize.Value),
		slog.Duration("result_flush_interval", c.ResultFlushInterval.Value),
		slog.Duration("assignment_ttl", c.AssignmentTTL.Value),
		slog.Duration("shutdown_timeout", c.ShutdownTimeout.Value),
		slog.Bool("heartbeat_interval_env_defined", c.HeartbeatInterval.Defined),
		slog.Bool("assignment_poll_interval_env_defined", c.AssignmentPollInterval.Defined),
		slog.Bool("initial_backoff_env_defined", c.InitialBackoff.Defined),
		slog.Bool("max_backoff_env_defined", c.MaxBackoff.Defined),
		slog.Bool("max_attempts_env_defined", c.MaxAttempts.Defined),
	}
}
