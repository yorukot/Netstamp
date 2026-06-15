package notify

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func TestRenderEmailMessageIncludesIncidentFields(t *testing.T) {
	message, err := renderEmailMessage(
		[]byte(renderedIncidentNotificationPayload),
		"alerts@example.com",
		[]string{"ops@example.com", "sre@example.com"},
		time.Date(2026, 6, 15, 5, 23, 19, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("renderEmailMessage returned error: %v", err)
	}

	text := string(message)
	for _, expected := range []string{
		`From: "Netstamp" <alerts@example.com>`,
		"To: <ops@example.com>, <sre@example.com>",
		"Subject: Netstamp alert: Ping Alert",
		"Incident: " + renderedIncidentURL,
		"Probe: Pan Tencent Cloud",
		"Check: GitHub Raw (PING)",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected email message to contain %q, got:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "Probe: 819adf83") || strings.Contains(text, "Check: bb4e5352") {
		t.Fatalf("expected readable names instead of UUID labels, got:\n%s", text)
	}
}

func TestSendEmailReportsUnconfiguredSMTP(t *testing.T) {
	config, err := json.Marshal(domainalert.EmailConfig{To: []string{"ops@example.com"}})
	if err != nil {
		t.Fatalf("marshal email config: %v", err)
	}
	sender := NewWebhookSender(time.Second)

	result := sender.SendNotification(context.Background(), domainalert.Notification{
		Type:   domainalert.NotificationTypeEmail,
		Config: config,
	}, []byte(renderedIncidentNotificationPayload))

	if result.Delivered {
		t.Fatal("expected email delivery to fail without SMTP config")
	}
	if result.Kind != "config" || result.Code != "smtp_unconfigured" {
		t.Fatalf("expected smtp_unconfigured config failure, got %#v", result)
	}
}
