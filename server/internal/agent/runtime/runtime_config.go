package runtime

import (
	"sync"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type RuntimeConfig struct {
	HeartbeatInterval      time.Duration
	AssignmentPollInterval time.Duration
	InitialBackoff         time.Duration
	MaxBackoff             time.Duration
	MaxAttempts            int
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfigFromDomain(domainprobe.DefaultRuntimeConfig())
}

func RuntimeConfigFromDomain(input domainprobe.RuntimeConfig) RuntimeConfig {
	cfg := RuntimeConfig{
		HeartbeatInterval:      secondsOrDefault(input.HeartbeatIntervalSeconds, domainprobe.DefaultRuntimeHeartbeatIntervalSeconds),
		AssignmentPollInterval: secondsOrDefault(input.AssignmentPollIntervalSeconds, domainprobe.DefaultRuntimeAssignmentPollIntervalSeconds),
		InitialBackoff:         secondsOrDefault(input.InitialBackoffSeconds, domainprobe.DefaultRuntimeInitialBackoffSeconds),
		MaxBackoff:             secondsOrDefault(input.MaxBackoffSeconds, domainprobe.DefaultRuntimeMaxBackoffSeconds),
		MaxAttempts:            int(input.MaxAttempts),
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = int(domainprobe.DefaultRuntimeMaxAttempts)
	}
	if cfg.MaxBackoff < cfg.InitialBackoff {
		cfg.MaxBackoff = cfg.InitialBackoff
	}

	return cfg
}

func secondsOrDefault(value, fallback int32) time.Duration {
	if value <= 0 {
		value = fallback
	}
	return time.Duration(value) * time.Second
}

type RuntimeConfigStore struct {
	mu        sync.RWMutex
	config    RuntimeConfig
	overrides agentconfig.RuntimeConfigOverrides
}

func NewRuntimeConfigStore(overrides agentconfig.RuntimeConfigOverrides) *RuntimeConfigStore {
	config := runtimeConfigWithOverrides(DefaultRuntimeConfig(), overrides)
	return &RuntimeConfigStore{
		config:    config,
		overrides: overrides,
	}
}

func (s *RuntimeConfigStore) Get() RuntimeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

func (s *RuntimeConfigStore) ApplyController(input domainprobe.RuntimeConfig) RuntimeConfig {
	config := runtimeConfigWithOverrides(RuntimeConfigFromDomain(input), s.overrides)

	s.mu.Lock()
	s.config = config
	s.mu.Unlock()

	return config
}

func runtimeConfigWithOverrides(config RuntimeConfig, overrides agentconfig.RuntimeConfigOverrides) RuntimeConfig {
	if overrides.HeartbeatInterval.Defined {
		config.HeartbeatInterval = overrides.HeartbeatInterval.Value
	}
	if overrides.AssignmentPollInterval.Defined {
		config.AssignmentPollInterval = overrides.AssignmentPollInterval.Value
	}
	if overrides.InitialBackoff.Defined {
		config.InitialBackoff = overrides.InitialBackoff.Value
	}
	if overrides.MaxBackoff.Defined {
		config.MaxBackoff = overrides.MaxBackoff.Value
	}
	if overrides.MaxAttempts.Defined {
		config.MaxAttempts = overrides.MaxAttempts.Value
	}

	return normalizeBackoff(config, overrides)
}

func normalizeBackoff(config RuntimeConfig, overrides agentconfig.RuntimeConfigOverrides) RuntimeConfig {
	if config.MaxBackoff >= config.InitialBackoff {
		return config
	}
	if overrides.MaxBackoff.Defined && !overrides.InitialBackoff.Defined {
		config.InitialBackoff = config.MaxBackoff
		return config
	}

	config.MaxBackoff = config.InitialBackoff
	return config
}
