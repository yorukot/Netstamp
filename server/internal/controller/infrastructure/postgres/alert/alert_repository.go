package pgalert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *Repository) ListRules(ctx context.Context, projectID string, status *domainalert.RuleStatus, checkType *domaincheck.Type) ([]domainalert.Rule, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "alert_rules", "postgres.alert_rules.list", "SELECT", "SELECT alert rules")
	defer span.End()

	projectUUID, err := postgres.ParseUUID(projectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	var statusParam *sqlc.AlertRuleStatus
	if status != nil {
		value := sqlcRuleStatus(*status)
		statusParam = &value
	}
	var typeParam *sqlc.CheckType
	if checkType != nil {
		value := sqlcCheckType(*checkType)
		typeParam = &value
	}
	rows, err := r.queries.ListAlertRules(ctx, sqlc.ListAlertRulesParams{
		ProjectID: projectUUID,
		Status:    statusParam,
		CheckType: typeParam,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	return r.mapRulesWithNotifications(ctx, projectUUID, rows)
}

func (r *Repository) GetRule(ctx context.Context, projectID, ruleID string) (domainalert.Rule, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "alert_rules", "postgres.alert_rules.get", "SELECT", "SELECT alert rule")
	defer span.End()

	projectUUID, ruleUUID, err := parseProjectScopedID(projectID, ruleID, domainalert.ErrRuleNotFound)
	if err != nil {
		return domainalert.Rule{}, err
	}
	row, err := r.queries.GetAlertRule(ctx, sqlc.GetAlertRuleParams{ProjectID: projectUUID, ID: ruleUUID})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainalert.Rule{}, mapNoRows(err, domainalert.ErrRuleNotFound)
	}
	notificationIDs, err := r.ruleNotificationIDs(ctx, projectUUID, ruleUUID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainalert.Rule{}, err
	}
	return mapRule(row, notificationIDs), nil
}

func (r *Repository) CreateRule(ctx context.Context, input domainalert.Rule) (domainalert.Rule, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "alert_rules", "postgres.alert_rules.create", "INSERT", "INSERT alert rule")
	defer span.End()

	var created domainalert.Rule
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		params, err := createRuleParams(input)
		if err != nil {
			return err
		}
		row, err := q.CreateAlertRule(ctx, params)
		if err != nil {
			return err
		}
		for _, notificationID := range input.NotificationIDs {
			notificationUUID, parseErr := postgres.ParseUUID(notificationID, domainalert.ErrNotificationNotFound)
			if parseErr != nil {
				return parseErr
			}
			if err := q.AddAlertNotification(ctx, sqlc.AddAlertNotificationParams{
				ProjectID:      row.ProjectID,
				RuleID:         row.ID,
				NotificationID: notificationUUID,
			}); err != nil {
				return err
			}
		}
		created = mapRule(row, input.NotificationIDs)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainalert.Rule{}, err
	}
	return created, nil
}

func (r *Repository) UpdateRule(ctx context.Context, input domainalert.Rule) (domainalert.Rule, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "alert_rules", "postgres.alert_rules.update", "UPDATE", "UPDATE alert rule")
	defer span.End()

	var updated domainalert.Rule
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		params, err := updateRuleParams(input)
		if err != nil {
			return err
		}
		row, err := q.UpdateAlertRule(ctx, params)
		if err != nil {
			return mapNoRows(err, domainalert.ErrRuleNotFound)
		}
		if err := q.ReplaceAlertNotifications(ctx, sqlc.ReplaceAlertNotificationsParams{ProjectID: row.ProjectID, RuleID: row.ID}); err != nil {
			return err
		}
		for _, notificationID := range input.NotificationIDs {
			notificationUUID, parseErr := postgres.ParseUUID(notificationID, domainalert.ErrNotificationNotFound)
			if parseErr != nil {
				return parseErr
			}
			if err := q.AddAlertNotification(ctx, sqlc.AddAlertNotificationParams{
				ProjectID:      row.ProjectID,
				RuleID:         row.ID,
				NotificationID: notificationUUID,
			}); err != nil {
				return err
			}
		}
		updated = mapRule(row, input.NotificationIDs)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainalert.Rule{}, err
	}
	return updated, nil
}

