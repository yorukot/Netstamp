package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	DefaultHTTPTimeout         = 10 * time.Second
	DefaultMaxWorkers          = 16
	DefaultResultQueueSize     = 10000
	DefaultResultBatchSize     = 100
	DefaultResultFlushInterval = 5 * time.Second
	DefaultAssignmentTTL       = 10 * time.Minute
	DefaultShutdownTimeout     = 10 * time.Second
	DefaultAPIVersion          = "v1"
	DefaultLogLevel            = slog.LevelInfo
)

type Config struct {
	ControllerURL       string
	APIVersion          string
	ProbeID             string
	ProbeSecret         string
	HTTPTimeout         time.Duration
	MaxWorkers          int
	ResultQueueSize     int
	ResultBatchSize     int
	ResultFlushInterval time.Duration
	AssignmentTTL       time.Duration
	ShutdownTimeout     time.Duration
	LogLevel            slog.Level
	Runtime             RuntimeConfigOverrides
}

type ConfigValue[T any] struct {
	Value   T
	Defined bool
}

type RuntimeConfigOverrides struct {
	HeartbeatInterval      ConfigValue[time.Duration]
	AssignmentPollInterval ConfigValue[time.Duration]
	InitialBackoff         ConfigValue[time.Duration]
	MaxBackoff             ConfigValue[time.Duration]
	MaxAttempts            ConfigValue[int]
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ControllerURL:       strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_CONTROLLER_URL")),
		APIVersion:          DefaultAPIVersion,
		ProbeID:             strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_ID")),
		ProbeSecret:         strings.TrimSpace(os.Getenv("NETSTAMP_PROBE_SECRET")),
		HTTPTimeout:         DefaultHTTPTimeout,
		MaxWorkers:          DefaultMaxWorkers,
		ResultQueueSize:     DefaultResultQueueSize,
		ResultBatchSize:     DefaultResultBatchSize,
		ResultFlushInterval: DefaultResultFlushInterval,
		AssignmentTTL:       DefaultAssignmentTTL,
		ShutdownTimeout:     DefaultShutdownTimeout,
		LogLevel:            DefaultLogLevel,
	}

	var err error
	if cfg.HTTPTimeout, err = parseOptionalDuration("NETSTAMP_PROBE_HTTP_TIMEOUT", cfg.HTTPTimeout); err != nil {
		return Config{}, err
	}
	if cfg.MaxWorkers, err = parseOptionalPositiveInt("NETSTAMP_PROBE_MAX_WORKERS", cfg.MaxWorkers); err != nil {
		return Config{}, err
	}
	if cfg.ResultQueueSize, err = parseOptionalPositiveInt("NETSTAMP_PROBE_RESULT_QUEUE_SIZE", cfg.ResultQueueSize); err != nil {
		return Config{}, err
	}
	if cfg.ResultBatchSize, err = parseOptionalPositiveInt("NETSTAMP_PROBE_RESULT_BATCH_SIZE", cfg.ResultBatchSize); err != nil {
		return Config{}, err
	}
	if cfg.ResultFlushInterval, err = parseOptionalDuration("NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL", cfg.ResultFlushInterval); err != nil {
		return Config{}, err
	}
	if cfg.AssignmentTTL, err = parseOptionalDuration("NETSTAMP_PROBE_ASSIGNMENT_TTL", cfg.AssignmentTTL); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = parseOptionalDuration("NETSTAMP_PROBE_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if cfg.LogLevel, err = parseOptionalLogLevel("NETSTAMP_PROBE_LOG_LEVEL", cfg.LogLevel); err != nil {
		return Config{}, err
	}
	if cfg.Runtime.HeartbeatInterval, err = parseOptionalDurationValue("NETSTAMP_PROBE_HEARTBEAT_INTERVAL"); err != nil {
		return Config{}, err
	}
	if cfg.Runtime.AssignmentPollInterval, err = parseOptionalDurationValue("NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL"); err != nil {
		return Config{}, err
	}
	if cfg.Runtime.InitialBackoff, err = parseOptionalDurationValue("NETSTAMP_PROBE_INITIAL_BACKOFF"); err != nil {
		return Config{}, err
	}
	if cfg.Runtime.MaxBackoff, err = parseOptionalDurationValue("NETSTAMP_PROBE_MAX_BACKOFF"); err != nil {
		return Config{}, err
	}
	if cfg.Runtime.MaxAttempts, err = parseOptionalPositiveIntValue("NETSTAMP_PROBE_MAX_ATTEMPTS"); err != nil {
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
	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_HTTP_TIMEOUT must be greater than zero")
	}
	if c.MaxWorkers <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_MAX_WORKERS must be greater than zero")
	}
	if c.ResultQueueSize <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_QUEUE_SIZE must be greater than zero")
	}
	if c.ResultBatchSize <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_BATCH_SIZE must be greater than zero")
	}
	if c.ResultBatchSize > domainprobe.MaxRuntimeResultGroupBatchSize {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_BATCH_SIZE must be less than or equal to %d", domainprobe.MaxRuntimeResultGroupBatchSize)
	}
	if c.ResultFlushInterval <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL must be greater than zero")
	}
	if c.AssignmentTTL <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_ASSIGNMENT_TTL must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("NETSTAMP_PROBE_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.Runtime.InitialBackoff.Defined && c.Runtime.MaxBackoff.Defined && c.Runtime.MaxBackoff.Value < c.Runtime.InitialBackoff.Value {
		return fmt.Errorf("NETSTAMP_PROBE_MAX_BACKOFF must be greater than or equal to NETSTAMP_PROBE_INITIAL_BACKOFF")
	}

	return nil
}

