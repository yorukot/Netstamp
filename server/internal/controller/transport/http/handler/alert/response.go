package alert

import (
	"encoding/json"
	"net/url"
	"time"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

type ruleResponseBody struct {
	ID                     string                   `json:"id"`
	Name                   string                   `json:"name"`
	Description            *string                  `json:"description,omitempty"`
	Enabled                bool                     `json:"enabled"`
	Severity               domainalert.Severity     `json:"severity"`
	Scope                  ruleScopeResponseBody    `json:"scope"`
	Condition              alertcondition.Condition `json:"condition"`
	CooldownSeconds        int32                    `json:"cooldownSeconds"`
	NotificationChannelIDs []string                 `json:"notificationChannelIds"`
	CreatedAt              time.Time                `json:"createdAt"`
	UpdatedAt              time.Time                `json:"updatedAt"`
}

type ruleScopeResponseBody struct {
	CheckType domaincheck.Type `json:"checkType"`
	ProbeID   *string          `json:"probeId,omitempty"`
	CheckID   *string          `json:"checkId,omitempty"`
}

type incidentResponseBody struct {
	ID                          string                         `json:"id"`
	RuleID                      string                         `json:"ruleId"`
	ProbeID                     string                         `json:"probeId"`
	CheckID                     string                         `json:"checkId"`
	Probe                       incidentProbeResponseBody      `json:"probe"`
	Check                       incidentCheckResponseBody      `json:"check"`
	CheckType                   domaincheck.Type               `json:"checkType"`
	Status                      domainalert.IncidentStatus     `json:"status"`
	Severity                    domainalert.Severity           `json:"severity"`
	LastEvaluationState         alertcondition.EvaluationState `json:"lastEvaluationState"`
	OpenedAt                    time.Time                      `json:"openedAt"`
	ResolvedAt                  *time.Time                     `json:"resolvedAt,omitempty"`
	LastEvaluatedAt             time.Time                      `json:"lastEvaluatedAt"`
	LastTriggeredAt             time.Time                      `json:"lastTriggeredAt"`
	LastValue                   *float64                       `json:"lastValue,omitempty"`
	LastSummary                 json.RawMessage                `json:"lastSummary"`
	LastNotificationSentAt      *time.Time                     `json:"lastNotificationSentAt,omitempty"`
	NextNotificationEligibleAt  *time.Time                     `json:"nextNotificationEligibleAt,omitempty"`
	SuppressedNotificationCount int32                          `json:"suppressedNotificationCount"`
	CreatedAt                   time.Time                      `json:"createdAt"`
	UpdatedAt                   time.Time                      `json:"updatedAt"`
}

type incidentProbeResponseBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type incidentCheckResponseBody struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Type   domaincheck.Type `json:"type"`
	Target string           `json:"target"`
}

type channelResponseBody struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Type      domainalert.ChannelType `json:"type"`
	Enabled   bool                    `json:"enabled"`
	Config    json.RawMessage         `json:"config"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type channelTestResponseBody struct {
	Delivered bool   `json:"delivered"`
	Retryable bool   `json:"retryable"`
	Kind      string `json:"kind,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

type telegramChannelResponseConfig struct {
	ChatID             string `json:"chatId"`
	BotTokenConfigured bool   `json:"botTokenConfigured"`
}

func ruleResponses(rules []domainalert.Rule) []ruleResponseBody {
	values := make([]ruleResponseBody, 0, len(rules))
	for _, rule := range rules {
		values = append(values, ruleResponse(rule))
	}
	return values
}

func ruleResponse(rule domainalert.Rule) ruleResponseBody {
	return ruleResponseBody{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Status == domainalert.RuleStatusEnabled,
		Severity:    rule.Severity,
		Scope: ruleScopeResponseBody{
			CheckType: rule.CheckType,
			ProbeID:   rule.ProbeID,
			CheckID:   rule.CheckID,
		},
		Condition:              rule.Condition,
		CooldownSeconds:        rule.CooldownSeconds,
		NotificationChannelIDs: rule.NotificationChannelIDs,
		CreatedAt:              rule.CreatedAt,
		UpdatedAt:              rule.UpdatedAt,
	}
}

