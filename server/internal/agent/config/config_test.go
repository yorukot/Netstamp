package config

import "testing"

func TestLoadConfigDefaultsLocalRuntimeSettings(t *testing.T) {
	t.Setenv("NETSTAMP_PROBE_CONTROLLER_URL", "http://localhost:8080")
	t.Setenv("NETSTAMP_PROBE_ID", "77777777-7777-7777-7777-777777777777")
	t.Setenv("NETSTAMP_PROBE_SECRET", "probe-secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.MaxWorkers.Value != 128 {
		t.Fatalf("expected default max workers 128, got %d", cfg.MaxWorkers.Value)
	}
	if cfg.PprofAddr.Value != "" || cfg.PprofAddr.Defined {
		t.Fatalf("expected pprof addr to be disabled by default, got %#v", cfg.PprofAddr)
	}
	if cfg.MetricsAddr.Value != "" || cfg.MetricsAddr.Defined {
		t.Fatalf("expected metrics addr to be disabled by default, got %#v", cfg.MetricsAddr)
	}
}

func TestLoadConfigEnablesObservabilityAddresses(t *testing.T) {
	t.Setenv("NETSTAMP_PROBE_CONTROLLER_URL", "http://localhost:8080")
	t.Setenv("NETSTAMP_PROBE_ID", "77777777-7777-7777-7777-777777777777")
	t.Setenv("NETSTAMP_PROBE_SECRET", "probe-secret")
	t.Setenv("NETSTAMP_PROBE_PPROF_ADDR", "127.0.0.1:6060")
	t.Setenv("NETSTAMP_PROBE_METRICS_ADDR", "127.0.0.1:9091")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.PprofAddr.Value != "127.0.0.1:6060" || !cfg.PprofAddr.Defined {
		t.Fatalf("expected pprof addr to be enabled, got %#v", cfg.PprofAddr)
	}
	if cfg.MetricsAddr.Value != "127.0.0.1:9091" || !cfg.MetricsAddr.Defined {
		t.Fatalf("expected metrics addr to be enabled, got %#v", cfg.MetricsAddr)
	}
}
