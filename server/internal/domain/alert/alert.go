package alert

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

var (
	ErrInvalidInput         = errors.New("alert input invalid")
	ErrRuleNotFound         = errors.New("alert rule not found")
	ErrIncidentNotFound     = errors.New("alert incident not found")
	ErrNotificationNotFound = errors.New("notification not found")
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type RuleStatus string

const (
	RuleStatusEnabled  RuleStatus = "enabled"
	RuleStatusDisabled RuleStatus = "disabled"
)

type IncidentStatus string

const (
	IncidentStatusOpen         IncidentStatus = "open"
	IncidentStatusAcknowledged IncidentStatus = "acknowledged"
	IncidentStatusResolved     IncidentStatus = "resolved"
)

type NotificationType string

const (
	NotificationTypeWebhook  NotificationType = "webhook"
	NotificationTypeSlack    NotificationType = "slack"
	NotificationTypeDiscord  NotificationType = "discord"
	NotificationTypeTelegram NotificationType = "telegram"
	NotificationTypeEmail    NotificationType = "email"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusSending   OutboxStatus = "sending"
	OutboxStatusDelivered OutboxStatus = "delivered"
	OutboxStatusFailed    OutboxStatus = "failed"
	OutboxStatusDiscarded OutboxStatus = "discarded"
)

const DefaultCooldownSeconds int32 = 900

type Rule struct {
	ID               string
	ProjectID        string
	Name             string
	Description      *string
	Status           RuleStatus
	Severity         Severity
	CheckType        domaincheck.Type
	ProbeID          *string
	CheckID          *string
	ProbeSelector    json.RawMessage
	Condition        alertcondition.Condition
	ConditionJSON    json.RawMessage
	ConditionVersion string
	CooldownSeconds  int32
	NotificationIDs  []string
	CreatedByUserID  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Incident struct {
	ID                          string
	ProjectID                   string
	RuleID                      string
	ProbeID                     string
	CheckID                     string
	Probe                       *IncidentProbeSummary
	Check                       *IncidentCheckSummary
	CheckType                   domaincheck.Type
	Status                      IncidentStatus
	Severity                    Severity
	LastEvaluationState         alertcondition.EvaluationState
	OpenedAt                    time.Time
	AcknowledgedAt              *time.Time
	AcknowledgedByUserID        *string
	ResolvedAt                  *time.Time
	ResolvedByUserID            *string
	LastEvaluatedAt             time.Time
	LastTriggeredAt             time.Time
	LastValue                   *float64
	LastSummary                 json.RawMessage
	LastNotificationSentAt      *time.Time
	NextNotificationEligibleAt  *time.Time
	SuppressedNotificationCount int32
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

type IncidentProbeSummary struct {
	ID   string
	Name string
}

type IncidentCheckSummary struct {
	ID     string
	Name   string
	Type   domaincheck.Type
	Target string
}

type Notification struct {
	ID              string
	ProjectID       string
	Name            string
	Type            NotificationType
	Enabled         bool
	Config          json.RawMessage
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NotificationOutboxJob struct {
	ID               string
	ProjectID        string
	IncidentID       string
	RuleID           string
	NotificationID   string
	NotificationType NotificationType
	EventType        string
	Status           OutboxStatus
	Payload          json.RawMessage
	AttemptCount     int32
	MaxAttempts      int32
	NextAttemptAt    time.Time
	LastAttemptAt    *time.Time
	DeliveredAt      *time.Time
	LastErrorKind    *string
	LastErrorCode    *string
	LastError        *string
	DedupeKey        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NotificationJobInput struct {
	ProjectID        string
	IncidentID       string
	RuleID           string
	NotificationID   string
	NotificationType NotificationType
	EventType        string
	Payload          json.RawMessage
	DedupeKey        string
}

type IncidentTransitionInput struct {
	Rule                       Rule
	ProbeID                    string
	CheckID                    string
	Probe                      *IncidentProbeSummary
	Check                      *IncidentCheckSummary
	CheckType                  domaincheck.Type
	Evaluation                 alertcondition.Evaluation
	Summary                    json.RawMessage
	At                         time.Time
	LastNotificationSentAt     *time.Time
	NextNotificationEligibleAt *time.Time
	Jobs                       []NotificationJobInput
}

type WebhookConfig struct {
	URL string `json:"url"`
}

type SlackConfig struct {
	URL string `json:"url"`
}

type DiscordConfig struct {
	URL string `json:"url"`
}

type TelegramConfig struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}

type EmailConfig struct {
	To []string `json:"to"`
}

func VNRuleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := spvalidator.Required(name); err != nil {
		return "", err
	}
	if err := spvalidator.Max(name, 128); err != nil {
		return "", err
	}
	return name, nil
}

func VNDescription(description *string) (*string, error) {
	if description == nil {
		return nil, nil //nolint:nilnil // nil is the canonical value for an omitted optional description.
	}
	trimmed := strings.TrimSpace(*description)
	if err := spvalidator.Required(trimmed); err != nil {
		return nil, err
	}
	if err := spvalidator.Max(trimmed, 1024); err != nil {
		return nil, err
	}
	return &trimmed, nil
}

func VNSeverity(severity Severity) (Severity, error) {
	switch Severity(strings.TrimSpace(string(severity))) {
	case SeverityInfo:
		return SeverityInfo, nil
	case SeverityWarning:
		return SeverityWarning, nil
	case SeverityCritical:
		return SeverityCritical, nil
	default:
		return "", errors.New("invalid alert severity")
	}
}

func VNRuleStatus(status RuleStatus) (RuleStatus, error) {
	switch RuleStatus(strings.TrimSpace(string(status))) {
	case "", RuleStatusEnabled:
		return RuleStatusEnabled, nil
	case RuleStatusDisabled:
		return RuleStatusDisabled, nil
	default:
		return "", errors.New("invalid alert rule status")
	}
}

func VNCooldownSeconds(value int32) (int32, error) {
	if value == 0 {
		value = DefaultCooldownSeconds
	}
	if err := spvalidator.Min(value, int32(60)); err != nil {
		return 0, err
	}
	if err := spvalidator.Max(value, int32(86400)); err != nil {
		return 0, err
	}
	return value, nil
}

func VNNotificationName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := spvalidator.Required(name); err != nil {
		return "", err
	}
	if err := spvalidator.Max(name, 128); err != nil {
		return "", err
	}
	return name, nil
}

func VNNotificationType(notificationType NotificationType) (NotificationType, error) {
	switch NotificationType(strings.TrimSpace(string(notificationType))) {
	case NotificationTypeWebhook:
		return NotificationTypeWebhook, nil
	case NotificationTypeSlack:
		return NotificationTypeSlack, nil
	case NotificationTypeDiscord:
		return NotificationTypeDiscord, nil
	case NotificationTypeTelegram:
		return NotificationTypeTelegram, nil
	case NotificationTypeEmail:
		return NotificationTypeEmail, nil
	default:
		return "", errors.New("invalid notification type")
	}
}

func VNWebhookConfig(raw json.RawMessage) (json.RawMessage, WebhookConfig, error) {
	if len(raw) == 0 {
		return nil, WebhookConfig{}, errors.New("webhook config is required")
	}
	var config WebhookConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, WebhookConfig{}, err
	}
	config.URL = strings.TrimSpace(config.URL)
	if err := validateWebhookURL(config.URL); err != nil {
		return nil, WebhookConfig{}, err
	}
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, WebhookConfig{}, err
	}
	return canonical, config, nil
}

