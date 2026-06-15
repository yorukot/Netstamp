package notify

import (
	"context"
	"encoding/json"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

const (
	discordMessageLimit = 2000
	discordTitleLimit   = 256
	discordFieldLimit   = 1024
)

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

func renderDiscordWebhookBody(payload []byte) ([]byte, error) {
	incident, ok := parseIncidentNotificationPayload(payload)
	if !ok {
		return json.Marshal(map[string]any{
			"username":         "Netstamp",
			"content":          truncateMessage("Netstamp alert\n\n"+string(payload), discordMessageLimit),
			"allowed_mentions": map[string]any{"parse": []string{}},
		})
	}

	embed := map[string]any{
		"title":  truncateMessage(incidentNotificationTitle(incident), discordTitleLimit),
		"color":  discordColor(incident),
		"fields": discordFields(incident),
	}
	if incident.Links.Incident != "" {
		embed["url"] = incident.Links.Incident
	}
	if description := incidentNotificationDescription(incident); description != "" {
		embed["description"] = description
	}
	if incident.SentAt != "" {
		embed["timestamp"] = incident.SentAt
	}

	return json.Marshal(map[string]any{
		"username":         "Netstamp",
		"embeds":           []map[string]any{embed},
		"allowed_mentions": map[string]any{"parse": []string{}},
	})
}

func discordFields(incident incidentNotificationView) []map[string]any {
	fields := make([]map[string]any, 0, 8)
	for _, field := range incidentNotificationFields(incident) {
		fields = append(fields, map[string]any{
			"name":   field.Name,
			"value":  truncateMessage(field.Value, discordFieldLimit),
			"inline": field.Inline,
		})
	}
	return fields
}

func discordColor(incident incidentNotificationView) int {
	if incident.EventType == "notification.test" {
		return 0x5865F2
	}
	switch incident.Rule.Severity {
	case string(domainalert.SeverityCritical):
		return 0xE5484D
	case string(domainalert.SeverityWarning):
		return 0xF5A524
	default:
		return 0x30A46C
	}
}
