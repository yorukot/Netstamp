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
	repo               Repository
	projectAccess      ProjectAccess
	notificationTester NotificationTester
}

func NewService(repo Repository, projectAccess ProjectAccess, notificationTester NotificationTester) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, notificationTester: notificationTester}
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

func (s *Service) ListNotifications(ctx context.Context, input ListNotificationsInput) ([]domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListNotifications(ctx, project.ID, input.Type)
}

func (s *Service) GetNotification(ctx context.Context, input GetNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.GetNotification(ctx, project.ID, input.NotificationID)
}

func (s *Service) CreateNotification(ctx context.Context, input CreateNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeCreateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.CreateNotification(ctx, notification)
}

func (s *Service) UpdateNotification(ctx context.Context, input UpdateNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeUpdateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.UpdateNotification(ctx, notification)
}

func (s *Service) DeleteNotification(ctx context.Context, input DeleteNotificationInput) error {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return actionErr
	}
	return s.repo.DeleteNotification(ctx, project.ID, input.NotificationID)
}

func (s *Service) TestNotification(ctx context.Context, input TestNotificationInput) (NotificationTestResult, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return NotificationTestResult{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return NotificationTestResult{}, actionErr
	}
	notification, err := s.repo.GetNotification(ctx, project.ID, input.NotificationID)
	if err != nil {
		return NotificationTestResult{}, err
	}
	if s.notificationTester == nil {
		return NotificationTestResult{Kind: "notification", Code: "tester_unavailable", Message: "notification tester is unavailable"}, nil
	}
	payload, err := testNotificationPayload(notification, time.Now().UTC())
	if err != nil {
		return NotificationTestResult{}, err
	}
	result := s.notificationTester.TestNotification(ctx, notification, payload)
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

func (s *Service) requireNotificationWrite(ctx context.Context, projectID, userID string) error {
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
	return normalizeRule(domainalert.Rule{ProjectID: projectID, CreatedByUserID: input.CurrentUserID, Status: status}, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.CooldownSeconds, input.NotificationIDs)
}

func normalizeUpdateRule(projectID string, input UpdateRuleInput) (domainalert.Rule, error) {
	status := domainalert.RuleStatusDisabled
	if input.Enabled {
		status = domainalert.RuleStatusEnabled
	}
	base := domainalert.Rule{ProjectID: projectID, ID: input.RuleID, CreatedByUserID: input.CurrentUserID, Status: status}
	return normalizeRule(base, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.CooldownSeconds, input.NotificationIDs)
}

func normalizeRule(base domainalert.Rule, name string, description *string, severity domainalert.Severity, checkType domaincheck.Type, probeID, checkID *string, condition alertcondition.Condition, cooldownSeconds int32, notificationIDs []string) (domainalert.Rule, error) {
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
	for _, notificationID := range notificationIDs {
		if _, parseErr := uuid.Parse(notificationID); parseErr != nil {
			return domainalert.Rule{}, fmt.Errorf("%w: invalid notification id", ErrInvalidInput)
		}
	}
	base.NotificationIDs = append([]string{}, notificationIDs...)
	return base, nil
}

func normalizeCreateNotification(projectID string, input CreateNotificationInput) (domainalert.Notification, error) {
	return normalizeNotification(domainalert.Notification{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, "", input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeUpdateNotification(projectID string, input UpdateNotificationInput) (domainalert.Notification, error) {
	return normalizeNotification(domainalert.Notification{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, input.NotificationID, input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeNotification(base domainalert.Notification, notificationID, name string, notificationType domainalert.NotificationType, enabled bool, config json.RawMessage) (domainalert.Notification, error) {
	var err error
	base.ID = notificationID
	if notificationID != "" {
		if _, parseErr := uuid.Parse(notificationID); parseErr != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: invalid notification id", ErrInvalidInput)
		}
	}
	base.Name, err = domainalert.VNNotificationName(name)
	if err != nil {
		return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Type, err = domainalert.VNNotificationType(notificationType)
	if err != nil {
		return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Enabled = enabled
	switch base.Type {
	case domainalert.NotificationTypeWebhook:
		canonical, _, err := domainalert.VNWebhookConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeSlack:
		canonical, _, err := domainalert.VNSlackConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeDiscord:
		canonical, _, err := domainalert.VNDiscordConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeTelegram:
		canonical, _, err := domainalert.VNTelegramConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeEmail:
		canonical, _, err := domainalert.VNEmailConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	}
	return base, nil
}

func testNotificationPayload(notification domainalert.Notification, at time.Time) (json.RawMessage, error) {
	data, err := json.Marshal(map[string]any{
		"eventType": "notification.test",
		"sentAt":    at.UTC(),
		"notification": map[string]any{
			"id":   notification.ID,
			"name": notification.Name,
			"type": notification.Type,
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
