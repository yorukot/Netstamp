package alerteval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

const (
	EventIncidentOpened   = "incident.opened"
	EventIncidentResolved = "incident.resolved"
)

type Service struct {
	repo    Repository
	enabled bool
	now     func() time.Time
}

func NewService(repo Repository, enabled bool) *Service {
	return &Service{repo: repo, enabled: enabled, now: func() time.Time { return time.Now().UTC() }}
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
	checkType, err := domaincheck.VNCheckType(domaincheck.Type(input.CheckType))
	if err != nil {
		return err
	}
	if checkType != domaincheck.TypePing && checkType != domaincheck.TypeTCP {
		return nil
	}
	rules, err := s.repo.ListEnabledRulesForAssignment(ctx, input.ProjectID, input.ProbeID, input.CheckID, checkType)
	if err != nil {
		return err
	}
	var errs []error
	for _, rule := range rules {
		if err := s.evaluateRule(ctx, input, rule); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Service) evaluateRule(ctx context.Context, input appproberuntime.ChangedAssignmentInput, rule domainalert.Rule) error {
	now := s.now()
	requirement := rule.Condition.Requirement()
	windowEnd := now
	windowStart := windowEnd.Add(-time.Duration(requirement.WindowSeconds) * time.Second)
	summary, err := s.repo.GetMetricSummary(ctx, requirement.Metric, input.ProbeStorageID, input.CheckStorageID, windowStart, windowEnd)
	if err != nil {
		return err
	}
	summary.MinSamples = requirement.MinSamples
	summary.WindowSeconds = requirement.WindowSeconds
	evaluation := rule.Condition.Evaluate(summary)
	summaryJSON, err := evaluationSummaryJSON(rule, evaluation)
	if err != nil {
		return err
	}

	active, activeErr := s.repo.GetActiveIncident(ctx, rule.ID, input.ProbeID, input.CheckID)
	if activeErr != nil && !errors.Is(activeErr, domainalert.ErrIncidentNotFound) {
		return activeErr
	}
	hasActive := activeErr == nil

	switch evaluation.State {
	case alertcondition.EvaluationStateFiring:
		if hasActive {
			_, err := s.repo.UpdateIncidentTriggered(ctx, active.ID, evaluation, summaryJSON, now)
			return err
		}
		if s.inReopenCooldown(ctx, rule, input, now) {
			return nil
		}
		incident, err := s.repo.CreateIncidentAndEnqueue(ctx, domainalert.IncidentTransitionInput{
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
			return err
		}
		incident = incidentWithAssignmentContext(incident, input)
		return s.enqueueNotifications(ctx, rule, incident, evaluation, EventIncidentOpened, now)
	case alertcondition.EvaluationStateClear:
		if hasActive {
			incident, err := s.repo.ResolveIncidentAndEnqueue(ctx, active.ID, summaryJSON, now, nil)
			if err != nil {
				return err
			}
			incident = incidentWithAssignmentContext(incident, input)
			return s.enqueueNotifications(ctx, rule, incident, evaluation, EventIncidentResolved, now)
		}
	case alertcondition.EvaluationStateInsufficientSamples, alertcondition.EvaluationStateNoData:
		if hasActive {
			_, err := s.repo.UpdateIncidentInsufficient(ctx, active.ID, evaluation.State, summaryJSON, now)
			return err
		}
	}
	return nil
}

func (s *Service) inReopenCooldown(ctx context.Context, rule domainalert.Rule, input appproberuntime.ChangedAssignmentInput, now time.Time) bool {
	if rule.CooldownSeconds <= 0 {
		return false
	}
	_, err := s.repo.GetRecentResolvedIncident(ctx, rule.ID, input.ProbeID, input.CheckID, now.Add(-time.Duration(rule.CooldownSeconds)*time.Second))
	return err == nil
}

func (s *Service) enqueueNotifications(ctx context.Context, rule domainalert.Rule, incident domainalert.Incident, evaluation alertcondition.Evaluation, eventType string, at time.Time) error {
	channels, err := s.repo.ListEnabledChannelsForRule(ctx, rule.ProjectID, rule.ID)
	if err != nil {
		return err
	}
	jobs := make([]domainalert.NotificationJobInput, 0, len(channels))
	for _, channel := range channels {
		if !supportedNotificationChannel(channel.Type) {
			continue
		}
		payload, err := notificationPayload(rule, incident, evaluation, eventType, at)
		if err != nil {
			return err
		}
		jobs = append(jobs, domainalert.NotificationJobInput{
			ProjectID:   rule.ProjectID,
			IncidentID:  incident.ID,
			RuleID:      rule.ID,
			ChannelID:   channel.ID,
			ChannelType: channel.Type,
			EventType:   eventType,
			Payload:     payload,
			DedupeKey:   fmt.Sprintf("%s:%s:%s:%s", incident.ID, channel.ID, eventType, at.UTC().Format(time.RFC3339Nano)),
		})
	}
	return s.repo.EnqueueNotificationJobs(ctx, jobs)
}

func supportedNotificationChannel(channelType domainalert.ChannelType) bool {
	switch channelType {
	case domainalert.ChannelTypeWebhook, domainalert.ChannelTypeDiscord, domainalert.ChannelTypeTelegram:
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

func notificationPayload(rule domainalert.Rule, incident domainalert.Incident, evaluation alertcondition.Evaluation, eventType string, at time.Time) (json.RawMessage, error) {
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
	data, err := json.Marshal(payload)
	return data, err
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
