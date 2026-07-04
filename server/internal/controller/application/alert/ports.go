package alert

import (
	"context"
	"encoding/json"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	ListRules(ctx context.Context, projectID string, status *domainalert.RuleStatus, checkType *domaincheck.Type) ([]domainalert.Rule, error)
	GetRule(ctx context.Context, projectID, ruleID string) (domainalert.Rule, error)
	CreateRule(ctx context.Context, input domainalert.Rule) (domainalert.Rule, error)
	UpdateRule(ctx context.Context, input domainalert.Rule) (domainalert.Rule, error)
	DeleteRule(ctx context.Context, projectID, ruleID string) error
	ListNotifications(ctx context.Context, projectID string, notificationType *domainalert.NotificationType) ([]domainalert.Notification, error)
	GetNotification(ctx context.Context, projectID, notificationID string) (domainalert.Notification, error)
	CreateNotification(ctx context.Context, input domainalert.Notification) (domainalert.Notification, error)
	UpdateNotification(ctx context.Context, input domainalert.Notification) (domainalert.Notification, error)
	DeleteNotification(ctx context.Context, projectID, notificationID string) error
	ListIncidents(ctx context.Context, projectID string, status *domainalert.IncidentStatus, limit int32) ([]domainalert.Incident, error)
	GetIncident(ctx context.Context, projectID, incidentID string) (domainalert.Incident, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type NotificationTester interface {
	TestNotification(ctx context.Context, notification domainalert.Notification, payload json.RawMessage) NotificationTestResult
}

type AlertAction string

const (
	AlertActionListRules          AlertAction = "rule.list"
	AlertActionGetRule            AlertAction = "rule.get"
	AlertActionCreateRule         AlertAction = "rule.create"
	AlertActionUpdateRule         AlertAction = "rule.update"
	AlertActionDeleteRule         AlertAction = "rule.delete"
	AlertActionListNotifications  AlertAction = "notification.list"
	AlertActionGetNotification    AlertAction = "notification.get"
	AlertActionCreateNotification AlertAction = "notification.create"
	AlertActionUpdateNotification AlertAction = "notification.update"
	AlertActionDeleteNotification AlertAction = "notification.delete"
	AlertActionTestNotification   AlertAction = "notification.test"
	AlertActionListIncidents      AlertAction = "incident.list"
	AlertActionGetIncident        AlertAction = "incident.get"
)

type AlertOutcome string

const (
	AlertOutcomeSuccess AlertOutcome = "success"
	AlertOutcomeFailure AlertOutcome = "failure"
)

type AlertReason string

const (
	AlertReasonInvalidInput                  AlertReason = "invalid_input"
	AlertReasonForbidden                     AlertReason = "forbidden"
	AlertReasonProjectNotFound               AlertReason = "project_not_found"
	AlertReasonUserNotFound                  AlertReason = "user_not_found"
	AlertReasonRuleNotFound                  AlertReason = "rule_not_found"
	AlertReasonNotificationNotFound          AlertReason = "notification_not_found"
	AlertReasonIncidentNotFound              AlertReason = "incident_not_found"
	AlertReasonNotificationTesterUnavailable AlertReason = "notification_tester_unavailable"
	AlertReasonProjectLookupFailed           AlertReason = "project_lookup_failed"
	AlertReasonRoleLookupFailed              AlertReason = "role_lookup_failed"
	AlertReasonRuleListFailed                AlertReason = "rule_list_failed"
	AlertReasonRuleLookupFailed              AlertReason = "rule_lookup_failed"
	AlertReasonRuleCreateFailed              AlertReason = "rule_create_failed"
	AlertReasonRuleUpdateFailed              AlertReason = "rule_update_failed"
	AlertReasonRuleDeleteFailed              AlertReason = "rule_delete_failed"
	AlertReasonNotificationListFailed        AlertReason = "notification_list_failed"
	AlertReasonNotificationLookupFailed      AlertReason = "notification_lookup_failed"
	AlertReasonNotificationCreateFailed      AlertReason = "notification_create_failed"
	AlertReasonNotificationUpdateFailed      AlertReason = "notification_update_failed"
	AlertReasonNotificationDeleteFailed      AlertReason = "notification_delete_failed"
	AlertReasonNotificationTestFailed        AlertReason = "notification_test_failed"
	AlertReasonIncidentListFailed            AlertReason = "incident_list_failed"
	AlertReasonIncidentLookupFailed          AlertReason = "incident_lookup_failed"
)
