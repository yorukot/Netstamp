package alerteval

import (
	"context"
	"encoding/json"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

type Repository interface {
	ListEnabledRulesForAssignment(ctx context.Context, projectID, probeID, checkID string, checkType domaincheck.Type) ([]domainalert.Rule, error)
	GetMetricSummary(ctx context.Context, metric string, probeStorageID, checkStorageID int64, from, to time.Time) (alertcondition.MetricSummary, error)
	GetActiveIncident(ctx context.Context, ruleID, probeID, checkID string) (domainalert.Incident, error)
	GetRecentResolvedIncident(ctx context.Context, ruleID, probeID, checkID string, resolvedAfter time.Time) (domainalert.Incident, error)
	CreateIncidentAndEnqueue(ctx context.Context, input domainalert.IncidentTransitionInput) (domainalert.Incident, error)
	EnqueueNotificationJobs(ctx context.Context, jobs []domainalert.NotificationJobInput) error
	UpdateIncidentTriggered(ctx context.Context, incidentID string, evaluation alertcondition.Evaluation, summary json.RawMessage, at time.Time) (domainalert.Incident, error)
	UpdateIncidentInsufficient(ctx context.Context, incidentID string, state alertcondition.EvaluationState, summary json.RawMessage, at time.Time) (domainalert.Incident, error)
	ResolveIncidentAndEnqueue(ctx context.Context, incidentID string, summary json.RawMessage, at time.Time, jobs []domainalert.NotificationJobInput) (domainalert.Incident, error)
	ListEnabledNotificationsForRule(ctx context.Context, projectID, ruleID string) ([]domainalert.Notification, error)
}

type ChangedAssignment = appproberuntime.ChangedAssignmentInput
