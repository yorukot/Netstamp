package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestConfigValidateRequiresCoreEnv(t *testing.T) {
	t.Parallel()

	cfg := Config{
		ControllerURL:       "http://localhost:8080",
		APIVersion:          DefaultAPIVersion,
		ProbeID:             "11111111-1111-1111-1111-111111111111",
		ProbeSecret:         "secret",
		HTTPTimeout:         time.Second,
		MaxWorkers:          1,
		ResultQueueSize:     1,
		ResultBatchSize:     1,
		ResultFlushInterval: time.Second,
		AssignmentTTL:       time.Minute,
		ShutdownTimeout:     time.Second,
		LogLevel:            slog.LevelInfo,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to be valid: %v", err)
	}

	cfg.ProbeSecret = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing probe secret to fail")
	}
}

func TestWorkerQueueCapacityUsesMaxWorkers(t *testing.T) {
	t.Parallel()

	if got := WorkerQueueCapacity(4); got != 8 {
		t.Fatalf("expected capacity 8, got %d", got)
	}
	if got := WorkerQueueCapacity(0); got != 1 {
		t.Fatalf("expected minimum capacity 1, got %d", got)
	}
}

func TestLoadConfigRuntimeOverrides(t *testing.T) {
	t.Setenv("NETSTAMP_PROBE_CONTROLLER_URL", "http://localhost:8080")
	t.Setenv("NETSTAMP_PROBE_ID", "11111111-1111-1111-1111-111111111111")
	t.Setenv("NETSTAMP_PROBE_SECRET", "secret")
	t.Setenv("NETSTAMP_PROBE_HEARTBEAT_INTERVAL", "7s")
	t.Setenv("NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL", "11s")
	t.Setenv("NETSTAMP_PROBE_INITIAL_BACKOFF", "2s")
	t.Setenv("NETSTAMP_PROBE_MAX_BACKOFF", "9s")
	t.Setenv("NETSTAMP_PROBE_MAX_ATTEMPTS", "3")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	assertDurationValue(t, "heartbeat", cfg.Runtime.HeartbeatInterval, 7*time.Second)
	assertDurationValue(t, "assignment poll", cfg.Runtime.AssignmentPollInterval, 11*time.Second)
	assertDurationValue(t, "initial backoff", cfg.Runtime.InitialBackoff, 2*time.Second)
	assertDurationValue(t, "max backoff", cfg.Runtime.MaxBackoff, 9*time.Second)
	if !cfg.Runtime.MaxAttempts.Defined || cfg.Runtime.MaxAttempts.Value != 3 {
		t.Fatalf("expected max attempts override to be defined with value 3, got %#v", cfg.Runtime.MaxAttempts)
	}
}

func TestConfigRejectsInvalidRuntimeBackoffOverrides(t *testing.T) {
	t.Parallel()

	cfg := Config{
		ControllerURL:       "http://localhost:8080",
		APIVersion:          DefaultAPIVersion,
		ProbeID:             "11111111-1111-1111-1111-111111111111",
		ProbeSecret:         "secret",
		HTTPTimeout:         time.Second,
		MaxWorkers:          1,
		ResultQueueSize:     1,
		ResultBatchSize:     1,
		ResultFlushInterval: time.Second,
		AssignmentTTL:       time.Minute,
		ShutdownTimeout:     time.Second,
		LogLevel:            slog.LevelInfo,
		Runtime: RuntimeConfigOverrides{
			InitialBackoff: ConfigValue[time.Duration]{Value: 10 * time.Second, Defined: true},
			MaxBackoff:     ConfigValue[time.Duration]{Value: 5 * time.Second, Defined: true},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid runtime backoff overrides to fail")
	}
}

func assertDurationValue(t *testing.T, name string, got ConfigValue[time.Duration], want time.Duration) {
	t.Helper()
	if !got.Defined || got.Value != want {
		t.Fatalf("expected %s override to be defined with value %s, got %#v", name, want, got)
	}
}
