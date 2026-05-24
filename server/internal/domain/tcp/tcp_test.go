package tcp

import (
	"testing"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	if got.Port != DefaultPort || got.TimeoutMs != DefaultTimeoutMs || got.IPFamily != nil {
		t.Fatalf("unexpected default config: %#v", got)
	}
}

func TestConfigValidation(t *testing.T) {
	if _, err := VNConfigPort(0); err == nil {
		t.Fatal("expected invalid low port")
	}
	if _, err := VNConfigPort(MaxPort + 1); err == nil {
		t.Fatal("expected invalid high port")
	}
	if _, err := VNConfigPort(443); err != nil {
		t.Fatalf("expected valid port: %v", err)
	}
	if _, err := VNConfigTimeoutMs(0); err == nil {
		t.Fatal("expected invalid timeout")
	}
	if _, err := VNConfigTimeoutMs(3000); err != nil {
		t.Fatalf("expected valid timeout: %v", err)
	}
}

func TestIPFamilyValidation(t *testing.T) {
	inet := domainnetwork.IPFamilyInet
	if got, err := VNConfigIPFamily(&inet); err != nil || got == nil || *got != inet {
		t.Fatalf("expected valid inet family, got %#v err=%v", got, err)
	}

	invalid := domainnetwork.IPFamily("ip4")
	if _, err := VNConfigIPFamily(&invalid); err == nil {
		t.Fatal("expected invalid ip family")
	}
}

func TestResultValidation(t *testing.T) {
	if _, err := VNResultStatus(StatusSuccessful); err != nil {
		t.Fatalf("expected valid status: %v", err)
	}
	if _, err := VNResultStatus(Status("partial")); err == nil {
		t.Fatal("expected invalid status")
	}

	connectDuration := 42.0
	got, err := VNResultConnectDurationMs(&connectDuration)
	if err != nil || got == nil || *got != connectDuration {
		t.Fatalf("expected valid connect duration, got %#v err=%v", got, err)
	}

	negative := -1.0
	if _, err := VNResultConnectDurationMs(&negative); err == nil {
		t.Fatal("expected invalid connect duration")
	}
}
