package notify

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderDiscordWebhookBodyUsesReadableNamesAndIncidentLink(t *testing.T) {
	data, err := renderDiscordWebhookBody([]byte(renderedIncidentNotificationPayload))
	if err != nil {
		t.Fatalf("renderDiscordWebhookBody returned error: %v", err)
	}

	var body struct {
		Embeds []struct {
			URL    string `json:"url"`
			Fields []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"fields"`
		} `json:"embeds"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("discord body is not valid JSON: %v", err)
	}
	if len(body.Embeds) != 1 {
		t.Fatalf("expected one embed, got %d", len(body.Embeds))
	}

	embed := body.Embeds[0]
	if embed.URL != renderedIncidentURL {
		t.Fatalf("expected embed URL %q, got %q", renderedIncidentURL, embed.URL)
	}
	if fieldValue(embed.Fields, "Probe") != "Pan Tencent Cloud" {
		t.Fatalf("expected probe name field, got %q", fieldValue(embed.Fields, "Probe"))
	}
	if fieldValue(embed.Fields, "Check") != "GitHub Raw (PING)" {
		t.Fatalf("expected check name field, got %q", fieldValue(embed.Fields, "Check"))
	}
	if fieldValue(embed.Fields, "Incident") != renderedIncidentURL {
		t.Fatalf("expected incident link field, got %q", fieldValue(embed.Fields, "Incident"))
	}
	if strings.Contains(fieldValue(embed.Fields, "Probe"), "819adf83") || strings.Contains(fieldValue(embed.Fields, "Check"), "bb4e5352") {
		t.Fatalf("expected readable names instead of UUID labels, got probe=%q check=%q", fieldValue(embed.Fields, "Probe"), fieldValue(embed.Fields, "Check"))
	}
}

func TestRenderDiscordWebhookBodyFallsBackForNonIncidentPayload(t *testing.T) {
	data, err := renderDiscordWebhookBody([]byte("raw notification body"))
	if err != nil {
		t.Fatalf("renderDiscordWebhookBody returned error: %v", err)
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("discord fallback body is not valid JSON: %v", err)
	}
	if !strings.Contains(body.Content, "raw notification body") {
		t.Fatalf("expected fallback content to include raw payload, got %q", body.Content)
	}
}

func fieldValue(fields []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}, name string,
) string {
	for _, field := range fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}
