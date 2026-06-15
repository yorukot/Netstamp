package notify

import "testing"

const (
	renderedIncidentNotificationPayload = `{
	"eventType": "incident.opened",
	"sentAt": "2026-06-15T05:23:19Z",
	"rule": {
		"name": "Ping Alert",
		"severity": "critical"
	},
	"target": {
		"probeId": "819adf83-2a58-475a-9a01-872c17296abb",
		"checkId": "bb4e5352-6164-4b9b-b6ce-a4b4afb8537d",
		"checkType": "ping",
		"probe": {
			"name": "Pan Tencent Cloud"
		},
		"check": {
			"name": "GitHub Raw",
			"type": "ping",
			"target": "raw.githubusercontent.com"
		}
	},
	"links": {
		"incident": "https://netstamp.dev/projects/demo/alerts/incident/22222222-2222-4222-8222-222222222222"
	},
	"summary": {
		"metric": "ping.loss_percent",
		"value": 51.01,
		"threshold": 10
	}
}`
	renderedIncidentURL = "https://netstamp.dev/projects/demo/alerts/incident/22222222-2222-4222-8222-222222222222"
)

func TestParseIncidentNotificationPayload(t *testing.T) {
	incident, ok := parseIncidentNotificationPayload([]byte(renderedIncidentNotificationPayload))
	if !ok {
		t.Fatal("expected incident payload to parse")
	}
	if incident.EventType != "incident.opened" {
		t.Fatalf("expected event type, got %q", incident.EventType)
	}
	if incidentProbeLabel(incident) != "Pan Tencent Cloud" {
		t.Fatalf("expected probe label, got %q", incidentProbeLabel(incident))
	}
	if incidentCheckLabel(incident) != "GitHub Raw (PING)" {
		t.Fatalf("expected check label, got %q", incidentCheckLabel(incident))
	}
	if incident.Links.Incident != renderedIncidentURL {
		t.Fatalf("expected incident link %q, got %q", renderedIncidentURL, incident.Links.Incident)
	}
}

func TestParseIncidentNotificationPayloadRejectsInvalidJSON(t *testing.T) {
	if _, ok := parseIncidentNotificationPayload([]byte("{")); ok {
		t.Fatal("expected invalid JSON to be rejected")
	}
}
