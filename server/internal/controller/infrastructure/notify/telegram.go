package notify

import (
	"context"
	"encoding/json"
	"strings"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

const telegramMessageLimit = 4096

func (s *WebhookSender) SendTelegram(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.TelegramConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid Telegram configuration")
	}
	endpoint := "https://api.telegram.org/bot" + config.BotToken + "/sendMessage"
	body, err := json.Marshal(map[string]any{
		"chat_id": config.ChatID,
		"text":    truncateMessage(renderTelegramMessage(payload), telegramMessageLimit),
	})
	if err != nil {
		return permanent("request", "invalid_request", "invalid Telegram request")
	}
	return s.postJSON(ctx, endpoint, body, "Telegram API")
}

func renderTelegramMessage(payload []byte) string {
	incident, ok := parseIncidentNotificationPayload(payload)
	if !ok {
		return "Netstamp alert\n\n" + string(payload)
	}

	lines := []string{incidentNotificationTitle(incident)}
	if description := incidentNotificationDescription(incident); description != "" {
		lines = append(lines, "Message: "+description)
	}
	for _, field := range incidentNotificationFields(incident) {
		lines = append(lines, field.Name+": "+field.Value)
	}
	return strings.Join(lines, "\n")
}
