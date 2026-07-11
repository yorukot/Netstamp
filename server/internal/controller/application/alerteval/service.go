package alerteval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

const (
	EventIncidentOpened   = "incident.opened"
	EventIncidentResolved = "incident.resolved"
)

type Service struct {
	repo                   Repository
	enabled                bool
	backendBaseURL         string
	events                 EventRecorder
	backendBaseURLProvider BackendBaseURLProvider
	tx                     Transactor
	now                    func() time.Time
}

type BackendBaseURLProvider interface {
	BackendBaseURL(ctx context.Context) (string, error)
}

func NewService(repo Repository, enabled bool, backendBaseURL string, transactors ...Transactor) *Service {
	return NewServiceWithEvents(repo, enabled, backendBaseURL, nil, transactors...)
}

func NewServiceWithEvents(repo Repository, enabled bool, backendBaseURL string, events EventRecorder, transactors ...Transactor) *Service {
	tx := Transactor(apptx.NoopTransactor{})
	if len(transactors) > 0 && transactors[0] != nil {
		tx = transactors[0]
	}

	return &Service{
		repo:           repo,
		enabled:        enabled,
		backendBaseURL: strings.TrimRight(strings.TrimSpace(backendBaseURL), "/"),
		events:         events,
		tx:             tx,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) ConfigureBackendBaseURLProvider(provider BackendBaseURLProvider) {
	s.backendBaseURLProvider = provider
}

func (s *Service) EvaluateChangedAssignments(ctx context.Context, inputs []appproberuntime.ChangedAssignmentInput) error {
	if !s.enabled {
		return nil
	}
	var errs []error
	for _, input := range inputs {
		if err := s.evaluateAssignment(ctx, input); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Service) evaluateAssignment(ctx context.Context, input appproberuntime.ChangedAssignmentInput) error {
	ctx, flow := s.startAssignmentFlow(ctx, input)
	defer flow.end()

	checkType, err := domaincheck.VNCheckType(domaincheck.Type(input.CheckType))
	if err != nil {
		return flow.failure(AlertEvalEventAssignmentEvaluateFailure, AlertEvalReasonInvalidCheckType, err)
	}
	if checkType != domaincheck.TypePing && checkType != domaincheck.TypeTCP && checkType != domaincheck.TypeHTTP {
		flow.success()
		return nil
	}
	rules, err := s.repo.ListEnabledRulesForAssignment(ctx, input.ProjectID, input.ProbeID, input.CheckID, checkType)
	if err != nil {
		return flow.failure(AlertEvalEventAssignmentEvaluateFailure, AlertEvalReasonRuleListFailed, err)
	}
	var errs []error
	for _, rule := range rules {
		if err := s.evaluateRule(ctx, input, rule); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return flow.failure(AlertEvalEventAssignmentEvaluateFailure, AlertEvalReasonRuleEvaluateFailed, err)
	}
	flow.success()
	return nil
}

func (s *Service) evaluateRule(ctx context.Context, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule) error {
	ctx, flow := s.startRuleFlow(ctx, input, rule)
	defer flow.end()

	now := s.now()
	requirement := rule.Condition.Requirement()
	windowEnd := now
	windowStart := windowEnd.Add(-time.Duration(requirement.WindowSeconds) * time.Second)
	summary, err := s.repo.GetMetricSummary(ctx, requirement.Metric, input.ProbeStorageID, input.CheckStorageID, windowStart, windowEnd)
	if err != nil {
		return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonMetricSummaryFailed, err)
	}
	summary.MinSamples = requirement.MinSamples
	summary.WindowSeconds = requirement.WindowSeconds
	evaluation := rule.Condition.Evaluate(summary)
	summaryJSON, err := evaluationSummaryJSON(rule, evaluation)
	if err != nil {
		return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonEvaluationSummaryFailed, err)
	}

	active, activeErr := s.repo.GetActiveIncident(ctx, rule.ID, input.ProbeID, input.CheckID)
	if activeErr != nil && !errors.Is(activeErr, domainalert.ErrIncidentNotFound) {
		return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentLookupFailed, activeErr)
	}
	if activeErr == nil {
		flow.setIncidentID(active.ID)
	}

	if err := s.handleEvaluationState(ctx, flow, input, rule, evaluation, summaryJSON, active, activeErr == nil, now); err != nil {
		return err
	}
	flow.success()
	return nil
}

