package alert

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

var (
	ErrInvalidInput     = errors.New("alert input invalid")
	ErrRuleNotFound     = errors.New("alert rule not found")
	ErrIncidentNotFound = errors.New("alert incident not found")
	ErrChannelNotFound  = errors.New("notification channel not found")
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

type ChannelType string

const (
	ChannelTypeWebhook ChannelType = "webhook"
	ChannelTypeEmail   ChannelType = "email"
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
	ID                     string
	ProjectID              string
	Name                   string
	Description            *string
	Status                 RuleStatus
	Severity               Severity
	CheckType              domaincheck.Type
	ProbeID                *string
	CheckID                *string
	ProbeSelector          json.RawMessage
	Condition              alertcondition.Condition
	ConditionJSON          json.RawMessage
	ConditionVersion       string
	CooldownSeconds        int32
	NotificationChannelIDs []string
	CreatedByUserID        string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Incident struct {
	ID                          string
	ProjectID                   string
	RuleID                      string
	ProbeID                     string
	CheckID                     string
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

type NotificationChannel struct {
	ID              string
	ProjectID       string
	Name            string
	Type            ChannelType
	Enabled         bool
	Config          json.RawMessage
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NotificationOutboxJob struct {
	ID            string
	ProjectID     string
	IncidentID    string
	RuleID        string
	ChannelID     string
	ChannelType   ChannelType
	EventType     string
	Status        OutboxStatus
	Payload       json.RawMessage
	AttemptCount  int32
	MaxAttempts   int32
	NextAttemptAt time.Time
	LastAttemptAt *time.Time
	DeliveredAt   *time.Time
	LastErrorKind *string
	LastErrorCode *string
	LastError     *string
	DedupeKey     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NotificationJobInput struct {
	ProjectID   string
	IncidentID  string
	RuleID      string
	ChannelID   string
	ChannelType ChannelType
	EventType   string
	Payload     json.RawMessage
	DedupeKey   string
}

type IncidentTransitionInput struct {
	Rule                       Rule
	ProbeID                    string
	CheckID                    string
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
		return nil, nil //nolint:nilnil
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

func VNChannelName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := spvalidator.Required(name); err != nil {
		return "", err
	}
	if err := spvalidator.Max(name, 128); err != nil {
		return "", err
	}
	return name, nil
}

func VNChannelType(channelType ChannelType) (ChannelType, error) {
	switch ChannelType(strings.TrimSpace(string(channelType))) {
	case ChannelTypeWebhook:
		return ChannelTypeWebhook, nil
	case ChannelTypeEmail:
		return "", errors.New("email notification channels are not supported in beta")
	default:
		return "", errors.New("invalid notification channel type")
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

func validateWebhookURL(value string) error {
	if err := spvalidator.Required(value); err != nil {
		return err
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("must be a valid webhook URL")
	}
	if parsed.Scheme != "https" {
		return errors.New("webhook URL must use https")
	}
	if parsed.User != nil {
		return errors.New("webhook URL must not include credentials")
	}
	return nil
}
