package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAgentVersion = "netstamp-probe/0.1.0"
	DefaultHTTPTimeout  = 10 * time.Second
	DefaultMaxWorkers   = 16

	envControllerURL = "NETSTAMP_PROBE_CONTROLLER_URL"
	envProbeID       = "NETSTAMP_PROBE_ID"
	// #nosec G101 -- This is the environment variable name, not a credential value.
	envProbeSecret  = "NETSTAMP_PROBE_SECRET"
	envHTTPTimeout  = "NETSTAMP_PROBE_HTTP_TIMEOUT"
	envMaxWorkers   = "NETSTAMP_PROBE_MAX_WORKERS"
)

var ErrInvalidConfig = errors.New("probe config invalid")

type Config struct {
	ControllerURL string
	ProbeID       string
	ProbeSecret   string
	AgentVersion  string
	HTTPTimeout   time.Duration
	MaxWorkers    int
}

func Load() (Config, error) {
	cfg := Config{
		ControllerURL: strings.TrimRight(strings.TrimSpace(os.Getenv(envControllerURL)), "/"),
		ProbeID:       strings.TrimSpace(os.Getenv(envProbeID)),
		ProbeSecret:   strings.TrimSpace(os.Getenv(envProbeSecret)),
		// We do not allow user to set the default agent version
		AgentVersion:  DefaultAgentVersion,
		HTTPTimeout:   DefaultHTTPTimeout,
		MaxWorkers:    DefaultMaxWorkers,
	}

	if err := require(cfg.ControllerURL, envControllerURL); err != nil {
		return Config{}, err
	}
	if err := require(cfg.ProbeID, envProbeID); err != nil {
		return Config{}, err
	}
	if err := require(cfg.ProbeSecret, envProbeSecret); err != nil {
		return Config{}, err
	}

	if raw := strings.TrimSpace(os.Getenv(envHTTPTimeout)); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil || timeout <= 0 {
			return Config{}, fmt.Errorf("%w: %s must be a positive duration", ErrInvalidConfig, envHTTPTimeout)
		}
		cfg.HTTPTimeout = timeout
	}
	if raw := strings.TrimSpace(os.Getenv(envMaxWorkers)); raw != "" {
		workers, err := strconv.Atoi(raw)
		if err != nil || workers <= 0 {
			return Config{}, fmt.Errorf("%w: %s must be a positive integer", ErrInvalidConfig, envMaxWorkers)
		}
		cfg.MaxWorkers = workers
	}

	return cfg, nil
}

func require(value, name string) error {
	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidConfig, name)
	}

	return nil
}