func (s *Service) handleEvaluationState(ctx context.Context, flow *alertEvalFlow, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule, evaluation alertcondition.Evaluation, summaryJSON []byte, active domainalert.Incident, hasActive bool, now time.Time) error {
	switch evaluation.State {
	case alertcondition.EvaluationStateFiring:
		return s.handleFiringEvaluation(ctx, flow, input, rule, evaluation, summaryJSON, active, hasActive, now)
	case alertcondition.EvaluationStateClear:
		return s.handleClearEvaluation(ctx, flow, input, rule, evaluation, summaryJSON, active, hasActive, now)
	case alertcondition.EvaluationStateInsufficientSamples, alertcondition.EvaluationStateNoData:
		return s.handleInsufficientEvaluation(ctx, flow, active, hasActive, evaluation, summaryJSON, now)
	default:
		return nil
	}
}

func (s *Service) handleFiringEvaluation(ctx context.Context, flow *alertEvalFlow, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule, evaluation alertcondition.Evaluation, summaryJSON []byte, active domainalert.Incident, hasActive bool, now time.Time) error {
	if hasActive {
		incident, err := s.repo.UpdateIncidentTriggered(ctx, active.ID, evaluation, summaryJSON, now)
		if err != nil {
			return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentTransitionFailed, err)
		}
		flow.setIncidentID(incident.ID)
		return nil
	}
	inCooldown, err := s.inReopenCooldown(ctx, flow, rule, input, now)
	if err != nil {
		return err
	}
	if inCooldown {
		return nil
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, err := s.repo.CreateIncident(ctx, domainalert.IncidentTransitionInput{
			Rule:                       rule,
			ProbeID:                    input.ProbeID,
			CheckID:                    input.CheckID,
			Probe:                      incidentProbeSummary(input),
			Check:                      incidentCheckSummary(input),
			CheckType:                  domaincheck.Type(input.CheckType),
			Evaluation:                 evaluation,
			Summary:                    summaryJSON,
			At:                         now,
			LastNotificationSentAt:     &now,
			NextNotificationEligibleAt: ptrTime(now.Add(time.Duration(rule.CooldownSeconds) * time.Second)),
		})
		if err != nil {
			return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentTransitionFailed, err)
		}
		created = incidentWithAssignmentContext(created, input)
		flow.setIncidentID(created.ID)
		return s.enqueueNotifications(ctx, flow, rule, created, evaluation, EventIncidentOpened, now)
	})
}

func (s *Service) handleClearEvaluation(ctx context.Context, flow *alertEvalFlow, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule, evaluation alertcondition.Evaluation, summaryJSON []byte, active domainalert.Incident, hasActive bool, now time.Time) error {
	if !hasActive {
		return nil
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		incident, err := s.repo.ResolveIncident(ctx, active.ID, summaryJSON, now)
		if err != nil {
			return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentTransitionFailed, err)
		}
		incident = incidentWithAssignmentContext(incident, input)
		flow.setIncidentID(incident.ID)
		return s.enqueueNotifications(ctx, flow, rule, incident, evaluation, EventIncidentResolved, now)
	})
}

func (s *Service) handleInsufficientEvaluation(ctx context.Context, flow *alertEvalFlow, active domainalert.Incident, hasActive bool, evaluation alertcondition.Evaluation, summaryJSON []byte, now time.Time) error {
	if !hasActive {
		return nil
	}
	incident, err := s.repo.UpdateIncidentInsufficient(ctx, active.ID, evaluation.State, summaryJSON, now)
	if err != nil {
		return flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentTransitionFailed, err)
	}
	flow.setIncidentID(incident.ID)
	return nil
}

