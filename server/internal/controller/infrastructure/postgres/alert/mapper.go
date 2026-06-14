package pgalert

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func mapRule(row sqlc.AlertRule, channelIDs []string) domainalert.Rule {
	condition, parseErr := alertcondition.Parse(row.Condition)
	if parseErr != nil {
		condition = alertcondition.Condition{}
	}
	return domainalert.Rule{
		ID:                     row.ID.String(),
		ProjectID:              row.ProjectID.String(),
		Name:                   row.Name,
		Description:            row.Description,
		Status:                 domainalert.RuleStatus(row.Status),
		Severity:               domainalert.Severity(row.Severity),
		CheckType:              domaincheck.Type(row.CheckType),
		ProbeID:                stringUUIDPtr(row.ProbeID),
		CheckID:                stringUUIDPtr(row.CheckID),
		ProbeSelector:          json.RawMessage(row.ProbeSelector),
		Condition:              condition,
		ConditionJSON:          json.RawMessage(row.Condition),
		ConditionVersion:       row.ConditionVersion,
		CooldownSeconds:        row.CooldownSeconds,
		NotificationChannelIDs: channelIDs,
		CreatedByUserID:        row.CreatedByUserID.String(),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
}

func mapIncident(row sqlc.AlertIncident) domainalert.Incident {
	return domainalert.Incident{
		ID:                          row.ID.String(),
		ProjectID:                   row.ProjectID.String(),
		RuleID:                      row.RuleID.String(),
		ProbeID:                     row.ProbeID.String(),
		CheckID:                     row.CheckID.String(),
		CheckType:                   domaincheck.Type(row.CheckType),
		Status:                      domainalert.IncidentStatus(row.Status),
		Severity:                    domainalert.Severity(row.Severity),
		LastEvaluationState:         alertcondition.EvaluationState(row.LastEvaluationState),
		OpenedAt:                    row.OpenedAt,
		AcknowledgedAt:              row.AcknowledgedAt,
		AcknowledgedByUserID:        stringUUIDPtr(row.AcknowledgedByUserID),
		ResolvedAt:                  row.ResolvedAt,
		ResolvedByUserID:            stringUUIDPtr(row.ResolvedByUserID),
		LastEvaluatedAt:             row.LastEvaluatedAt,
		LastTriggeredAt:             row.LastTriggeredAt,
		LastValue:                   row.LastValue,
		LastSummary:                 json.RawMessage(row.LastSummary),
		LastNotificationSentAt:      row.LastNotificationSentAt,
		NextNotificationEligibleAt:  row.NextNotificationEligibleAt,
		SuppressedNotificationCount: row.SuppressedNotificationCount,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	}
}

func mapGetIncident(row sqlc.GetAlertIncidentRow) domainalert.Incident {
	incident := mapIncident(sqlc.AlertIncident{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		RuleID:                      row.RuleID,
		ProbeID:                     row.ProbeID,
		CheckID:                     row.CheckID,
		CheckType:                   row.CheckType,
		Status:                      row.Status,
		Severity:                    row.Severity,
		LastEvaluationState:         row.LastEvaluationState,
		OpenedAt:                    row.OpenedAt,
		AcknowledgedAt:              row.AcknowledgedAt,
		AcknowledgedByUserID:        row.AcknowledgedByUserID,
		ResolvedAt:                  row.ResolvedAt,
		ResolvedByUserID:            row.ResolvedByUserID,
		LastEvaluatedAt:             row.LastEvaluatedAt,
		LastTriggeredAt:             row.LastTriggeredAt,
		LastValue:                   row.LastValue,
		LastSummary:                 row.LastSummary,
		LastNotificationSentAt:      row.LastNotificationSentAt,
		NextNotificationEligibleAt:  row.NextNotificationEligibleAt,
		SuppressedNotificationCount: row.SuppressedNotificationCount,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
	incident.Probe = &domainalert.IncidentProbeSummary{ID: row.ProbeID.String(), Name: row.ProbeName}
	incident.Check = &domainalert.IncidentCheckSummary{ID: row.CheckID.String(), Name: row.CheckName, Type: domaincheck.Type(row.CheckSummaryType), Target: row.CheckTarget}
	return incident
}

func mapListIncident(row sqlc.ListAlertIncidentsRow) domainalert.Incident {
	incident := mapIncident(sqlc.AlertIncident{
		ID:                          row.ID,
		ProjectID:                   row.ProjectID,
		RuleID:                      row.RuleID,
		ProbeID:                     row.ProbeID,
		CheckID:                     row.CheckID,
		CheckType:                   row.CheckType,
		Status:                      row.Status,
		Severity:                    row.Severity,
		LastEvaluationState:         row.LastEvaluationState,
		OpenedAt:                    row.OpenedAt,
		AcknowledgedAt:              row.AcknowledgedAt,
		AcknowledgedByUserID:        row.AcknowledgedByUserID,
		ResolvedAt:                  row.ResolvedAt,
		ResolvedByUserID:            row.ResolvedByUserID,
		LastEvaluatedAt:             row.LastEvaluatedAt,
		LastTriggeredAt:             row.LastTriggeredAt,
		LastValue:                   row.LastValue,
		LastSummary:                 row.LastSummary,
		LastNotificationSentAt:      row.LastNotificationSentAt,
		NextNotificationEligibleAt:  row.NextNotificationEligibleAt,
		SuppressedNotificationCount: row.SuppressedNotificationCount,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	})
	incident.Probe = &domainalert.IncidentProbeSummary{ID: row.ProbeID.String(), Name: row.ProbeName}
	incident.Check = &domainalert.IncidentCheckSummary{ID: row.CheckID.String(), Name: row.CheckName, Type: domaincheck.Type(row.CheckSummaryType), Target: row.CheckTarget}
	return incident
}

func mapChannel(row sqlc.NotificationChannel) domainalert.NotificationChannel {
	return domainalert.NotificationChannel{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domainalert.ChannelType(row.Type),
		Enabled:         row.Enabled,
		Config:          json.RawMessage(row.Config),
		CreatedByUserID: row.CreatedByUserID.String(),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapOutbox(row sqlc.NotificationOutbox) domainalert.NotificationOutboxJob {
	return domainalert.NotificationOutboxJob{
		ID:            row.ID.String(),
		ProjectID:     row.ProjectID.String(),
		IncidentID:    row.IncidentID.String(),
		RuleID:        row.RuleID.String(),
		ChannelID:     row.ChannelID.String(),
		ChannelType:   domainalert.ChannelType(row.ChannelType),
		EventType:     row.EventType,
		Status:        domainalert.OutboxStatus(row.Status),
		Payload:       json.RawMessage(row.Payload),
		AttemptCount:  row.AttemptCount,
		MaxAttempts:   row.MaxAttempts,
		NextAttemptAt: row.NextAttemptAt,
		LastAttemptAt: row.LastAttemptAt,
		DeliveredAt:   row.DeliveredAt,
		LastErrorKind: row.LastErrorKind,
		LastErrorCode: row.LastErrorCode,
		LastError:     row.LastError,
		DedupeKey:     row.DedupeKey,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func stringUUIDPtr(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}

func sqlcRuleStatus(value domainalert.RuleStatus) sqlc.AlertRuleStatus {
	return sqlc.AlertRuleStatus(value)
}

func sqlcSeverity(value domainalert.Severity) sqlc.AlertSeverity {
	return sqlc.AlertSeverity(value)
}

func sqlcCheckType(value domaincheck.Type) sqlc.CheckType {
	return sqlc.CheckType(value)
}

func sqlcEvaluationState(value alertcondition.EvaluationState) sqlc.AlertEvaluationState {
	return sqlc.AlertEvaluationState(value)
}

func sqlcChannelType(value domainalert.ChannelType) sqlc.NotificationChannelType {
	return sqlc.NotificationChannelType(value)
}
