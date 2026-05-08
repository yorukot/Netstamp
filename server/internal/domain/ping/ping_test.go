package ping

import (
	"encoding/json"
	"errors"
	"testing"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.PacketCount != DefaultPacketCount ||
		config.PacketSizeBytes != DefaultPacketSizeBytes ||
		config.TimeoutMs != DefaultTimeoutMs ||
		config.IPFamily != nil {
		t.Fatalf("unexpected default config: %#v", config)
	}
}

func TestNewConfigAppliesOverrides(t *testing.T) {
	var packetCount int32 = 5
	var packetSizeBytes int32 = 128
	var timeoutMs int32 = 1000
	ipFamilyValue := " inet6 "

	config, err := NewConfig(&packetCount, &packetSizeBytes, &timeoutMs, &ipFamilyValue)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	if config.PacketCount != 5 || config.PacketSizeBytes != 128 || config.TimeoutMs != 1000 {
		t.Fatalf("unexpected config values: %#v", config)
	}
	if config.IPFamily == nil || *config.IPFamily != domainnetwork.IPFamilyInet6 {
		t.Fatalf("expected inet6 preference, got %#v", config.IPFamily)
	}
}

func TestNewConfigUsesDefaultsWhenOverridesAreOmitted(t *testing.T) {
	config, err := NewConfig(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("new default config: %v", err)
	}
	if config.PacketCount != DefaultPacketCount ||
		config.PacketSizeBytes != DefaultPacketSizeBytes ||
		config.TimeoutMs != DefaultTimeoutMs ||
		config.IPFamily != nil {
		t.Fatalf("unexpected default config: %#v", config)
	}
}

func TestConfigValidationRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "packet count", run: func() error { return ValidatePacketCount(0) }},
		{name: "new config packet count", run: func() error {
			var value int32
			_, err := NewConfig(&value, nil, nil, nil)
			return err
		}},
		{name: "packet size negative", run: func() error { return ValidatePacketSizeBytes(-1) }},
		{name: "packet size too large", run: func() error { return ValidatePacketSizeBytes(MaxPacketSizeBytes + 1) }},
		{name: "new config packet size", run: func() error {
			value := MaxPacketSizeBytes + 1
			_, err := NewConfig(nil, &value, nil, nil)
			return err
		}},
		{name: "timeout", run: func() error { return ValidateTimeoutMs(0) }},
		{name: "new config timeout", run: func() error {
			var value int32
			_, err := NewConfig(nil, nil, &value, nil)
			return err
		}},
		{name: "ip family", run: func() error {
			raw := "ipv10"
			_, err := NewConfig(nil, nil, nil, &raw)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.run(), ErrInvalidConfig) {
				t.Fatal("expected invalid config error")
			}
		})
	}
}

func TestConfigVersionPayloadIsCanonical(t *testing.T) {
	inet := domainnetwork.IPFamilyInet
	payload := ConfigVersionPayload(Config{
		PacketCount:     4,
		PacketSizeBytes: 56,
		TimeoutMs:       3000,
		IPFamily:        &inet,
	})

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if string(raw) != `{"packetCount":4,"packetSizeBytes":56,"timeoutMs":3000,"ipFamily":"inet"}` {
		t.Fatalf("unexpected payload json: %s", raw)
	}
}