func incidentResponses(incidents []domainalert.Incident) []incidentResponseBody {
	values := make([]incidentResponseBody, 0, len(incidents))
	for _, incident := range incidents {
		values = append(values, incidentResponse(incident))
	}
	return values
}

func incidentResponse(incident domainalert.Incident) incidentResponseBody {
	return incidentResponseBody{
		ID:                          incident.ID,
		RuleID:                      incident.RuleID,
		ProbeID:                     incident.ProbeID,
		CheckID:                     incident.CheckID,
		Probe:                       incidentProbeResponse(incident),
		Check:                       incidentCheckResponse(incident),
		CheckType:                   incident.CheckType,
		Status:                      incident.Status,
		Severity:                    incident.Severity,
		LastEvaluationState:         incident.LastEvaluationState,
		OpenedAt:                    incident.OpenedAt,
		ResolvedAt:                  incident.ResolvedAt,
		LastEvaluatedAt:             incident.LastEvaluatedAt,
		LastTriggeredAt:             incident.LastTriggeredAt,
		LastValue:                   incident.LastValue,
		LastSummary:                 incident.LastSummary,
		LastNotificationSentAt:      incident.LastNotificationSentAt,
		NextNotificationEligibleAt:  incident.NextNotificationEligibleAt,
		SuppressedNotificationCount: incident.SuppressedNotificationCount,
		CreatedAt:                   incident.CreatedAt,
		UpdatedAt:                   incident.UpdatedAt,
	}
}

func incidentProbeResponse(incident domainalert.Incident) incidentProbeResponseBody {
	if incident.Probe == nil {
		return incidentProbeResponseBody{ID: incident.ProbeID, Name: incident.ProbeID}
	}
	return incidentProbeResponseBody{ID: incident.Probe.ID, Name: incident.Probe.Name}
}

func incidentCheckResponse(incident domainalert.Incident) incidentCheckResponseBody {
	if incident.Check == nil {
		return incidentCheckResponseBody{ID: incident.CheckID, Name: incident.CheckID, Type: incident.CheckType, Target: incident.CheckID}
	}
	return incidentCheckResponseBody{ID: incident.Check.ID, Name: incident.Check.Name, Type: incident.Check.Type, Target: incident.Check.Target}
}

func channelResponses(channels []domainalert.NotificationChannel) []channelResponseBody {
	values := make([]channelResponseBody, 0, len(channels))
	for _, channel := range channels {
		values = append(values, channelResponse(channel))
	}
	return values
}

func channelResponse(channel domainalert.NotificationChannel) channelResponseBody {
	return channelResponseBody{
		ID:        channel.ID,
		Name:      channel.Name,
		Type:      channel.Type,
		Enabled:   channel.Enabled,
		Config:    channelResponseConfig(channel),
		CreatedAt: channel.CreatedAt,
		UpdatedAt: channel.UpdatedAt,
	}
}

func channelResponseConfig(channel domainalert.NotificationChannel) json.RawMessage {
	switch channel.Type {
	case domainalert.ChannelTypeWebhook:
		var config domainalert.WebhookConfig
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return json.RawMessage(`{}`)
		}
		value, ok := redactedChannelURL(config.URL)
		if !ok {
			return json.RawMessage(`{}`)
		}
		return mustJSON(domainalert.WebhookConfig{URL: value})
	case domainalert.ChannelTypeDiscord:
		var config domainalert.DiscordConfig
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return json.RawMessage(`{}`)
		}
		value, ok := redactedChannelURL(config.URL)
		if !ok {
			return json.RawMessage(`{}`)
		}
		return mustJSON(domainalert.DiscordConfig{URL: value})
	case domainalert.ChannelTypeTelegram:
		var config domainalert.TelegramConfig
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return json.RawMessage(`{}`)
		}
		return mustJSON(telegramChannelResponseConfig{ChatID: config.ChatID, BotTokenConfigured: config.BotToken != ""})
	default:
		return json.RawMessage(`{}`)
	}
}

func channelTestResponse(result appalert.ChannelTestResult) channelTestResponseBody {
	return channelTestResponseBody{
		Delivered: result.Delivered,
		Retryable: result.Retryable,
		Kind:      result.Kind,
		Code:      result.Code,
		Message:   result.Message,
	}
}

func redactedChannelURL(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", false
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = "/..."
	return parsed.String(), true
}

func mustJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}
