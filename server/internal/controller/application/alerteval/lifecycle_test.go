package alerteval

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func TestReconcileStoppedEvaluationsResolvesAndNotifiesWhenEvaluationDisabled(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 8, 11, 1, 2, 3, 0, time.UTC)
	rule := pendingTestRule()
	incident := domainalert.Incident{
		ID:                  pendingTestIncidentID,
		ProjectID:           pendingTestProjectID,
		RuleID:              pendingTestRuleID,
		ProbeID:             pendingTestProbeID,
		CheckID:             pendingTestCheckID,
		CheckType:           domaincheck.TypePing,
		Status:              domainalert.IncidentStatusOpen,
		Severity:            domainalert.SeverityCritical,
		LastEvaluationState: alertcondition.EvaluationStateFiring,
		OpenedAt:            at.Add(-time.Hour),
		LastEvaluatedAt:     at.Add(-time.Minute),
		LastTriggeredAt:     at.Add(-time.Minute),
		LastSummary:         json.RawMessage(`{"state":"firing","metric":"ping.loss_percent","samples":4}`),
		Probe:               &domainalert.IncidentProbeSummary{ID: pendingTestProbeID, Name: "Deleted probe"},
		Check:               &domainalert.IncidentCheckSummary{ID: pendingTestCheckID, Name: "GitHub Raw", Type: domaincheck.TypePing, Target: "raw.githubusercontent.com"},
	}
	repo := newPendingEvaluationRepository()
	repo.inactiveCandidates = []domainalert.InactiveIncidentCandidate{{Incident: incident, Rule: rule}}
	repo.notifications = []domainalert.Notification{{
		ID:        "66666666-6666-4666-8666-666666666666",
		ProjectID: pendingTestProjectID,
		Name:      "Operations",
		Type:      domainalert.NotificationTypeWebhook,
		Enabled:   true,
	}}
	service := NewService(repo, false, "https://app.netstamp.test")
	service.now = func() time.Time { return at }

	if err := service.ReconcileStoppedEvaluations(context.Background(), pendingTestProjectID); err != nil {
		t.Fatalf("reconcile stopped evaluations returned error: %v", err)
	}
	if len(repo.resolvedInactiveIncidents) != 1 {
		t.Fatalf("resolved inactive incidents = %d, want 1", len(repo.resolvedInactiveIncidents))
	}
	resolved := repo.resolvedInactiveIncidents[0]
	if resolved.ResolutionReason == nil || *resolved.ResolutionReason != domainalert.IncidentResolutionReasonTargetNoLongerEvaluated {
		t.Fatalf("resolution reason = %v, want target_no_longer_evaluated", resolved.ResolutionReason)
	}
	if resolved.LastEvaluationState != alertcondition.EvaluationStateNoData {
		t.Fatalf("last evaluation state = %q, want no_data", resolved.LastEvaluationState)
	}
	if resolved.ResolvedAt == nil || !resolved.ResolvedAt.Equal(at) {
		t.Fatalf("resolved at = %v, want %v", resolved.ResolvedAt, at)
	}
	if repo.deleteInactivePendingCalls != 1 {
		t.Fatalf("pending cleanup calls = %d, want 1", repo.deleteInactivePendingCalls)
	}
	if len(repo.jobs) != 1 {
		t.Fatalf("notification jobs = %d, want 1", len(repo.jobs))
	}
	if repo.jobs[0].EventType != EventIncidentResolved {
		t.Fatalf("notification event = %q, want %q", repo.jobs[0].EventType, EventIncidentResolved)
	}

	var payload struct {
		Incident struct {
			ResolutionReason domainalert.IncidentResolutionReason `json:"resolutionReason"`
		} `json:"incident"`
		Summary struct {
			State alertcondition.EvaluationState `json:"state"`
			Value *float64                       `json:"value"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(repo.jobs[0].Payload, &payload); err != nil {
		t.Fatalf("notification payload is invalid: %v", err)
	}
	if payload.Incident.ResolutionReason != domainalert.IncidentResolutionReasonTargetNoLongerEvaluated {
		t.Fatalf("payload resolution reason = %q, want target_no_longer_evaluated", payload.Incident.ResolutionReason)
	}
	if payload.Summary.State != alertcondition.EvaluationStateNoData {
		t.Fatalf("payload state = %q, want no_data", payload.Summary.State)
	}
	if payload.Summary.Value != nil {
		t.Fatalf("payload value = %v, want omitted", payload.Summary.Value)
	}
}
