package check

import (
	"encoding/json"
	"testing"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestCheckVersionUsesExecutionSpec(t *testing.T) {
	base := ExecutionSpec{
		Type:            TypePing,
		Target:          "api.netstamp.io",
		IntervalSeconds: 30,
		PingConfig: domainping.Config{
			PacketCount:     4,
			PacketSizeBytes: 56,
			TimeoutMs:       3000,
		},
	}

	baseVersion := CheckVersion(base)
	if baseVersion == "" {
		t.Fatal("expected check version")
	}

	changedTarget := base
	changedTarget.Target = "status.netstamp.io"
	if got := CheckVersion(changedTarget); got == baseVersion {
		t.Fatal("expected target change to change check version")
	}

	changedInterval := base
	changedInterval.IntervalSeconds = 60
	if got := CheckVersion(changedInterval); got == baseVersion {
		t.Fatal("expected interval change to change check version")
	}

	changedConfig := base
	changedConfig.PingConfig.TimeoutMs = 1500
	if got := CheckVersion(changedConfig); got == baseVersion {
		t.Fatal("expected ping config change to change check version")
	}
}

func TestCheckVersionIncludesIPFamily(t *testing.T) {
	base := ExecutionSpec{
		Type:            TypePing,
		Target:          "api.netstamp.io",
		IntervalSeconds: 30,
		PingConfig: domainping.Config{
			PacketCount:     4,
			PacketSizeBytes: 56,
			TimeoutMs:       3000,
		},
	}

	inet := domainnetwork.IPFamilyInet
	withIPFamily := base
	withIPFamily.PingConfig.IPFamily = &inet

	if CheckVersion(base) == CheckVersion(withIPFamily) {
		t.Fatal("expected ip family change to change check version")
	}
}

func TestSelectorVersionHashesCanonicalSelector(t *testing.T) {
	matchAll := json.RawMessage(`{}`)
	regionTokyo := json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`)

	if SelectorVersion(matchAll) == "" {
		t.Fatal("expected selector version")
	}
	if SelectorVersion(matchAll) == SelectorVersion(regionTokyo) {
		t.Fatal("expected different selectors to have different versions")
	}
}

func TestHashJSONPanicsWhenValueCannotBeMarshaled(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = hashJSON(func() {})
}
