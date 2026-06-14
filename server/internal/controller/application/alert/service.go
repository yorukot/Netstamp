package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	channelTester ChannelTester
}

func NewService(repo Repository, projectAccess ProjectAccess, channelTester ChannelTester) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, channelTester: channelTester}
}

func (s *Service) ListRules(ctx context.Context, input ListRulesInput) ([]domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListRules(ctx, project.ID, input.Status, input.CheckType)
}

func (s *Service) GetRule(ctx context.Context, input GetRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.GetRule(ctx, project.ID, input.RuleID)
}

func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeCreateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.CreateRule(ctx, rule)
}

func (s *Service) UpdateRule(ctx context.Context, input UpdateRuleInput) (domainalert.Rule, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Rule{}, err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return domainalert.Rule{}, actionErr
	}
	rule, err := normalizeUpdateRule(project.ID, input)
	if err != nil {
		return domainalert.Rule{}, err
	}
	return s.repo.UpdateRule(ctx, rule)
}

func (s *Service) DeleteRule(ctx context.Context, input DeleteRuleInput) error {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireProjectAction(ctx, project.ID, input.CurrentUserID, domainproject.ActionManageAlerts); actionErr != nil {
		return actionErr
	}
	return s.repo.DeleteRule(ctx, project.ID, input.RuleID)
}

func (s *Service) ListChannels(ctx context.Context, input ListChannelsInput) ([]domainalert.NotificationChannel, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListChannels(ctx, project.ID, input.Type)
}

func (s *Service) GetChannel(ctx context.Context, input GetChannelInput) (domainalert.NotificationChannel, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.NotificationChannel{}, err
	}
	return s.repo.GetChannel(ctx, project.ID, input.ChannelID)
}

func (s *Service) CreateChannel(ctx context.Context, input CreateChannelInput) (domainalert.NotificationChannel, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.NotificationChannel{}, err
	}
	if actionErr := s.requireChannelWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.NotificationChannel{}, actionErr
	}
	channel, err := normalizeCreateChannel(project.ID, input)
	if err != nil {
		return domainalert.NotificationChannel{}, err
	}
	return s.repo.CreateChannel(ctx, channel)
}

func (s *Service) UpdateChannel(ctx context.Context, input UpdateChannelInput) (domainalert.NotificationChannel, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.NotificationChannel{}, err
	}
	if actionErr := s.requireChannelWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.NotificationChannel{}, actionErr
	}
	channel, err := normalizeUpdateChannel(project.ID, input)
	if err != nil {
		return domainalert.NotificationChannel{}, err
	}
	return s.repo.UpdateChannel(ctx, channel)
}

func (s *Service) DeleteChannel(ctx context.Context, input DeleteChannelInput) error {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireChannelWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return actionErr
	}
	return s.repo.DeleteChannel(ctx, project.ID, input.ChannelID)
}

func (s *Service) TestChannel(ctx context.Context, input TestChannelInput) (ChannelTestResult, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return ChannelTestResult{}, err
	}
	if actionErr := s.requireChannelWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return ChannelTestResult{}, actionErr
	}
	channel, err := s.repo.GetChannel(ctx, project.ID, input.ChannelID)
	if err != nil {
		return ChannelTestResult{}, err
	}
	if s.channelTester == nil {
		return ChannelTestResult{Kind: "channel", Code: "tester_unavailable", Message: "notification tester is unavailable"}, nil
	}
	payload, err := testNotificationPayload(channel, time.Now().UTC())
	if err != nil {
		return ChannelTestResult{}, err
	}
	result := s.channelTester.TestChannel(ctx, channel, payload)
	if result.Delivered && result.Message == "" {
		result.Message = "Test notification delivered."
	}
	return result, nil
}

func (s *Service) ListIncidents(ctx context.Context, input ListIncidentsInput) ([]domainalert.Incident, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListIncidents(ctx, project.ID, input.Status, input.Limit)
}

func (s *Service) GetIncident(ctx context.Context, input GetIncidentInput) (domainalert.Incident, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Incident{}, err
	}
	return s.repo.GetIncident(ctx, project.ID, input.IncidentID)
}

func (s *Service) loadProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	return s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
}

func (s *Service) requireProjectAction(ctx context.Context, projectID, userID string, action domainproject.Action) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !domainproject.Can(role, action) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireChannelWrite(ctx context.Context, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if role != domainproject.RoleOwner && role != domainproject.RoleAdmin {
		return ErrForbidden
	}
	return nil
}

func normalizeCreateRule(projectID string, input CreateRuleInput) (domainalert.Rule, error) {
	status := domainalert.RuleStatusDisabled
	if input.Enabled {
		status = domainalert.RuleStatusEnabled
	}
	return normalizeRule(domainalert.Rule{ProjectID: projectID, CreatedByUserID: input.CurrentUserID, Status: status}, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.CooldownSeconds, input.NotificationChannelIDs)
}