func (r *Repository) DeleteRule(ctx context.Context, projectID, ruleID string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "alert_rules", "postgres.alert_rules.delete", "UPDATE", "SOFT DELETE alert rule")
	defer span.End()

	projectUUID, ruleUUID, err := parseProjectScopedID(projectID, ruleID, domainalert.ErrRuleNotFound)
	if err != nil {
		return err
	}
	rows, err := r.queries.SoftDeleteAlertRule(ctx, sqlc.SoftDeleteAlertRuleParams{ProjectID: projectUUID, ID: ruleUUID})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	if rows == 0 {
		return domainalert.ErrRuleNotFound
	}
	return nil
}

func (r *Repository) ListNotifications(ctx context.Context, projectID string, notificationType *domainalert.NotificationType) ([]domainalert.Notification, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgalertTracer, "notifications", "postgres.notifications.list", "SELECT", "SELECT notifications")
	defer span.End()

	projectUUID, err := postgres.ParseUUID(projectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	var typeParam *sqlc.NotificationType
	if notificationType != nil {
		value := sqlcNotificationType(*notificationType)
		typeParam = &value
	}
	rows, err := r.queries.ListNotifications(ctx, sqlc.ListNotificationsParams{ProjectID: projectUUID, NotificationType: typeParam})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	notifications := make([]domainalert.Notification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, mapNotification(row))
	}
	return notifications, nil
}

func (r *Repository) GetNotification(ctx context.Context, projectID, notificationID string) (domainalert.Notification, error) {
	projectUUID, notificationUUID, err := parseProjectScopedID(projectID, notificationID, domainalert.ErrNotificationNotFound)
	if err != nil {
		return domainalert.Notification{}, err
	}
	row, err := r.queries.GetNotification(ctx, sqlc.GetNotificationParams{ProjectID: projectUUID, ID: notificationUUID})
	if err != nil {
		return domainalert.Notification{}, mapNoRows(err, domainalert.ErrNotificationNotFound)
	}
	return mapNotification(row), nil
}

