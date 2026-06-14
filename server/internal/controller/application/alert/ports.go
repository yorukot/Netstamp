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