func normalizeUpdateRule(projectID string, input UpdateRuleInput) (domainalert.Rule, error) {
	status := domainalert.RuleStatusDisabled
	if input.Enabled {
		status = domainalert.RuleStatusEnabled
	}
	base := domainalert.Rule{ProjectID: projectID, ID: input.RuleID, CreatedByUserID: input.CurrentUserID, Status: status}
	return normalizeRule(base, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.CooldownSeconds, input.NotificationChannelIDs)
}

func normalizeRule(base domainalert.Rule, name string, description *string, severity domainalert.Severity, checkType domaincheck.Type, probeID, checkID *string, condition alertcondition.Condition, cooldownSeconds int32, channelIDs []string) (domainalert.Rule, error) {
	var err error
	base.Name, err = domainalert.VNRuleName(name)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Description, err = domainalert.VNDescription(description)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Severity, err = domainalert.VNSeverity(severity)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.CheckType, err = domaincheck.VNCheckType(checkType)
	if err != nil || base.CheckType == domaincheck.TypeTraceroute {
		return domainalert.Rule{}, fmt.Errorf("%w: unsupported check type", ErrInvalidInput)
	}
	base.Condition, err = alertcondition.Validate(condition)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if !alertcondition.CompatibleWithCheckType(base.Condition.Metric, base.CheckType) {
		return domainalert.Rule{}, fmt.Errorf("%w: condition metric is not compatible with check type", ErrInvalidInput)
	}
	base.ConditionJSON, err = alertcondition.CanonicalJSON(base.Condition)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.ConditionVersion = "metric_threshold.v1"
	base.CooldownSeconds, err = domainalert.VNCooldownSeconds(cooldownSeconds)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.ProbeID = probeID
	base.CheckID = checkID
	if err := validateUUIDPtr(probeID); err != nil {
		return domainalert.Rule{}, err
	}
	if err := validateUUIDPtr(checkID); err != nil {
		return domainalert.Rule{}, err
	}
	base.ProbeSelector = json.RawMessage(`{}`)
	for _, channelID := range channelIDs {
		if _, parseErr := uuid.Parse(channelID); parseErr != nil {
			return domainalert.Rule{}, fmt.Errorf("%w: invalid notification channel id", ErrInvalidInput)
		}
	}
	base.NotificationChannelIDs = append([]string{}, channelIDs...)
	return base, nil
}

func normalizeCreateChannel(projectID string, input CreateChannelInput) (domainalert.NotificationChannel, error) {
	return normalizeChannel(domainalert.NotificationChannel{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, "", input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeUpdateChannel(projectID string, input UpdateChannelInput) (domainalert.NotificationChannel, error) {
	return normalizeChannel(domainalert.NotificationChannel{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, input.ChannelID, input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeChannel(base domainalert.NotificationChannel, channelID, name string, channelType domainalert.ChannelType, enabled bool, config json.RawMessage) (domainalert.NotificationChannel, error) {
	var err error
	base.ID = channelID
	if channelID != "" {
		if _, parseErr := uuid.Parse(channelID); parseErr != nil {
			return domainalert.NotificationChannel{}, fmt.Errorf("%w: invalid channel id", ErrInvalidInput)
		}
	}
	base.Name, err = domainalert.VNChannelName(name)
	if err != nil {
		return domainalert.NotificationChannel{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Type, err = domainalert.VNChannelType(channelType)
	if err != nil {
		return domainalert.NotificationChannel{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Enabled = enabled
	switch base.Type {
	case domainalert.ChannelTypeWebhook:
		canonical, _, err := domainalert.VNWebhookConfig(config)
		if err != nil {
			return domainalert.NotificationChannel{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.ChannelTypeDiscord:
		canonical, _, err := domainalert.VNDiscordConfig(config)
		if err != nil {
			return domainalert.NotificationChannel{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.ChannelTypeTelegram:
		canonical, _, err := domainalert.VNTelegramConfig(config)
		if err != nil {
			return domainalert.NotificationChannel{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	}
	return base, nil
}

func testNotificationPayload(channel domainalert.NotificationChannel, at time.Time) (json.RawMessage, error) {
	data, err := json.Marshal(map[string]any{
		"eventType": "notification.test",
		"sentAt":    at.UTC(),
		"channel": map[string]any{
			"id":   channel.ID,
			"name": channel.Name,
			"type": channel.Type,
		},
		"rule": map[string]any{
			"name":     "Netstamp test alert",
			"severity": domainalert.SeverityInfo,
		},
		"summary": map[string]any{
			"state":   "test",
			"message": "This is a test notification from Netstamp.",
		},
	})
	return data, err
}

func validateUUIDPtr(value *string) error {
	if value == nil {
		return nil
	}
	if _, err := uuid.Parse(*value); err != nil {
		return fmt.Errorf("%w: invalid uuid", ErrInvalidInput)
	}
	return nil
}