func (r *Repository) CreateNotification(ctx context.Context, input domainalert.Notification) (domainalert.Notification, error) {
	params, err := createNotificationParams(input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	row, err := r.queries.CreateNotification(ctx, params)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return mapNotification(row), nil
}

func (r *Repository) UpdateNotification(ctx context.Context, input domainalert.Notification) (domainalert.Notification, error) {
	params, err := updateNotificationParams(input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	row, err := r.queries.UpdateNotification(ctx, params)
	if err != nil {
		return domainalert.Notification{}, mapNoRows(err, domainalert.ErrNotificationNotFound)
	}
	return mapNotification(row), nil
}

func (r *Repository) DeleteNotification(ctx context.Context, projectID, notificationID string) error {
	projectUUID, notificationUUID, err := parseProjectScopedID(projectID, notificationID, domainalert.ErrNotificationNotFound)
	if err != nil {
		return err
	}
	rows, err := r.queries.SoftDeleteNotification(ctx, sqlc.SoftDeleteNotificationParams{ProjectID: projectUUID, ID: notificationUUID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domainalert.ErrNotificationNotFound
	}
	return nil
}

func (r *Repository) ListIncidents(ctx context.Context, projectID string, status *domainalert.IncidentStatus, limit int32) ([]domainalert.Incident, error) {
	projectUUID, err := postgres.ParseUUID(projectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var statusParam *sqlc.AlertIncidentStatus
	if status != nil {
		value := sqlc.AlertIncidentStatus(*status)
		statusParam = &value
	}
	rows, err := r.queries.ListAlertIncidents(ctx, sqlc.ListAlertIncidentsParams{ProjectID: projectUUID, Status: statusParam, LimitCount: limit})
	if err != nil {
		return nil, err
	}
	incidents := make([]domainalert.Incident, 0, len(rows))
	for _, row := range rows {
		incidents = append(incidents, mapListIncident(row))
	}
	return incidents, nil
}

func (r *Repository) GetIncident(ctx context.Context, projectID, incidentID string) (domainalert.Incident, error) {
	projectUUID, incidentUUID, err := parseProjectScopedID(projectID, incidentID, domainalert.ErrIncidentNotFound)
	if err != nil {
		return domainalert.Incident{}, err
	}
	row, err := r.queries.GetAlertIncident(ctx, sqlc.GetAlertIncidentParams{ProjectID: projectUUID, ID: incidentUUID})
	if err != nil {
		return domainalert.Incident{}, mapNoRows(err, domainalert.ErrIncidentNotFound)
	}
	return mapGetIncident(row), nil
}

func (r *Repository) ListEnabledRulesForAssignment(ctx context.Context, projectID, probeID, checkID string, checkType domaincheck.Type) ([]domainalert.Rule, error) {
	projectUUID, probeUUID, checkUUID, err := parseAssignmentIDs(projectID, probeID, checkID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListEnabledAlertRulesForAssignment(ctx, sqlc.ListEnabledAlertRulesForAssignmentParams{
		ProjectID: projectUUID,
		CheckType: sqlcCheckType(checkType),
		ProbeID:   &probeUUID,
		CheckID:   &checkUUID,
	})
	if err != nil {
		return nil, err
	}
	return r.mapRulesWithNotifications(ctx, projectUUID, rows)
}

func (r *Repository) GetMetricSummary(ctx context.Context, metric string, probeStorageID, checkStorageID int64, from, to time.Time) (alertcondition.MetricSummary, error) {
	switch {
	case alertcondition.CompatibleWithCheckType(metric, domaincheck.TypePing):
		row, err := r.queries.GetPingAlertMetricSummary(ctx, sqlc.GetPingAlertMetricSummaryParams{
			Metric: metric, ProbeStorageID: probeStorageID, CheckStorageID: checkStorageID, StartedAtFrom: from.UTC(), StartedAtTo: to.UTC(),
		})
		if err != nil {
			return alertcondition.MetricSummary{}, err
		}
		return alertcondition.MetricSummary{Metric: metric, WindowStart: from.UTC(), WindowEnd: to.UTC(), Samples: row.Samples, Value: row.Value, HasValue: row.Samples > 0}, nil
	case alertcondition.CompatibleWithCheckType(metric, domaincheck.TypeTCP):
		row, err := r.queries.GetTCPAlertMetricSummary(ctx, sqlc.GetTCPAlertMetricSummaryParams{
			Metric: metric, ProbeStorageID: probeStorageID, CheckStorageID: checkStorageID, StartedAtFrom: from.UTC(), StartedAtTo: to.UTC(),
		})
		if err != nil {
			return alertcondition.MetricSummary{}, err
		}
		return alertcondition.MetricSummary{Metric: metric, WindowStart: from.UTC(), WindowEnd: to.UTC(), Samples: row.Samples, Value: row.Value, HasValue: row.Samples > 0}, nil
	case alertcondition.CompatibleWithCheckType(metric, domaincheck.TypeHTTP):
		row, err := r.queries.GetHTTPAlertMetricSummary(ctx, sqlc.GetHTTPAlertMetricSummaryParams{Metric: metric, ProbeStorageID: probeStorageID, CheckStorageID: checkStorageID, StartedAtFrom: from.UTC(), StartedAtTo: to.UTC()})
		if err != nil {
			return alertcondition.MetricSummary{}, err
		}
		return alertcondition.MetricSummary{Metric: metric, WindowStart: from.UTC(), WindowEnd: to.UTC(), Samples: row.Samples, Value: row.Value, HasValue: row.Samples > 0}, nil
	default:
		return alertcondition.MetricSummary{}, fmt.Errorf("unsupported metric: %s", metric)
	}
}

func (r *Repository) GetActiveIncident(ctx context.Context, ruleID, probeID, checkID string) (domainalert.Incident, error) {
	_, ruleUUID, probeUUID, checkUUID, err := parseRuleTargetIDs(ruleID, probeID, checkID)
	if err != nil {
		return domainalert.Incident{}, err
	}
	row, err := r.queries.GetActiveAlertIncident(ctx, sqlc.GetActiveAlertIncidentParams{RuleID: ruleUUID, ProbeID: probeUUID, CheckID: checkUUID})
	if err != nil {
		return domainalert.Incident{}, mapNoRows(err, domainalert.ErrIncidentNotFound)
	}
	return mapIncident(row), nil
}

func (r *Repository) GetRecentResolvedIncident(ctx context.Context, ruleID, probeID, checkID string, resolvedAfter time.Time) (domainalert.Incident, error) {
	_, ruleUUID, probeUUID, checkUUID, err := parseRuleTargetIDs(ruleID, probeID, checkID)
	if err != nil {
		return domainalert.Incident{}, err
	}
	resolvedAfterUTC := resolvedAfter.UTC()
	row, err := r.queries.GetRecentResolvedAlertIncident(ctx, sqlc.GetRecentResolvedAlertIncidentParams{
		RuleID:        ruleUUID,
		ProbeID:       probeUUID,
		CheckID:       checkUUID,
		ResolvedAfter: &resolvedAfterUTC,
	})
	if err != nil {
		return domainalert.Incident{}, mapNoRows(err, domainalert.ErrIncidentNotFound)
	}
	return mapIncident(row), nil
}

func (r *Repository) CreateIncident(ctx context.Context, input domainalert.IncidentTransitionInput) (domainalert.Incident, error) {
	var incident domainalert.Incident
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		params, err := createIncidentParams(input)
		if err != nil {
			return err
		}
		row, err := q.CreateAlertIncident(ctx, params)
		if err != nil {
			return mapNoRows(err, domainalert.ErrIncidentNotFound)
		}
		incident = mapIncident(row)
		incident.Probe = input.Probe
		incident.Check = input.Check
		return nil
	})
	if err != nil {
		return domainalert.Incident{}, err
	}
	return incident, nil
}

func (r *Repository) EnqueueNotificationJobs(ctx context.Context, jobs []domainalert.NotificationJobInput) error {
	q := postgres.Queries(ctx, r.queries)
	for _, job := range jobs {
		if err := enqueueJob(ctx, q, job); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpdateIncidentTriggered(ctx context.Context, incidentID string, evaluation alertcondition.Evaluation, summary json.RawMessage, at time.Time) (domainalert.Incident, error) {
	id, err := postgres.ParseUUID(incidentID, domainalert.ErrIncidentNotFound)
	if err != nil {
		return domainalert.Incident{}, err
	}
	row, err := r.queries.UpdateActiveAlertIncidentTriggered(ctx, sqlc.UpdateActiveAlertIncidentTriggeredParams{
		ID:              id,
		LastEvaluatedAt: at.UTC(),
		LastTriggeredAt: at.UTC(),
		LastValue:       &evaluation.Value,
		LastSummary:     summary,
	})
	if err != nil {
		return domainalert.Incident{}, mapNoRows(err, domainalert.ErrIncidentNotFound)
	}
	return mapIncident(row), nil
}

func (r *Repository) UpdateIncidentInsufficient(ctx context.Context, incidentID string, state alertcondition.EvaluationState, summary json.RawMessage, at time.Time) (domainalert.Incident, error) {
	id, err := postgres.ParseUUID(incidentID, domainalert.ErrIncidentNotFound)
	if err != nil {
		return domainalert.Incident{}, err
	}
	row, err := r.queries.UpdateActiveAlertIncidentInsufficient(ctx, sqlc.UpdateActiveAlertIncidentInsufficientParams{
		ID:                  id,
		LastEvaluationState: sqlcEvaluationState(state),
		LastEvaluatedAt:     at.UTC(),
		LastSummary:         summary,
	})
	if err != nil {
		return domainalert.Incident{}, mapNoRows(err, domainalert.ErrIncidentNotFound)
	}
	return mapIncident(row), nil
}

func (r *Repository) ResolveIncident(ctx context.Context, incidentID string, summary json.RawMessage, at time.Time) (domainalert.Incident, error) {
	id, err := postgres.ParseUUID(incidentID, domainalert.ErrIncidentNotFound)
	if err != nil {
		return domainalert.Incident{}, err
	}
	var incident domainalert.Incident
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		resolvedAt := at.UTC()
		row, resolveErr := q.ResolveActiveAlertIncident(ctx, sqlc.ResolveActiveAlertIncidentParams{
			ID:              id,
			ResolvedAt:      &resolvedAt,
			LastEvaluatedAt: at.UTC(),
			LastSummary:     summary,
		})
		if resolveErr != nil {
			return mapNoRows(resolveErr, domainalert.ErrIncidentNotFound)
		}
		incident = mapIncident(row)
		return nil
	})
	if err != nil {
		return domainalert.Incident{}, err
	}
	return incident, nil
}

func (r *Repository) ListEnabledNotificationsForRule(ctx context.Context, projectID, ruleID string) ([]domainalert.Notification, error) {
	projectUUID, ruleUUID, err := parseProjectScopedID(projectID, ruleID, domainalert.ErrRuleNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := postgres.Queries(ctx, r.queries).ListEnabledNotificationsForRule(ctx, sqlc.ListEnabledNotificationsForRuleParams{ProjectID: projectUUID, RuleID: ruleUUID})
	if err != nil {
		return nil, err
	}
	notifications := make([]domainalert.Notification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, mapNotification(row))
	}
	return notifications, nil
}

func (r *Repository) ClaimOutbox(ctx context.Context, limit int32, staleBefore time.Time) ([]domainalert.NotificationOutboxJob, error) {
	stale := staleBefore.UTC()
	if _, err := r.queries.RecoverStaleNotificationOutbox(ctx, &stale); err != nil {
		return nil, err
	}
	rows, err := r.queries.ClaimNotificationOutbox(ctx, limit)
	if err != nil {
		return nil, err
	}
	jobs := make([]domainalert.NotificationOutboxJob, 0, len(rows))
	for _, row := range rows {
		jobs = append(jobs, mapOutbox(row))
	}
	return jobs, nil
}

func (r *Repository) MarkOutboxDelivered(ctx context.Context, id string, at time.Time) error {
	uuidValue, err := postgres.ParseUUID(id, domainalert.ErrInvalidInput)
	if err != nil {
		return err
	}
	deliveredAt := at.UTC()
	return r.queries.MarkNotificationOutboxDelivered(ctx, sqlc.MarkNotificationOutboxDeliveredParams{ID: uuidValue, DeliveredAt: &deliveredAt})
}

func (r *Repository) MarkOutboxRetry(ctx context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainalert.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkNotificationOutboxRetry(ctx, sqlc.MarkNotificationOutboxRetryParams{
		ID: uuidValue, NextAttemptAt: nextAttemptAt.UTC(), LastErrorKind: kindPtr, LastErrorCode: codePtr, LastError: messagePtr,
	})
}

func (r *Repository) MarkOutboxFailed(ctx context.Context, id, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainalert.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkNotificationOutboxFailed(ctx, sqlc.MarkNotificationOutboxFailedParams{
		ID: uuidValue, LastErrorKind: kindPtr, LastErrorCode: codePtr, LastError: messagePtr,
	})
}

func (r *Repository) MarkOutboxDiscarded(ctx context.Context, id, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainalert.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkNotificationOutboxDiscarded(ctx, sqlc.MarkNotificationOutboxDiscardedParams{
		ID: uuidValue, LastErrorKind: kindPtr, LastErrorCode: codePtr, LastError: messagePtr,
	})
}

func (r *Repository) mapRulesWithNotifications(ctx context.Context, projectUUID uuid.UUID, rows []sqlc.AlertRule) ([]domainalert.Rule, error) {
	rules := make([]domainalert.Rule, 0, len(rows))
	for _, row := range rows {
		notificationIDs, err := r.ruleNotificationIDs(ctx, projectUUID, row.ID)
		if err != nil {
			return nil, err
		}
		rules = append(rules, mapRule(row, notificationIDs))
	}
	return rules, nil
}

func (r *Repository) ruleNotificationIDs(ctx context.Context, projectUUID, ruleUUID uuid.UUID) ([]string, error) {
	rows, err := r.queries.ListAlertNotificationIDs(ctx, sqlc.ListAlertNotificationIDsParams{ProjectID: projectUUID, RuleIds: []uuid.UUID{ruleUUID}})
	if err != nil {
		return nil, err
	}
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		values = append(values, row.String())
	}
	return values, nil
}

func createRuleParams(input domainalert.Rule) (sqlc.CreateAlertRuleParams, error) {
	projectUUID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return sqlc.CreateAlertRuleParams{}, err
	}
	createdByUUID, err := uuid.Parse(input.CreatedByUserID)
	if err != nil {
		return sqlc.CreateAlertRuleParams{}, err
	}
	return sqlc.CreateAlertRuleParams{
		ProjectID:        projectUUID,
		Name:             input.Name,
		Description:      input.Description,
		Status:           sqlcRuleStatus(input.Status),
		Severity:         sqlcSeverity(input.Severity),
		CheckType:        sqlcCheckType(input.CheckType),
		ProbeID:          optionalUUID(input.ProbeID),
		CheckID:          optionalUUID(input.CheckID),
		ProbeSelector:    input.ProbeSelector,
		Condition:        input.ConditionJSON,
		ConditionVersion: input.ConditionVersion,
		CooldownSeconds:  input.CooldownSeconds,
		CreatedByUserID:  createdByUUID,
	}, nil
}

func updateRuleParams(input domainalert.Rule) (sqlc.UpdateAlertRuleParams, error) {
	projectUUID, ruleUUID, err := parseProjectScopedID(input.ProjectID, input.ID, domainalert.ErrRuleNotFound)
	if err != nil {
		return sqlc.UpdateAlertRuleParams{}, err
	}
	return sqlc.UpdateAlertRuleParams{
		ProjectID:        projectUUID,
		ID:               ruleUUID,
		Name:             input.Name,
		Description:      input.Description,
		Status:           sqlcRuleStatus(input.Status),
		Severity:         sqlcSeverity(input.Severity),
		CheckType:        sqlcCheckType(input.CheckType),
		ProbeID:          optionalUUID(input.ProbeID),
		CheckID:          optionalUUID(input.CheckID),
		ProbeSelector:    input.ProbeSelector,
		Condition:        input.ConditionJSON,
		ConditionVersion: input.ConditionVersion,
		CooldownSeconds:  input.CooldownSeconds,
	}, nil
}

func createNotificationParams(input domainalert.Notification) (sqlc.CreateNotificationParams, error) {
	projectUUID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return sqlc.CreateNotificationParams{}, err
	}
	createdByUUID, err := uuid.Parse(input.CreatedByUserID)
	if err != nil {
		return sqlc.CreateNotificationParams{}, err
	}
	return sqlc.CreateNotificationParams{
		ProjectID:        projectUUID,
		Name:             input.Name,
		NotificationType: sqlcNotificationType(input.Type),
		Enabled:          input.Enabled,
		Config:           input.Config,
		CreatedByUserID:  createdByUUID,
	}, nil
}