func VNSlackConfig(raw json.RawMessage) (json.RawMessage, SlackConfig, error) {
	if len(raw) == 0 {
		return nil, SlackConfig{}, errors.New("slack config is required")
	}
	var config SlackConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, SlackConfig{}, err
	}
	config.URL = strings.TrimSpace(config.URL)
	if err := validateSlackWebhookURL(config.URL); err != nil {
		return nil, SlackConfig{}, err
	}
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, SlackConfig{}, err
	}
	return canonical, config, nil
}

func VNDiscordConfig(raw json.RawMessage) (json.RawMessage, DiscordConfig, error) {
	if len(raw) == 0 {
		return nil, DiscordConfig{}, errors.New("discord config is required")
	}
	var config DiscordConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, DiscordConfig{}, err
	}
	config.URL = strings.TrimSpace(config.URL)
	if err := validateDiscordWebhookURL(config.URL); err != nil {
		return nil, DiscordConfig{}, err
	}
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, DiscordConfig{}, err
	}
	return canonical, config, nil
}

func VNTelegramConfig(raw json.RawMessage) (json.RawMessage, TelegramConfig, error) {
	if len(raw) == 0 {
		return nil, TelegramConfig{}, errors.New("telegram config is required")
	}
	var config TelegramConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, TelegramConfig{}, err
	}
	config.BotToken = strings.TrimSpace(config.BotToken)
	config.ChatID = strings.TrimSpace(config.ChatID)
	if err := validateTelegramBotToken(config.BotToken); err != nil {
		return nil, TelegramConfig{}, err
	}
	if err := validateTelegramChatID(config.ChatID); err != nil {
		return nil, TelegramConfig{}, err
	}
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, TelegramConfig{}, err
	}
	return canonical, config, nil
}

