package alert

import (
	"encoding/json"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

type ProjectInput struct {
	ProjectRef    string
	CurrentUserID string
}

type ListRulesInput struct {
	ProjectInput
	Status    *domainalert.RuleStatus
	CheckType *domaincheck.Type
}

type GetRuleInput struct {
	ProjectInput
	RuleID string
}

type CreateRuleInput struct {
	ProjectInput
	Name            string
	Description     *string
	Enabled         bool
	Severity        domainalert.Severity
	CheckType       domaincheck.Type
	ProbeID         *string
	CheckID         *string
	Condition       alertcondition.Condition
	CooldownSeconds int32
	NotificationIDs []string
}

type UpdateRuleInput struct {
	ProjectInput
	RuleID          string
	Name            string
	Description     *string
	Enabled         bool
	Severity        domainalert.Severity
	CheckType       domaincheck.Type
	ProbeID         *string
	CheckID         *string
	Condition       alertcondition.Condition
	CooldownSeconds int32
	NotificationIDs []string
}

type DeleteRuleInput struct {
	ProjectInput
	RuleID string
}

type ListNotificationsInput struct {
	ProjectInput
	Type *domainalert.NotificationType
}

type GetNotificationInput struct {
	ProjectInput
	NotificationID string
}

type CreateNotificationInput struct {
	ProjectInput
	Name    string
	Type    domainalert.NotificationType
	Enabled bool
	Config  json.RawMessage
}

type UpdateNotificationInput struct {
	ProjectInput
	NotificationID string
	Name           string
	Type           domainalert.NotificationType
	Enabled        bool
	Config         json.RawMessage
}

type DeleteNotificationInput struct {
	ProjectInput
	NotificationID string
}

type TestNotificationInput struct {
	ProjectInput
	NotificationID string
}

type NotificationTestResult struct {
	Delivered bool
	Retryable bool
	Kind      string
	Code      string
	Message   string
}

type ListIncidentsInput struct {
	ProjectInput
	Status *domainalert.IncidentStatus
	Limit  int32
}

type GetIncidentInput struct {
	ProjectInput
	IncidentID string
}
