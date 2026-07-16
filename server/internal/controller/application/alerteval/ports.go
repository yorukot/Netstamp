package alerteval

import (
	"context"
	"encoding/json"
	"time"

	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

type Repository interface {
	ListEnabledRulesForAssignment(ctx context.Context, projectID, probeID, checkID string, checkType domaincheck.Type) ([]domainalert.Rule, error)
	GetMetricSummary(ctx context.Context, metric string, probeStorageID, checkStorageID int64, from, to time.Time) (alertcondition.MetricSummary, error)
	StartOrGetPendingEvaluation(ctx context.Context, projectID, ruleID, probeID, checkID string, firingSince time.Time) (time.Time, error)
	ClearPendingEvaluation(ctx context.Context, projectID, ruleID, probeID, checkID string) error
	GetActiveIncident(ctx context.Context, ruleID, probeID, checkID string) (domainalert.Incident, error)
	GetRecentResolvedIncident(ctx context.Context, ruleID, probeID, checkID string, resolvedAfter time.Time) (domainalert.Incident, error)
	CreateIncident(ctx context.Context, input domainalert.IncidentTransitionInput) (domainalert.Incident, error)
	EnqueueNotificationJobs(ctx context.Context, jobs []domainalert.NotificationJobInput) error
	UpdateIncidentTriggered(ctx context.Context, incidentID string, evaluation alertcondition.Evaluation, summary json.RawMessage, at time.Time) (domainalert.Incident, error)
	UpdateIncidentInsufficient(ctx context.Context, incidentID string, state alertcondition.EvaluationState, summary json.RawMessage, at time.Time) (domainalert.Incident, error)
	ResolveIncident(ctx context.Context, incidentID string, summary json.RawMessage, at time.Time) (domainalert.Incident, error)
	ListEnabledNotificationsForRule(ctx context.Context, projectID, ruleID string) ([]domainalert.Notification, error)
}

type Transactor = apptx.Transactor

type EventRecorder interface {
	RecordAlertEvalEvent(ctx context.Context, event AlertEvalEvent)
}

type AlertEvalEventName string

const (
	AlertEvalEventAssignmentEvaluateFailure AlertEvalEventName = "alert_eval.assignment.evaluate.failure"
	AlertEvalEventRuleEvaluateFailure       AlertEvalEventName = "alert_eval.rule.evaluate.failure"
	AlertEvalEventNotificationEnqueueFail   AlertEvalEventName = "alert_eval.notification.enqueue.failure"
)

type AlertEvalAction string

const (
	AlertEvalActionEvaluateAssignment AlertEvalAction = "assignment.evaluate"
	AlertEvalActionEvaluateRule       AlertEvalAction = "rule.evaluate"
)

type AlertEvalOutcome string

const (
	AlertEvalOutcomeSuccess AlertEvalOutcome = "success"
	AlertEvalOutcomeFailure AlertEvalOutcome = "failure"
)

type AlertEvalReason string

const (
	AlertEvalReasonInvalidCheckType         AlertEvalReason = "invalid_check_type"
	AlertEvalReasonRuleListFailed           AlertEvalReason = "rule_list_failed"
	AlertEvalReasonRuleEvaluateFailed       AlertEvalReason = "rule_evaluate_failed"
	AlertEvalReasonMetricSummaryFailed      AlertEvalReason = "metric_summary_failed"
	AlertEvalReasonEvaluationSummaryFailed  AlertEvalReason = "evaluation_summary_failed"
	AlertEvalReasonPendingTransitionFailed  AlertEvalReason = "pending_transition_failed"
	AlertEvalReasonIncidentLookupFailed     AlertEvalReason = "incident_lookup_failed"
	AlertEvalReasonIncidentTransitionFailed AlertEvalReason = "incident_transition_failed"
	AlertEvalReasonNotificationListFailed   AlertEvalReason = "notification_list_failed"
	AlertEvalReasonNotificationPayloadFail  AlertEvalReason = "notification_payload_failed"
	AlertEvalReasonNotificationEnqueueFail  AlertEvalReason = "notification_enqueue_failed"
)

type AlertEvalEvent struct {
	Name       AlertEvalEventName
	Action     AlertEvalAction
	Outcome    AlertEvalOutcome
	Reason     AlertEvalReason
	ProjectID  string
	ProbeID    string
	CheckID    string
	CheckType  string
	RuleID     string
	IncidentID string
	Err        error
}
