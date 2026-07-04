package alert

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func normalizeCreateNotification(projectID string, input CreateNotificationInput) (domainalert.Notification, error) {
	return normalizeNotification(domainalert.Notification{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, "", input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeUpdateNotification(projectID string, input UpdateNotificationInput) (domainalert.Notification, error) {
	return normalizeNotification(domainalert.Notification{ProjectID: projectID, CreatedByUserID: input.CurrentUserID}, input.NotificationID, input.Name, input.Type, input.Enabled, input.Config)
}

func normalizeNotification(base domainalert.Notification, notificationID, name string, notificationType domainalert.NotificationType, enabled bool, config json.RawMessage) (domainalert.Notification, error) {
	var err error
	base.ID = notificationID
	if notificationID != "" {
		if _, parseErr := uuid.Parse(notificationID); parseErr != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: invalid notification id", ErrInvalidInput)
		}
	}
	base.Name, err = domainalert.VNNotificationName(name)
	if err != nil {
		return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Type, err = domainalert.VNNotificationType(notificationType)
	if err != nil {
		return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Enabled = enabled
	switch base.Type {
	case domainalert.NotificationTypeWebhook:
		canonical, _, err := domainalert.VNWebhookConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeSlack:
		canonical, _, err := domainalert.VNSlackConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeDiscord:
		canonical, _, err := domainalert.VNDiscordConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeTelegram:
		canonical, _, err := domainalert.VNTelegramConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	case domainalert.NotificationTypeEmail:
		canonical, _, err := domainalert.VNEmailConfig(config)
		if err != nil {
			return domainalert.Notification{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		base.Config = canonical
	}
	return base, nil
}
