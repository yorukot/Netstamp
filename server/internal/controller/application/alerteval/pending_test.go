package alerteval

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

const (
	pendingTestProjectID  = "11111111-1111-4111-8111-111111111111"
	pendingTestRuleID     = "22222222-2222-4222-8222-222222222222"
	pendingTestProbeID    = "33333333-3333-4333-8333-333333333333"
	pendingTestCheckID    = "44444444-4444-4444-8444-444444444444"
	pendingTestIncidentID = "55555555-5555-4555-8555-555555555555"
)

func TestServiceRequiresSustainedFiringBeforeOpeningIncident(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)
	repo := newPendingEvaluationRepository()
	service := NewService(repo, true, "")
	now := base
	service.now = func() time.Time { return now }

	evaluatePendingTestAssignment(t, service)
	if repo.pendingSince == nil || !repo.pendingSince.Equal(base) {
		t.Fatalf("pending since = %v, want %v", repo.pendingSince, base)
	}
	if repo.createCalls != 0 {
		t.Fatalf("incident create calls = %d, want 0", repo.createCalls)
	}

	now = base.Add(59 * time.Second)
	evaluatePendingTestAssignment(t, service)
	if repo.createCalls != 0 {
		t.Fatalf("incident create calls before duration = %d, want 0", repo.createCalls)
	}

	now = base.Add(time.Minute)
	evaluatePendingTestAssignment(t, service)
	if repo.createCalls != 1 {
		t.Fatalf("incident create calls at duration = %d, want 1", repo.createCalls)
	}
	if repo.pendingSince != nil {
		t.Fatalf("pending evaluation was not cleared: %v", repo.pendingSince)
	}
}

func TestServiceClearEvaluationResetsPendingDuration(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)
	repo := newPendingEvaluationRepository()
	service := NewService(repo, true, "")
	now := base
	service.now = func() time.Time { return now }

	evaluatePendingTestAssignment(t, service)
	repo.summary.Value = 0
	now = base.Add(30 * time.Second)
	evaluatePendingTestAssignment(t, service)
	if repo.pendingSince != nil {
		t.Fatalf("clear evaluation kept pending start: %v", repo.pendingSince)
	}

	repo.summary.Value = 20
	now = base.Add(2 * time.Minute)
	evaluatePendingTestAssignment(t, service)
	if repo.pendingSince == nil || !repo.pendingSince.Equal(now) {
		t.Fatalf("new pending since = %v, want %v", repo.pendingSince, now)
	}
	if repo.createCalls != 0 {
		t.Fatalf("incident create calls after reset = %d, want 0", repo.createCalls)
	}
}

func TestServiceMissingDataPreservesPendingButCannotOpenIncident(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		summary alertcondition.MetricSummary
	}{
		{name: "no data", summary: alertcondition.MetricSummary{}},
		{name: "insufficient samples", summary: alertcondition.MetricSummary{Samples: 1, Value: 20, HasValue: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			base := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)
			repo := newPendingEvaluationRepository()
			service := NewService(repo, true, "")
			now := base
			service.now = func() time.Time { return now }

			evaluatePendingTestAssignment(t, service)
			repo.summary = tt.summary
			now = base.Add(2 * time.Minute)
			evaluatePendingTestAssignment(t, service)
			if repo.pendingSince == nil || !repo.pendingSince.Equal(base) {
				t.Fatalf("missing data changed pending since = %v, want %v", repo.pendingSince, base)
			}
			if repo.createCalls != 0 {
				t.Fatalf("missing data opened %d incidents, want 0", repo.createCalls)
			}

			repo.summary = firingMetricSummary()
			now = base.Add(2*time.Minute + time.Second)
			evaluatePendingTestAssignment(t, service)
			if repo.createCalls != 1 {
				t.Fatalf("confirmed firing created %d incidents, want 1", repo.createCalls)
			}
		})
	}
}

func evaluatePendingTestAssignment(t *testing.T, service *Service) {
	t.Helper()
	if err := service.EvaluateChangedAssignments(context.Background(), []appproberuntime.ChangedAssignmentInput{pendingTestAssignment()}); err != nil {
		t.Fatalf("evaluate changed assignments returned error: %v", err)
	}
}

func pendingTestAssignment() appproberuntime.ChangedAssignmentInput {
	return appproberuntime.ChangedAssignmentInput{
		ProjectID:      pendingTestProjectID,
		ProbeID:        pendingTestProbeID,
		ProbeName:      "Test probe",
		CheckID:        pendingTestCheckID,
		CheckName:      "Test check",
		CheckType:      string(domaincheck.TypePing),
		CheckTarget:    "example.com",
		ProbeStorageID: 1,
		CheckStorageID: 2,
	}
}

