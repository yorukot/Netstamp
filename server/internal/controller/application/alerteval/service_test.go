package alerteval

import (
	"encoding/json"
	"testing"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func TestNotificationPayloadIncludesReadableTargetAndIncidentLink(t *testing.T) {
	at := time.Date(2026, 6, 15, 5, 23, 19, 0, time.UTC)
	rule := domainalert.Rule{
		ID:        "3a3d13bf-0000-4000-8000-000000000001",
		ProjectID: "11111111-1111-4111-8111-111111111111",
		Name:      "Ping Alert",
		Severity:  domainalert.SeverityCritical,
		CheckType: domaincheck.TypePing,
		Condition: alertcondition.Condition{
			Metric:        alertcondition.MetricPingLossPercent,
			Operator:      alertcondition.OperatorGTE,
			Threshold:     10,
			WindowSeconds: 300,
			MinSamples:    3,
		},
	}
	incident := domainalert.Incident{
		ID:        "22222222-2222-4222-8222-222222222222",
		ProjectID: rule.ProjectID,
		RuleID:    rule.ID,
		ProbeID:   "819adf83-2a58-475a-9a01-872c17296abb",
		CheckID:   "bb4e5352-6164-4b9b-b6ce-a4b4afb8537d",
		Probe:     &domainalert.IncidentProbeSummary{ID: "819adf83-2a58-475a-9a01-872c17296abb", Name: "Pan Tencent Cloud"},
		Check:     &domainalert.IncidentCheckSummary{ID: "bb4e5352-6164-4b9b-b6ce-a4b4afb8537d", Name: "GitHub Raw", Type: domaincheck.TypePing, Target: "raw.githubusercontent.com"},
		CheckType: domaincheck.TypePing,
		Status:    domainalert.IncidentStatusOpen,
		OpenedAt:  at,
	}
	notification := domainalert.Notification{
		ID:   "33333333-3333-4333-8333-333333333333",
		Name: "Discord ops",
		Type: domainalert.NotificationTypeDiscord,
	}
	evaluation := alertcondition.Evaluation{
		State: alertcondition.EvaluationStateFiring,
		Value: 51.01,
		Summary: alertcondition.MetricSummary{
			Samples: 5,
		},
	}

	payload, err := notificationPayload(rule, incident, notification, evaluation, EventIncidentOpened, at, "https://app.netstamp.dev/")
	if err != nil {
		t.Fatalf("notificationPayload returned error: %v", err)
	}

	var body struct {
		Target struct {
			ProbeID string `json:"probeId"`
			CheckID string `json:"checkId"`
			Probe   struct {
				Name string `json:"name"`
			} `json:"probe"`
			Check struct {
				Name   string `json:"name"`
				Type   string `json:"type"`
				Target string `json:"target"`
			} `json:"check"`
		} `json:"target"`
		Notification struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"notification"`
		Links struct {
			Incident string `json:"incident"`
		} `json:"links"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}

	if body.Target.Probe.Name != "Pan Tencent Cloud" {
		t.Fatalf("expected probe name in payload, got %q", body.Target.Probe.Name)
	}
	if body.Target.Check.Name != "GitHub Raw" || body.Target.Check.Target != "raw.githubusercontent.com" {
		t.Fatalf("expected check summary in payload, got %#v", body.Target.Check)
	}
	if body.Target.ProbeID != incident.ProbeID || body.Target.CheckID != incident.CheckID {
		t.Fatalf("expected target IDs to remain available, got probe=%q check=%q", body.Target.ProbeID, body.Target.CheckID)
	}
	if body.Notification.Name != "Discord ops" || body.Notification.Type != string(domainalert.NotificationTypeDiscord) {
		t.Fatalf("expected notification summary in payload, got %#v", body.Notification)
	}
	expectedURL := "https://app.netstamp.dev/projects/11111111-1111-4111-8111-111111111111/alerts/incident/22222222-2222-4222-8222-222222222222"
	if body.Links.Incident != expectedURL {
		t.Fatalf("expected incident link %q, got %q", expectedURL, body.Links.Incident)
	}
}

func TestAlertIncidentURLOmitsWhenBaseURLUnset(t *testing.T) {
	if got := alertIncidentURL("", "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"); got != "" {
		t.Fatalf("expected empty link without base URL, got %q", got)
	}
}
