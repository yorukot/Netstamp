package notify

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderSlackWebhookBodyUsesReadableNamesAndIncidentLink(t *testing.T) {
	data, err := renderSlackWebhookBody([]byte(renderedIncidentNotificationPayload))
	if err != nil {
		t.Fatalf("renderSlackWebhookBody returned error: %v", err)
	}

	var body struct {
		Text   string `json:"text"`
		Blocks []struct {
			Fields []struct {
				Text string `json:"text"`
			} `json:"fields"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("slack body is not valid JSON: %v", err)
	}
	if !strings.Contains(body.Text, "Netstamp alert") {
		t.Fatalf("expected fallback text to include alert title, got %q", body.Text)
	}

	allFields := slackFieldText(body.Blocks)
	for _, expected := range []string{
		"*Incident*\n" + renderedIncidentURL,
		"*Probe*\nPan Tencent Cloud",
		"*Check*\nGitHub Raw (PING)",
	} {
		if !strings.Contains(allFields, expected) {
			t.Fatalf("expected Slack fields to contain %q, got:\n%s", expected, allFields)
		}
	}
	if strings.Contains(allFields, "Probe*\n819adf83") || strings.Contains(allFields, "Check*\nbb4e5352") {
		t.Fatalf("expected readable names instead of UUID labels, got:\n%s", allFields)
	}
}

func TestRenderSlackWebhookBodyFallsBackForNonIncidentPayload(t *testing.T) {
	data, err := renderSlackWebhookBody([]byte("raw notification body"))
	if err != nil {
		t.Fatalf("renderSlackWebhookBody returned error: %v", err)
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("slack fallback body is not valid JSON: %v", err)
	}
	if !strings.Contains(body.Text, "raw notification body") {
		t.Fatalf("expected fallback text to include raw payload, got %q", body.Text)
	}
}

func slackFieldText(blocks []struct {
	Fields []struct {
		Text string `json:"text"`
	} `json:"fields"`
}) string {
	var values []string
	for _, block := range blocks {
		for _, field := range block.Fields {
			values = append(values, field.Text)
		}
	}
	return strings.Join(values, "\n")
}
