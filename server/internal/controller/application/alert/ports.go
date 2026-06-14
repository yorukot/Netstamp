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
	ListChannels(ctx context.Context, projectID string, channelType *domainalert.ChannelType) ([]domainalert.NotificationChannel, error)
	GetChannel(ctx context.Context, projectID, channelID string) (domainalert.NotificationChannel, error)
	CreateChannel(ctx context.Context, input domainalert.NotificationChannel) (domainalert.NotificationChannel, error)
	UpdateChannel(ctx context.Context, input domainalert.NotificationChannel) (domainalert.NotificationChannel, error)
	DeleteChannel(ctx context.Context, projectID, channelID string) error
	ListIncidents(ctx context.Context, projectID string, status *domainalert.IncidentStatus, limit int32) ([]domainalert.Incident, error)
	GetIncident(ctx context.Context, projectID, incidentID string) (domainalert.Incident, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type ChannelTester interface {
	TestChannel(ctx context.Context, channel domainalert.NotificationChannel, payload json.RawMessage) ChannelTestResult
}