func updateNotificationParams(input domainalert.Notification) (sqlc.UpdateNotificationParams, error) {
	projectUUID, notificationUUID, err := parseProjectScopedID(input.ProjectID, input.ID, domainalert.ErrNotificationNotFound)
	if err != nil {
		return sqlc.UpdateNotificationParams{}, err
	}
	return sqlc.UpdateNotificationParams{
		ProjectID:        projectUUID,
		ID:               notificationUUID,
		Name:             input.Name,
		NotificationType: sqlcNotificationType(input.Type),
		Enabled:          input.Enabled,
		Config:           input.Config,
	}, nil
}

func createIncidentParams(input domainalert.IncidentTransitionInput) (sqlc.CreateAlertIncidentParams, error) {
	projectUUID, ruleUUID, probeUUID, checkUUID, err := parseTransitionIDs(input)
	if err != nil {
		return sqlc.CreateAlertIncidentParams{}, err
	}
	return sqlc.CreateAlertIncidentParams{
		ProjectID:                  projectUUID,
		RuleID:                     ruleUUID,
		ProbeID:                    probeUUID,
		CheckID:                    checkUUID,
		CheckType:                  sqlcCheckType(input.CheckType),
		Severity:                   sqlcSeverity(input.Rule.Severity),
		LastEvaluationState:        sqlcEvaluationState(input.Evaluation.State),
		OpenedAt:                   input.At.UTC(),
		LastEvaluatedAt:            input.At.UTC(),
		LastTriggeredAt:            input.At.UTC(),
		LastValue:                  &input.Evaluation.Value,
		LastSummary:                input.Summary,
		LastNotificationSentAt:     input.LastNotificationSentAt,
		NextNotificationEligibleAt: input.NextNotificationEligibleAt,
	}, nil
}