func (c Config) SafeLogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("controller_url", c.ControllerURL),
		slog.String("api_version", c.APIVersion),
		slog.String("probe_id", c.ProbeID),
		slog.Duration("http_timeout", c.HTTPTimeout),
		slog.Int("max_workers", c.MaxWorkers),
		slog.Int("result_queue_size", c.ResultQueueSize),
		slog.Int("result_batch_size", c.ResultBatchSize),
		slog.Duration("result_flush_interval", c.ResultFlushInterval),
		slog.Duration("assignment_ttl", c.AssignmentTTL),
		slog.Duration("shutdown_timeout", c.ShutdownTimeout),
		slog.Bool("runtime_heartbeat_interval_env_defined", c.Runtime.HeartbeatInterval.Defined),
		slog.Bool("runtime_assignment_poll_interval_env_defined", c.Runtime.AssignmentPollInterval.Defined),
		slog.Bool("runtime_initial_backoff_env_defined", c.Runtime.InitialBackoff.Defined),
		slog.Bool("runtime_max_backoff_env_defined", c.Runtime.MaxBackoff.Defined),
		slog.Bool("runtime_max_attempts_env_defined", c.Runtime.MaxAttempts.Defined),
	}
}

func WorkerQueueCapacity(maxWorkers int) int {
	return max(1, maxWorkers*2)
}

func parseOptionalDuration(name string, fallback time.Duration) (time.Duration, error) {
	value, err := parseOptionalDurationValue(name)
	if err != nil {
		return 0, err
	}
	if !value.Defined {
		return fallback, nil
	}

	return value.Value, nil
}

func parseOptionalPositiveInt(name string, fallback int) (int, error) {
	value, err := parseOptionalPositiveIntValue(name)
	if err != nil {
		return 0, err
	}
	if !value.Defined {
		return fallback, nil
	}

	return value.Value, nil
}

func parseOptionalDurationValue(name string) (ConfigValue[time.Duration], error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return ConfigValue[time.Duration]{}, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return ConfigValue[time.Duration]{}, fmt.Errorf("%s is invalid: %w", name, err)
	}
	if value <= 0 {
		return ConfigValue[time.Duration]{}, fmt.Errorf("%s must be greater than zero", name)
	}
	return ConfigValue[time.Duration]{Value: value, Defined: true}, nil
}

func parseOptionalPositiveIntValue(name string) (ConfigValue[int], error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return ConfigValue[int]{}, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return ConfigValue[int]{}, fmt.Errorf("%s is invalid: %w", name, err)
	}
	if value <= 0 {
		return ConfigValue[int]{}, fmt.Errorf("%s must be greater than zero", name)
	}
	return ConfigValue[int]{Value: value, Defined: true}, nil
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