func VNEmailConfig(raw json.RawMessage) (json.RawMessage, EmailConfig, error) {
	if len(raw) == 0 {
		return nil, EmailConfig{}, errors.New("email config is required")
	}
	var config EmailConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, EmailConfig{}, err
	}
	recipients, err := validateEmailRecipients(config.To)
	if err != nil {
		return nil, EmailConfig{}, err
	}
	config.To = recipients
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, EmailConfig{}, err
	}
	return canonical, config, nil
}

func validateWebhookURL(value string) error {
	_, err := validateHTTPSURL(value, "webhook URL")
	return err
}

func validateSlackWebhookURL(value string) error {
	parsed, err := validateHTTPSURL(value, "slack webhook URL")
	if err != nil {
		return err
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "hooks.slack.com" && host != "hooks.slack-gov.com" {
		return errors.New("slack webhook URL must use a Slack webhook host")
	}
	if !strings.HasPrefix(parsed.EscapedPath(), "/services/") {
		return errors.New("slack webhook URL must be an incoming webhook URL")
	}
	return nil
}

func validateDiscordWebhookURL(value string) error {
	parsed, err := validateHTTPSURL(value, "discord webhook URL")
	if err != nil {
		return err
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "discord.com" && host != "discordapp.com" && host != "canary.discord.com" && host != "ptb.discord.com" {
		return errors.New("discord webhook URL must use a Discord host")
	}
	if !strings.HasPrefix(parsed.EscapedPath(), "/api/webhooks/") {
		return errors.New("discord webhook URL must be a webhook execute URL")
	}
	return nil
}

func validateHTTPSURL(value, label string) (*url.URL, error) {
	if err := spvalidator.Required(value); err != nil {
		return nil, err
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("must be a valid %s", label)
	}
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s must use https", label)
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("%s must not include credentials", label)
	}
	return parsed, nil
}

func validateTelegramBotToken(value string) error {
	if err := spvalidator.Required(value); err != nil {
		return err
	}
	if err := spvalidator.Max(value, 256); err != nil {
		return err
	}
	botID, token, ok := strings.Cut(value, ":")
	if !ok || botID == "" || token == "" {
		return errors.New("telegram bot token is invalid")
	}
	if !digitsOnly(botID) || !validTelegramTokenSecret(token) {
		return errors.New("telegram bot token is invalid")
	}
	return nil
}

func digitsOnly(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func validTelegramTokenSecret(value string) bool {
	for _, char := range value {
		if !validTelegramTokenChar(char) {
			return false
		}
	}
	return true
}

func validTelegramTokenChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' ||
		char == '-'
}

func validateTelegramChatID(value string) error {
	if err := spvalidator.Required(value); err != nil {
		return err
	}
	return spvalidator.Max(value, 128)
}

func validateEmailRecipients(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, errors.New("email recipient is required")
	}
	if len(values) > 50 {
		return nil, errors.New("email recipient list must contain 50 or fewer addresses")
	}
	seen := make(map[string]struct{}, len(values))
	recipients := make([]string, 0, len(values))
	for _, value := range values {
		recipient, err := validateEmailRecipient(value)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[recipient]; ok {
			continue
		}
		seen[recipient] = struct{}{}
		recipients = append(recipients, recipient)
	}
	if len(recipients) == 0 {
		return nil, errors.New("email recipient is required")
	}
	return recipients, nil
}

func validateEmailRecipient(value string) (string, error) {
	recipient := strings.ToLower(strings.TrimSpace(value))
	if err := spvalidator.Email(recipient); err != nil {
		return "", err
	}
	if err := spvalidator.Max(recipient, 254); err != nil {
		return "", err
	}
	return recipient, nil
}