func enqueueJob(ctx context.Context, q *sqlc.Queries, input domainalert.NotificationJobInput) error {
	projectUUID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return err
	}
	incidentUUID, err := postgres.ParseUUID(input.IncidentID, domainalert.ErrIncidentNotFound)
	if err != nil {
		return err
	}
	ruleUUID, err := postgres.ParseUUID(input.RuleID, domainalert.ErrRuleNotFound)
	if err != nil {
		return err
	}
	notificationUUID, err := postgres.ParseUUID(input.NotificationID, domainalert.ErrNotificationNotFound)
	if err != nil {
		return err
	}
	_, err = q.EnqueueNotificationOutbox(ctx, sqlc.EnqueueNotificationOutboxParams{
		ProjectID: projectUUID, IncidentID: incidentUUID, RuleID: ruleUUID, NotificationID: notificationUUID,
		NotificationType: sqlcNotificationType(input.NotificationType), EventType: input.EventType, Payload: input.Payload, DedupeKey: input.DedupeKey,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func optionalUUID(value *string) *uuid.UUID {
	if value == nil {
		return nil
	}
	parsed, err := uuid.Parse(*value)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseProjectScopedID(projectID, value string, wrap error) (uuid.UUID, uuid.UUID, error) {
	projectUUID, err := postgres.ParseUUID(projectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	valueUUID, err := postgres.ParseUUID(value, wrap)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return projectUUID, valueUUID, nil
}

func parseAssignmentIDs(projectID, probeID, checkID string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectUUID, err := postgres.ParseUUID(projectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	probeUUID, err := postgres.ParseUUID(probeID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	checkUUID, err := postgres.ParseUUID(checkID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectUUID, probeUUID, checkUUID, nil
}

func parseRuleTargetIDs(ruleID, probeID, checkID string) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, error) {
	ruleUUID, err := postgres.ParseUUID(ruleID, domainalert.ErrRuleNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	probeUUID, err := postgres.ParseUUID(probeID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	checkUUID, err := postgres.ParseUUID(checkID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return uuid.Nil, ruleUUID, probeUUID, checkUUID, nil
}

func parseTransitionIDs(input domainalert.IncidentTransitionInput) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectUUID, ruleUUID, err := parseProjectScopedID(input.Rule.ProjectID, input.Rule.ID, domainalert.ErrRuleNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	probeUUID, err := postgres.ParseUUID(input.ProbeID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	checkUUID, err := postgres.ParseUUID(input.CheckID, domainalert.ErrInvalidInput)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectUUID, ruleUUID, probeUUID, checkUUID, nil
}

func mapNoRows(err, wrap error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return wrap
	}
	return err
}
