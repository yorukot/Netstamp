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
	Name                   string
	Description            *string
	Enabled                bool
	Severity               domainalert.Severity
	CheckType              domaincheck.Type
	ProbeID                *string
	CheckID                *string
	Condition              alertcondition.Condition
	CooldownSeconds        int32
	NotificationChannelIDs []string
}

type UpdateRuleInput struct {
	ProjectInput
	RuleID                 string
	Name                   string
	Description            *string
	Enabled                bool
	Severity               domainalert.Severity
	CheckType              domaincheck.Type
	ProbeID                *string
	CheckID                *string
	Condition              alertcondition.Condition
	CooldownSeconds        int32
	NotificationChannelIDs []string
}

type DeleteRuleInput struct {
	ProjectInput
	RuleID string
}

type ListChannelsInput struct {
	ProjectInput
	Type *domainalert.ChannelType
}

type GetChannelInput struct {
	ProjectInput
	ChannelID string
}

type CreateChannelInput struct {
	ProjectInput
	Name    string
	Type    domainalert.ChannelType
	Enabled bool
	Config  json.RawMessage
}

type UpdateChannelInput struct {
	ProjectInput
	ChannelID string
	Name      string
	Type      domainalert.ChannelType
	Enabled   bool
	Config    json.RawMessage
}

type DeleteChannelInput struct {
	ProjectInput
	ChannelID string
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
