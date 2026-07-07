package notify

import (
	"context"
	"encoding/json"
	"time"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type NotificationDispatcher struct {
	webhook  *WebhookSender
	slack    *SlackSender
	discord  *DiscordSender
	telegram *TelegramSender
	email    *AlertEmailSender
}

func NewWebhookSender(timeout time.Duration) *NotificationDispatcher {
	return NewSender(timeout, SMTPConfig{})
}

func NewSender(timeout time.Duration, smtpConfig SMTPConfig) *NotificationDispatcher {
	poster := NewJSONPoster(timeout)
	smtp := NewSMTPSender(smtpConfig)

	return &NotificationDispatcher{
		webhook:  NewWebhookNotificationSender(poster),
		slack:    NewSlackSender(poster),
		discord:  NewDiscordSender(poster),
		telegram: NewTelegramSender(poster),
		email:    NewAlertEmailSender(smtp),
	}
}

func NewDynamicSender(timeout time.Duration, smtpProvider SMTPConfigProvider) *NotificationDispatcher {
	poster := NewJSONPoster(timeout)
	smtp := NewDynamicSMTPSender(smtpProvider)

	return &NotificationDispatcher{
		webhook:  NewWebhookNotificationSender(poster),
		slack:    NewSlackSender(poster),
		discord:  NewDiscordSender(poster),
		telegram: NewTelegramSender(poster),
		email:    NewDynamicAlertEmailSender(smtp),
	}
}

func (d *NotificationDispatcher) SendNotification(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	switch notification.Type {
	case domainalert.NotificationTypeWebhook:
		return d.webhook.Send(ctx, notification, payload)
	case domainalert.NotificationTypeSlack:
		return d.slack.Send(ctx, notification, payload)
	case domainalert.NotificationTypeDiscord:
		return d.discord.Send(ctx, notification, payload)
	case domainalert.NotificationTypeTelegram:
		return d.telegram.Send(ctx, notification, payload)
	case domainalert.NotificationTypeEmail:
		return d.email.Send(ctx, notification, payload)
	default:
		return permanent("notification", "unsupported_type", "unsupported notification type")
	}
}

func (d *NotificationDispatcher) EmailConfigured() bool {
	return d.email != nil && d.email.Configured()
}

func (d *NotificationDispatcher) TestNotification(ctx context.Context, notification domainalert.Notification, payload json.RawMessage) appalert.NotificationTestResult {
	result := d.SendNotification(ctx, notification, payload)
	return appalert.NotificationTestResult{
		Delivered: result.Delivered,
		Retryable: result.Retryable,
		Kind:      result.Kind,
		Code:      result.Code,
		Message:   result.Message,
	}
}