func (s *Service) inReopenCooldown(ctx context.Context, flow *alertEvalFlow, rule domainalert.Rule, input appproberuntime.ChangedAssignmentInput, now time.Time) (bool, error) {
	if rule.CooldownSeconds <= 0 {
		return false, nil
	}
	_, err := s.repo.GetRecentResolvedIncident(ctx, rule.ID, input.ProbeID, input.CheckID, now.Add(-time.Duration(rule.CooldownSeconds)*time.Second))
	if err != nil && !errors.Is(err, domainalert.ErrIncidentNotFound) {
		return false, flow.failure(AlertEvalEventRuleEvaluateFailure, AlertEvalReasonIncidentLookupFailed, err)
	}
	return err == nil, nil
}

func (s *Service) enqueueNotifications(ctx context.Context, flow *alertEvalFlow, rule domainalert.Rule, incident domainalert.Incident, evaluation alertcondition.Evaluation, eventType string, at time.Time) error {
	notifications, err := s.repo.ListEnabledNotificationsForRule(ctx, rule.ProjectID, rule.ID)
	if err != nil {
		return flow.failure(AlertEvalEventNotificationEnqueueFail, AlertEvalReasonNotificationListFailed, err)
	}
	jobs := make([]domainalert.NotificationJobInput, 0, len(notifications))
	backendBaseURL := s.effectiveBackendBaseURL(ctx)
	for _, notification := range notifications {
		if !supportedNotification(notification.Type) {
			continue
		}
		payload, err := notificationPayload(rule, incident, notification, evaluation, eventType, at, backendBaseURL)
		if err != nil {
			return flow.failure(AlertEvalEventNotificationEnqueueFail, AlertEvalReasonNotificationPayloadFail, err)
		}
		jobs = append(jobs, domainalert.NotificationJobInput{
			ProjectID:        rule.ProjectID,
			IncidentID:       incident.ID,
			RuleID:           rule.ID,
			NotificationID:   notification.ID,
			NotificationType: notification.Type,
			EventType:        eventType,
			Payload:          payload,
			DedupeKey:        fmt.Sprintf("%s:%s:%s:%s", incident.ID, notification.ID, eventType, at.UTC().Format(time.RFC3339Nano)),
		})
	}
	if err := s.repo.EnqueueNotificationJobs(ctx, jobs); err != nil {
		return flow.failure(AlertEvalEventNotificationEnqueueFail, AlertEvalReasonNotificationEnqueueFail, err)
	}
	return nil
}

func (s *Service) effectiveBackendBaseURL(ctx context.Context) string {
	if s.backendBaseURLProvider == nil {
		return s.backendBaseURL
	}
	value, err := s.backendBaseURLProvider.BackendBaseURL(ctx)
	if err != nil {
		return s.backendBaseURL
	}
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return s.backendBaseURL
	}
	return value
}

func supportedNotification(notificationType domainalert.NotificationType) bool {
	switch notificationType {
	case domainalert.NotificationTypeWebhook, domainalert.NotificationTypeSlack, domainalert.NotificationTypeDiscord, domainalert.NotificationTypeTelegram, domainalert.NotificationTypeEmail:
		return true
	default:
		return false
	}
}

func evaluationSummaryJSON(rule domainalert.Rule, evaluation alertcondition.Evaluation) (json.RawMessage, error) {
	payload := map[string]any{
		"state":         evaluation.State,
		"metric":        rule.Condition.Metric,
		"operator":      rule.Condition.Operator,
		"threshold":     rule.Condition.Threshold,
		"value":         evaluation.Value,
		"samples":       evaluation.Summary.Samples,
		"minSamples":    rule.Condition.MinSamples,
		"windowSeconds": rule.Condition.WindowSeconds,
		"windowStart":   evaluation.Summary.WindowStart,
		"windowEnd":     evaluation.Summary.WindowEnd,
	}
	data, err := json.Marshal(payload)
	return data, err
}

