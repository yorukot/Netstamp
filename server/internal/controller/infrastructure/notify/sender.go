package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type WebhookSender struct {
	client *http.Client
	smtp   *SMTPSender
}

func NewWebhookSender(timeout time.Duration) *WebhookSender {
	return NewSender(timeout, SMTPConfig{})
}

func NewSender(timeout time.Duration, smtpConfig SMTPConfig) *WebhookSender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSender{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, _ []*http.Request) error {
				return validateWebhookTarget(req.Context(), req.URL.String())
			},
		},
		smtp: NewSMTPSender(smtpConfig),
	}
}

func (s *WebhookSender) SendNotification(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	switch notification.Type {
	case domainalert.NotificationTypeWebhook:
		return s.SendWebhook(ctx, notification, payload)
	case domainalert.NotificationTypeSlack:
		return s.SendSlack(ctx, notification, payload)
	case domainalert.NotificationTypeDiscord:
		return s.SendDiscord(ctx, notification, payload)
	case domainalert.NotificationTypeTelegram:
		return s.SendTelegram(ctx, notification, payload)
	case domainalert.NotificationTypeEmail:
		return s.SendEmail(ctx, notification, payload)
	default:
		return permanent("notification", "unsupported_type", "unsupported notification type")
	}
}

func (s *WebhookSender) EmailConfigured() bool {
	return s.smtp != nil && s.smtp.Configured()
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
