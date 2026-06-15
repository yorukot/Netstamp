package notify

import (
	"context"
	"encoding/json"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *WebhookSender) SendWebhook(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	var config domainalert.WebhookConfig
	if err := json.Unmarshal(notification.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid webhook configuration")
	}
	if err := validateWebhookTarget(ctx, config.URL); err != nil {
		return permanent("security", "blocked_target", err.Error())
	}
	return s.postJSON(ctx, config.URL, renderWebhookBody(payload), "webhook")
}

func renderWebhookBody(payload []byte) []byte {
	return payload
}