func notificationPayload(rule domainalert.Rule, incident domainalert.Incident, notification domainalert.Notification, evaluation alertcondition.Evaluation, eventType string, at time.Time, backendBaseURL string) (json.RawMessage, error) {
	incidentPayload := map[string]any{
		"id":       incident.ID,
		"status":   incident.Status,
		"openedAt": incident.OpenedAt,
	}
	probePayload := map[string]any{"id": incident.ProbeID}
	if incident.Probe != nil {
		probePayload["name"] = incident.Probe.Name
	}
	checkPayload := map[string]any{
		"id":   incident.CheckID,
		"type": incident.CheckType,
	}
	if incident.Check != nil {
		checkPayload["name"] = incident.Check.Name
		checkPayload["target"] = incident.Check.Target
		checkPayload["type"] = incident.Check.Type
	}
	payload := map[string]any{
		"eventType": eventType,
		"sentAt":    at.UTC(),
		"rule": map[string]any{
			"id":       rule.ID,
			"name":     rule.Name,
			"severity": rule.Severity,
		},
		"incident": incidentPayload,
		"target": map[string]any{
			"probeId":   incident.ProbeID,
			"checkId":   incident.CheckID,
			"checkType": incident.CheckType,
			"probe":     probePayload,
			"check":     checkPayload,
		},
		"notification": map[string]any{
			"id":   notification.ID,
			"name": notification.Name,
			"type": notification.Type,
		},
		"summary": map[string]any{
			"state":         evaluation.State,
			"metric":        rule.Condition.Metric,
			"value":         evaluation.Value,
			"threshold":     rule.Condition.Threshold,
			"operator":      rule.Condition.Operator,
			"windowSeconds": rule.Condition.WindowSeconds,
			"samples":       evaluation.Summary.Samples,
			"minSamples":    rule.Condition.MinSamples,
		},
	}
	if incident.ResolvedAt != nil {
		incidentPayload["resolvedAt"] = *incident.ResolvedAt
	}
	if incidentURL := alertIncidentURL(backendBaseURL, incident.ProjectID, incident.ID); incidentURL != "" {
		payload["links"] = map[string]any{"incident": incidentURL}
	}
	data, err := json.Marshal(payload)
	return data, err
}

func alertIncidentURL(backendBaseURL, projectID, incidentID string) string {
	backendBaseURL = strings.TrimRight(strings.TrimSpace(backendBaseURL), "/")
	if backendBaseURL == "" || projectID == "" || incidentID == "" {
		return ""
	}
	return backendBaseURL + "/projects/" + url.PathEscape(projectID) + "/alerts/incident/" + url.PathEscape(incidentID)
}

func incidentProbeSummary(input appproberuntime.ChangedAssignmentInput) *domainalert.IncidentProbeSummary {
	if input.ProbeName == "" {
		return nil
	}
	return &domainalert.IncidentProbeSummary{ID: input.ProbeID, Name: input.ProbeName}
}

func incidentCheckSummary(input appproberuntime.ChangedAssignmentInput) *domainalert.IncidentCheckSummary {
	if input.CheckName == "" && input.CheckTarget == "" {
		return nil
	}
	return &domainalert.IncidentCheckSummary{ID: input.CheckID, Name: input.CheckName, Type: domaincheck.Type(input.CheckType), Target: input.CheckTarget}
}

func incidentWithAssignmentContext(incident domainalert.Incident, input appproberuntime.ChangedAssignmentInput) domainalert.Incident {
	if incident.Probe == nil {
		incident.Probe = incidentProbeSummary(input)
	}
	if incident.Check == nil {
		incident.Check = incidentCheckSummary(input)
	}
	return incident
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
