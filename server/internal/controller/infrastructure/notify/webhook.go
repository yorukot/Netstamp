package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

const (
	discordMessageLimit  = 2000
	discordTitleLimit    = 256
	discordFieldLimit    = 1024
	telegramMessageLimit = 4096
)

type notificationPayloadView struct {
	EventType string `json:"eventType"`
	SentAt    string `json:"sentAt"`
	Rule      struct {
		Name     string `json:"name"`
		Severity string `json:"severity"`
	} `json:"rule"`
	Target struct {
		ProbeID   string `json:"probeId"`
		CheckID   string `json:"checkId"`
		CheckType string `json:"checkType"`
		Probe     struct {
			Name string `json:"name"`
		} `json:"probe"`
		Check struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Target string `json:"target"`
		} `json:"check"`
	} `json:"target"`
	Notification struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"notification"`
	Summary map[string]any `json:"summary"`
}

type notificationField struct {
	Name   string
	Value  string
	Inline bool
}

type WebhookSender struct {
	client *http.Client
}

func NewWebhookSender(timeout time.Duration) *WebhookSender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	sender := &WebhookSender{}
	sender.client = &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return validateWebhookTarget(req.Context(), req.URL.String())
		},
	}
	return sender
}

func (s *WebhookSender) SendNotification(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	switch notification.Type {
	case domainalert.NotificationTypeWebhook:
		return s.SendWebhook(ctx, notification, payload)
	case domainalert.NotificationTypeDiscord:
		return s.SendDiscord(ctx, notification, payload)
	case domainalert.NotificationTypeTelegram:
		return s.SendTelegram(ctx, notification, payload)
	default:
		return permanent("notification", "unsupported_type", "unsupported notification type")
	}
}

func (s *WebhookSender) TestNotification(ctx context.Context, notification domainalert.Notification, payload json.RawMessage) appalert.NotificationTestResult {
	result := s.SendNotification(ctx, notification, payload)
	return appalert.NotificationTestResult{
		Delivered: result.Delivered,
		Retryable: result.Retryable,
		Kind:      result.Kind,
		Code:      result.Code,
		Message:   result.Message,
	}
}

func (s *WebhookSender) SendWebhook(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.WebhookConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid webhook configuration")
	}
	if err := validateWebhookTarget(ctx, config.URL); err != nil {
		return permanent("security", "blocked_target", err.Error())
	}
	return s.postJSON(ctx, config.URL, payload, "webhook")
}

func (s *WebhookSender) SendDiscord(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.DiscordConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid Discord configuration")
	}
	if err := validateWebhookTarget(ctx, config.URL); err != nil {
		return permanent("security", "blocked_target", err.Error())
	}
	body, err := renderDiscordWebhookBody(payload)
	if err != nil {
		return permanent("request", "invalid_request", "invalid Discord request")
	}
	return s.postJSON(ctx, config.URL, body, "Discord webhook")
}

func (s *WebhookSender) SendTelegram(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.TelegramConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid Telegram configuration")
	}
	endpoint := "https://api.telegram.org/bot" + config.BotToken + "/sendMessage"
	body, err := json.Marshal(map[string]any{
		"chat_id": config.ChatID,
		"text":    truncateMessage(renderNotificationText(payload), telegramMessageLimit),
	})
	if err != nil {
		return permanent("request", "invalid_request", "invalid Telegram request")
	}
	return s.postJSON(ctx, endpoint, body, "Telegram API")
}

func (s *WebhookSender) postJSON(ctx context.Context, endpoint string, body []byte, targetName string) appnotification.DeliveryResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return permanent("request", "invalid_request", "invalid "+targetName+" request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "netstamp-alerts/1")

	resp, err := s.client.Do(req)
	if err != nil {
		return retryable("network", "request_failed", targetName+" request failed")
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return appnotification.DeliveryResult{Delivered: true}
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode >= 500:
		return retryable("http", fmt.Sprintf("status_%d", resp.StatusCode), targetName+" returned retryable status")
	default:
		return permanent("http", fmt.Sprintf("status_%d", resp.StatusCode), targetName+" returned permanent status")
	}
}

func renderDiscordWebhookBody(payload []byte) ([]byte, error) {
	body, ok := parseNotificationPayload(payload)
	if !ok {
		return json.Marshal(map[string]any{
			"username":         "Netstamp",
			"content":          truncateMessage("Netstamp alert\n\n"+string(payload), discordMessageLimit),
			"allowed_mentions": map[string]any{"parse": []string{}},
		})
	}

	title := notificationTitle(body)
	embed := map[string]any{
		"title":  truncateMessage(title, discordTitleLimit),
		"color":  discordColor(body),
		"fields": discordFields(body),
	}
	if description := notificationDescription(body); description != "" {
		embed["description"] = description
	}
	if body.SentAt != "" {
		embed["timestamp"] = body.SentAt
	}

	return json.Marshal(map[string]any{
		"username":         "Netstamp",
		"embeds":           []map[string]any{embed},
		"allowed_mentions": map[string]any{"parse": []string{}},
	})
}

func renderNotificationText(payload []byte) string {
	body, ok := parseNotificationPayload(payload)
	if !ok {
		return "Netstamp alert\n\n" + string(payload)
	}

	lines := []string{notificationTitle(body)}
	if description := notificationDescription(body); description != "" {
		lines = append(lines, "Message: "+description)
	}
	for _, field := range notificationFields(body) {
		lines = append(lines, field.Name+": "+field.Value)
	}
	return strings.Join(lines, "\n")
}

func parseNotificationPayload(payload []byte) (notificationPayloadView, bool) {
	var body notificationPayloadView
	if err := json.Unmarshal(payload, &body); err != nil {
		return notificationPayloadView{}, false
	}
	if body.Summary == nil {
		body.Summary = map[string]any{}
	}
	return body, true
}

func notificationTitle(body notificationPayloadView) string {
	if body.EventType == "notification.test" {
		return "Netstamp test notification"
	}
	return "Netstamp alert"
}

func notificationDescription(body notificationPayloadView) string {
	if message, ok := body.Summary["message"].(string); ok && message != "" {
		return truncateMessage(message, discordFieldLimit)
	}
	if body.Target.CheckType != "" {
		return strings.ToUpper(body.Target.CheckType) + " alert"
	}
	return ""
}

func discordFields(body notificationPayloadView) []map[string]any {
	fields := make([]map[string]any, 0, 8)
	for _, field := range notificationFields(body) {
		fields = append(fields, map[string]any{
			"name":   field.Name,
			"value":  truncateMessage(field.Value, discordFieldLimit),
			"inline": field.Inline,
		})
	}
	return fields
}

func notificationFields(body notificationPayloadView) []notificationField {
	fields := []notificationField{}
	add := func(name, value string, inline bool) {
		if value == "" {
			return
		}
		fields = append(fields, notificationField{Name: name, Value: value, Inline: inline})
	}

	add("Rule", body.Rule.Name, true)
	add("Severity", body.Rule.Severity, true)
	add("Event", body.EventType, true)
	add("Probe", targetProbeLabel(body), true)
	add("Check", targetCheckLabel(body), true)
	add("Target", body.Target.Check.Target, true)
	if metric, ok := body.Summary["metric"].(string); ok {
		add("Metric", metric, true)
	}
	if value, ok := body.Summary["value"]; ok && value != nil {
		add("Value", fmt.Sprintf("%v", value), true)
	}
	if threshold, ok := body.Summary["threshold"]; ok && threshold != nil {
		add("Threshold", fmt.Sprintf("%v", threshold), true)
	}
	add("Notification", body.Notification.Name, true)
	add("Sent", body.SentAt, false)
	return fields
}

func targetProbeLabel(body notificationPayloadView) string {
	if body.Target.Probe.Name != "" {
		return body.Target.Probe.Name
	}
	return body.Target.ProbeID
}

func targetCheckLabel(body notificationPayloadView) string {
	name := body.Target.Check.Name
	if name == "" {
		name = body.Target.CheckID
	}
	checkType := body.Target.Check.Type
	if checkType == "" {
		checkType = body.Target.CheckType
	}
	if checkType == "" {
		return name
	}
	return name + " (" + strings.ToUpper(checkType) + ")"
}

func discordColor(body notificationPayloadView) int {
	if body.EventType == "notification.test" {
		return 0x5865F2
	}
	switch body.Rule.Severity {
	case string(domainalert.SeverityCritical):
		return 0xE5484D
	case string(domainalert.SeverityWarning):
		return 0xF5A524
	default:
		return 0x30A46C
	}
}

func truncateMessage(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 1 {
		return value[:limit]
	}
	return value[:limit-1] + "..."
}

func validateWebhookTarget(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid webhook URL")
	}
	if parsed.Scheme != "https" {
		return errors.New("webhook URL must use https")
	}
	host := parsed.Hostname()
	if host == "" {
		return errors.New("webhook URL host is required")
	}
	if ip, parseErr := netip.ParseAddr(host); parseErr == nil {
		if blockedAddr(ip) {
			return errors.New("webhook URL points to a blocked address")
		}
		return nil
	}
	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return errors.New("webhook host could not be resolved")
	}
	if len(addrs) == 0 {
		return errors.New("webhook host resolved no addresses")
	}
	for _, addr := range addrs {
		if blockedAddr(addr) {
			return errors.New("webhook host resolves to a blocked address")
		}
	}
	return nil
}

func blockedAddr(addr netip.Addr) bool {
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return true
	}
	if addr.Is4() {
		if addr == netip.MustParseAddr("169.254.169.254") {
			return true
		}
	}
	return false
}

func retryable(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: true, Kind: kind, Code: code, Message: message}
}

func permanent(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: false, Kind: kind, Code: code, Message: message}
}
