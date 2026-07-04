package alert

import (
	"context"
	"encoding/json"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *Service) ListNotifications(ctx context.Context, input ListNotificationsInput) ([]domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListNotifications(ctx, project.ID, input.Type)
}

func (s *Service) GetNotification(ctx context.Context, input GetNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.GetNotification(ctx, project.ID, input.NotificationID)
}

func (s *Service) CreateNotification(ctx context.Context, input CreateNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeCreateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.CreateNotification(ctx, notification)
}

func (s *Service) UpdateNotification(ctx context.Context, input UpdateNotificationInput) (domainalert.Notification, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeUpdateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, err
	}
	return s.repo.UpdateNotification(ctx, notification)
}

func (s *Service) DeleteNotification(ctx context.Context, input DeleteNotificationInput) error {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return actionErr
	}
	return s.repo.DeleteNotification(ctx, project.ID, input.NotificationID)
}

func (s *Service) TestNotification(ctx context.Context, input TestNotificationInput) (NotificationTestResult, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return NotificationTestResult{}, err
	}
	if actionErr := s.requireNotificationWrite(ctx, project.ID, input.CurrentUserID); actionErr != nil {
		return NotificationTestResult{}, actionErr
	}
	notification, err := s.repo.GetNotification(ctx, project.ID, input.NotificationID)
	if err != nil {
		return NotificationTestResult{}, err
	}
	if s.notificationTester == nil {
		return NotificationTestResult{Kind: "notification", Code: "tester_unavailable", Message: "notification tester is unavailable"}, nil
	}
	payload, err := testNotificationPayload(notification, time.Now().UTC())
	if err != nil {
		return NotificationTestResult{}, err
	}
	result := s.notificationTester.TestNotification(ctx, notification, payload)
	if result.Delivered && result.Message == "" {
		result.Message = "Test notification delivered."
	}
	return result, nil
}

func testNotificationPayload(notification domainalert.Notification, at time.Time) (json.RawMessage, error) {
	data, err := json.Marshal(map[string]any{
		"eventType": "notification.test",
		"sentAt":    at.UTC(),
		"notification": map[string]any{
			"id":   notification.ID,
			"name": notification.Name,
			"type": notification.Type,
		},
		"rule": map[string]any{
			"name":     "Netstamp test alert",
			"severity": domainalert.SeverityInfo,
		},
		"summary": map[string]any{
			"state":   "test",
			"message": "This is a test notification from Netstamp.",
		},
	})
	return data, err
}
