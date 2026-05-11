package config

import (
	"errors"
	"testing"
	"time"
)

func TestLoadConfigFromEnvironment(t *testing.T) {
	t.Setenv(envControllerURL, " http://localhost:8080/ ")
	t.Setenv(envProbeID, "probe-id")
	t.Setenv(envProbeSecret, "secret")
	t.Setenv(envHTTPTimeout, "5s")
	t.Setenv(envMaxWorkers, "4")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.ControllerURL != "http://localhost:8080" {
		t.Fatalf("unexpected controller url: %q", cfg.ControllerURL)
	}
	if cfg.HTTPTimeout != 5*time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.HTTPTimeout)
	}
	if cfg.MaxWorkers != 4 {
		t.Fatalf("unexpected max workers: %d", cfg.MaxWorkers)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv(envControllerURL, "http://localhost:8080")
	t.Setenv(envProbeID, "probe-id")
	t.Setenv(envProbeSecret, "secret")
	t.Setenv(envHTTPTimeout, "")
	t.Setenv(envMaxWorkers, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.AgentVersion != DefaultAgentVersion {
		t.Fatalf("unexpected agent version: %q", cfg.AgentVersion)
	}
	if cfg.HTTPTimeout != DefaultHTTPTimeout {
		t.Fatalf("unexpected timeout: %s", cfg.HTTPTimeout)
	}
	if cfg.MaxWorkers != DefaultMaxWorkers {
		t.Fatalf("unexpected max workers: %d", cfg.MaxWorkers)
	}
}

func TestLoadConfigRequiresControllerURLProbeIDAndSecret(t *testing.T) {
	t.Setenv(envControllerURL, "")
	t.Setenv(envProbeID, "probe-id")
	t.Setenv(envProbeSecret, "secret")

	_, err := Load()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestLoadConfigRejectsInvalidOptionalValues(t *testing.T) {
	t.Setenv(envControllerURL, "http://localhost:8080")
	t.Setenv(envProbeID, "probe-id")
	t.Setenv(envProbeSecret, "secret")
	t.Setenv(envHTTPTimeout, "-1s")

	_, err := Load()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}
