package notify

import (
	"strings"
	"testing"
)

func TestRenderTelegramMessageIncludesIncidentLink(t *testing.T) {
	text := renderTelegramMessage([]byte(renderedIncidentNotificationPayload))
	for _, expected := range []string{
		"Incident: " + renderedIncidentURL,
		"Probe: Pan Tencent Cloud",
		"Check: GitHub Raw (PING)",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected text to contain %q, got:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "Probe: 819adf83") || strings.Contains(text, "Check: bb4e5352") {
		t.Fatalf("expected readable names instead of UUID labels, got:\n%s", text)
	}
}

func TestRenderTelegramMessageFallsBackForNonIncidentPayload(t *testing.T) {
	text := renderTelegramMessage([]byte("raw notification body"))
	if !strings.Contains(text, "raw notification body") {
		t.Fatalf("expected fallback text to include raw payload, got %q", text)
	}
}
