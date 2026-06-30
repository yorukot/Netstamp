package notify

import (
	"context"
	"encoding/json"
	"strings"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

const (
	slackTextLimit  = 3000
	slackFieldLimit = 2000
)

type SlackSender struct {
	poster *JSONPoster
}

func NewSlackSender(poster *JSONPoster) *SlackSender {
	return &SlackSender{poster: poster}
}

func (s *SlackSender) Send(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.SlackConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid Slack configuration")
	}
	if err := validateWebhookTarget(ctx, config.URL); err != nil {
		return permanent("security", "blocked_target", err.Error())
	}
	body, err := renderSlackWebhookBody(payload)
	if err != nil {
		return permanent("request", "invalid_request", "invalid Slack request")
	}
	return s.poster.PostJSON(ctx, config.URL, body, "Slack webhook")
}

func renderSlackWebhookBody(payload []byte) ([]byte, error) {
	incident, ok := parseIncidentNotificationPayload(payload)
	if !ok {
		return json.Marshal(map[string]any{
			"text": truncateMessage("Netstamp alert\n\n"+string(payload), slackTextLimit),
		})
	}

	title := incidentNotificationTitle(incident)
	description := incidentNotificationDescription(incident)
	textLines := []string{title}
	if description != "" {
		textLines = append(textLines, description)
	}

	blocks := []map[string]any{
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "*" + slackEscape(title) + "*",
			},
		},
	}
	if description != "" {
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": slackEscape(description),
			},
		})
	}
	if fields := slackFields(incident); len(fields) > 0 {
		blocks = append(blocks, map[string]any{
			"type":   "section",
			"fields": fields,
		})
	}

	return json.Marshal(map[string]any{
		"text":   truncateMessage(strings.Join(textLines, "\n"), slackTextLimit),
		"blocks": blocks,
	})
}

func slackFields(incident incidentNotificationView) []map[string]any {
	fields := make([]map[string]any, 0, 10)
	for _, field := range incidentNotificationFields(incident) {
		value := truncateMessage(slackEscape(field.Value), slackFieldLimit)
		fields = append(fields, map[string]any{
			"type": "mrkdwn",
			"text": "*" + slackEscape(field.Name) + "*\n" + value,
		})
	}
	return fields
}

func slackEscape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	return value
}
