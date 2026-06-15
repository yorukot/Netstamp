package alert

import (
	"encoding/json"
	"testing"
)

func TestVNSlackConfig(t *testing.T) {
	canonical, config, err := VNSlackConfig(json.RawMessage(`{"url":" https://hooks.slack.com/services/T000/B000/token "}`))
	if err != nil {
		t.Fatalf("expected valid Slack config: %v", err)
	}
	if config.URL != "https://hooks.slack.com/services/T000/B000/token" {
		t.Fatalf("expected trimmed Slack URL, got %q", config.URL)
	}
	if string(canonical) != `{"url":"https://hooks.slack.com/services/T000/B000/token"}` {
		t.Fatalf("expected canonical Slack config, got %s", canonical)
	}
}

func TestVNSlackConfigRejectsNonSlackURL(t *testing.T) {
	_, _, err := VNSlackConfig(json.RawMessage(`{"url":"https://hooks.example.com/services/T000/B000/token"}`))
	if err == nil {
		t.Fatal("expected non-Slack URL to be rejected")
	}
}

func TestVNEmailConfigCanonicalizesRecipients(t *testing.T) {
	canonical, config, err := VNEmailConfig(json.RawMessage(`{"to":[" Ops@Example.COM ","sre@example.com","ops@example.com"]}`))
	if err != nil {
		t.Fatalf("expected valid email config: %v", err)
	}
	if len(config.To) != 2 || config.To[0] != "ops@example.com" || config.To[1] != "sre@example.com" {
		t.Fatalf("expected normalized unique recipients, got %#v", config.To)
	}
	if string(canonical) != `{"to":["ops@example.com","sre@example.com"]}` {
		t.Fatalf("expected canonical email config, got %s", canonical)
	}
}

func TestVNEmailConfigRejectsEmptyRecipients(t *testing.T) {
	_, _, err := VNEmailConfig(json.RawMessage(`{"to":[]}`))
	if err == nil {
		t.Fatal("expected empty recipient list to be rejected")
	}
}