func pendingTestRule() domainalert.Rule {
	return domainalert.Rule{
		ID:                  pendingTestRuleID,
		ProjectID:           pendingTestProjectID,
		Name:                "Sustained packet loss",
		Status:              domainalert.RuleStatusEnabled,
		Severity:            domainalert.SeverityCritical,
		CheckType:           domaincheck.TypePing,
		TriggerAfterSeconds: domainalert.DefaultTriggerAfterSeconds,
		CooldownSeconds:     domainalert.DefaultCooldownSeconds,
		Condition: alertcondition.Condition{
			Type:          alertcondition.TypeMetricThreshold,
			Metric:        alertcondition.MetricPingLossPercent,
			Operator:      alertcondition.OperatorGT,
			Threshold:     10,
			WindowSeconds: 60,
			MinSamples:    2,
		},
	}
}

func firingMetricSummary() alertcondition.MetricSummary {
	return alertcondition.MetricSummary{Samples: 2, Value: 20, HasValue: true}
}

type pendingEvaluationRepository struct {
	rule         domainalert.Rule
	summary      alertcondition.MetricSummary
	pendingSince *time.Time
	active       *domainalert.Incident
	createCalls  int
}

func newPendingEvaluationRepository() *pendingEvaluationRepository {
	return &pendingEvaluationRepository{rule: pendingTestRule(), summary: firingMetricSummary()}
}

func (r *pendingEvaluationRepository) ListEnabledRulesForAssignment(context.Context, string, string, string, domaincheck.Type) ([]domainalert.Rule, error) {
	return []domainalert.Rule{r.rule}, nil
}

func (r *pendingEvaluationRepository) GetMetricSummary(context.Context, string, int64, int64, time.Time, time.Time) (alertcondition.MetricSummary, error) {
	return r.summary, nil
}

func (r *pendingEvaluationRepository) StartOrGetPendingEvaluation(_ context.Context, _, _, _, _ string, firingSince time.Time) (time.Time, error) {
	if r.pendingSince == nil {
		started := firingSince.UTC()
		r.pendingSince = &started
	}
	return *r.pendingSince, nil
}

func (r *pendingEvaluationRepository) ClearPendingEvaluation(context.Context, string, string, string, string) error {
	r.pendingSince = nil
	return nil
}

func (r *pendingEvaluationRepository) GetActiveIncident(context.Context, string, string, string) (domainalert.Incident, error) {
	if r.active == nil {
		return domainalert.Incident{}, domainalert.ErrIncidentNotFound
	}
	return *r.active, nil
}

func (r *pendingEvaluationRepository) GetRecentResolvedIncident(context.Context, string, string, string, time.Time) (domainalert.Incident, error) {
	return domainalert.Incident{}, domainalert.ErrIncidentNotFound
}

func (r *pendingEvaluationRepository) CreateIncident(_ context.Context, input domainalert.IncidentTransitionInput) (domainalert.Incident, error) {
	r.createCalls++
	incident := domainalert.Incident{
		ID:                  pendingTestIncidentID,
		ProjectID:           input.Rule.ProjectID,
		RuleID:              input.Rule.ID,
		ProbeID:             input.ProbeID,
		CheckID:             input.CheckID,
		CheckType:           input.CheckType,
		Status:              domainalert.IncidentStatusOpen,
		Severity:            input.Rule.Severity,
		LastEvaluationState: input.Evaluation.State,
		OpenedAt:            input.At,
		LastEvaluatedAt:     input.At,
		LastTriggeredAt:     input.At,
	}
	r.active = &incident
	return incident, nil
}

func (r *pendingEvaluationRepository) EnqueueNotificationJobs(context.Context, []domainalert.NotificationJobInput) error {
	return nil
}

func (r *pendingEvaluationRepository) UpdateIncidentTriggered(_ context.Context, _ string, _ alertcondition.Evaluation, _ json.RawMessage, _ time.Time) (domainalert.Incident, error) {
	return *r.active, nil
}

func (r *pendingEvaluationRepository) UpdateIncidentInsufficient(_ context.Context, _ string, _ alertcondition.EvaluationState, _ json.RawMessage, _ time.Time) (domainalert.Incident, error) {
	return *r.active, nil
}

func (r *pendingEvaluationRepository) ResolveIncident(_ context.Context, _ string, _ json.RawMessage, at time.Time) (domainalert.Incident, error) {
	incident := *r.active
	incident.Status = domainalert.IncidentStatusResolved
	incident.ResolvedAt = &at
	r.active = nil
	return incident, nil
}

func (r *pendingEvaluationRepository) ListEnabledNotificationsForRule(context.Context, string, string) ([]domainalert.Notification, error) {
	return nil, nil
}
