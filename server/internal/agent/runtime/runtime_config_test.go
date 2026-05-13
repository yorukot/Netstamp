package runtime

import (
	"testing"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func TestRuntimeConfigStoreAppliesControllerWhenEnvUnset(t *testing.T) {
	t.Parallel()

	store := NewRuntimeConfigStore(agentconfig.RuntimeConfigOverrides{})
	got := store.ApplyController(domainprobe.RuntimeConfig{
		HeartbeatIntervalSeconds:      7,
		AssignmentPollIntervalSeconds: 11,
		InitialBackoffSeconds:         2,
		MaxBackoffSeconds:             9,
		MaxAttempts:                   3,
	})

	if got.HeartbeatInterval != 7*time.Second {
		t.Fatalf("expected controller heartbeat interval, got %s", got.HeartbeatInterval)
	}
	if got.AssignmentPollInterval != 11*time.Second {
		t.Fatalf("expected controller assignment poll interval, got %s", got.AssignmentPollInterval)
	}
	if got.InitialBackoff != 2*time.Second || got.MaxBackoff != 9*time.Second || got.MaxAttempts != 3 {
		t.Fatalf("expected controller retry config, got %#v", got)
	}
}

func TestRuntimeConfigStorePreservesEnvDefinedValues(t *testing.T) {
	t.Parallel()

	store := NewRuntimeConfigStore(agentconfig.RuntimeConfigOverrides{
		HeartbeatInterval:      agentconfig.ConfigValue[time.Duration]{Value: 5 * time.Second, Defined: true},
		AssignmentPollInterval: agentconfig.ConfigValue[time.Duration]{Value: 6 * time.Second, Defined: true},
		InitialBackoff:         agentconfig.ConfigValue[time.Duration]{Value: 3 * time.Second, Defined: true},
		MaxBackoff:             agentconfig.ConfigValue[time.Duration]{Value: 12 * time.Second, Defined: true},
		MaxAttempts:            agentconfig.ConfigValue[int]{Value: 8, Defined: true},
	})

	got := store.ApplyController(domainprobe.RuntimeConfig{
		HeartbeatIntervalSeconds:      70,
		AssignmentPollIntervalSeconds: 80,
		InitialBackoffSeconds:         20,
		MaxBackoffSeconds:             90,
		MaxAttempts:                   30,
	})

	if got.HeartbeatInterval != 5*time.Second ||
		got.AssignmentPollInterval != 6*time.Second ||
		got.InitialBackoff != 3*time.Second ||
		got.MaxBackoff != 12*time.Second ||
		got.MaxAttempts != 8 {
		t.Fatalf("expected env-defined values to be preserved, got %#v", got)
	}
}

func TestRuntimeConfigStoreDefaultsZeroControllerValues(t *testing.T) {
	t.Parallel()

	store := NewRuntimeConfigStore(agentconfig.RuntimeConfigOverrides{})
	got := store.ApplyController(domainprobe.RuntimeConfig{})
	want := DefaultRuntimeConfig()
	if got != want {
		t.Fatalf("expected zero controller config to default to %#v, got %#v", want, got)
	}
}

func TestRuntimeConfigStoreNormalizesBackoffWithEnvLocks(t *testing.T) {
	t.Parallel()

	store := NewRuntimeConfigStore(agentconfig.RuntimeConfigOverrides{
		MaxBackoff: agentconfig.ConfigValue[time.Duration]{Value: 5 * time.Second, Defined: true},
	})
	got := store.ApplyController(domainprobe.RuntimeConfig{
		InitialBackoffSeconds: 10,
		MaxBackoffSeconds:     30,
	})

	if got.InitialBackoff != 5*time.Second || got.MaxBackoff != 5*time.Second {
		t.Fatalf("expected env max backoff to cap controller initial backoff, got %#v", got)
	}
}
