package alert

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/trace"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *Service) ListNotifications(ctx context.Context, input ListNotificationsInput) ([]domainalert.Notification, error) {
	ctx, span := alertTracer.Start(ctx, "alert.notification.list", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionListNotifications)),
		attrProjectRef.String(input.ProjectRef),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	notifications, err := s.repo.ListNotifications(ctx, project.ID, input.Type)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonNotificationListFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return notifications, nil
}

func (s *Service) GetNotification(ctx context.Context, input GetNotificationInput) (domainalert.Notification, error) {
	ctx, span := alertTracer.Start(ctx, "alert.notification.get", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionGetNotification)),
		attrProjectRef.String(input.ProjectRef),
		attrAlertNotificationID.String(input.NotificationID),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	notification, err := s.repo.GetNotification(ctx, project.ID, input.NotificationID)
	if err != nil {
		return domainalert.Notification{}, recordAlertQueryFailure(span, AlertReasonNotificationLookupFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return notification, nil
}

func (s *Service) CreateNotification(ctx context.Context, input CreateNotificationInput) (domainalert.Notification, error) {
	ctx, flow := s.startAlertFlow(ctx, "alert.notification.create", AlertActionCreateNotification, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeCreateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, flow.writeFailure(AlertReasonNotificationCreateFailed, err)
	}
	created, err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		return domainalert.Notification{}, flow.writeFailure(AlertReasonNotificationCreateFailed, err)
	}
	flow.setNotificationID(created.ID)
	flow.success()
	return created, nil
}

func (s *Service) UpdateNotification(ctx context.Context, input UpdateNotificationInput) (domainalert.Notification, error) {
	ctx, flow := s.startAlertFlow(ctx, "alert.notification.update", AlertActionUpdateNotification, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setNotificationID(input.NotificationID)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Notification{}, err
	}
	if actionErr := s.requireNotificationWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); actionErr != nil {
		return domainalert.Notification{}, actionErr
	}
	notification, err := normalizeUpdateNotification(project.ID, input)
	if err != nil {
		return domainalert.Notification{}, flow.writeFailure(AlertReasonNotificationUpdateFailed, err)
	}
	updated, err := s.repo.UpdateNotification(ctx, notification)
	if err != nil {
		return domainalert.Notification{}, flow.writeFailure(AlertReasonNotificationUpdateFailed, err)
	}
	flow.setNotificationID(updated.ID)
	flow.success()
	return updated, nil
}

func (s *Service) DeleteNotification(ctx context.Context, input DeleteNotificationInput) error {
	ctx, flow := s.startAlertFlow(ctx, "alert.notification.delete", AlertActionDeleteNotification, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setNotificationID(input.NotificationID)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if actionErr := s.requireNotificationWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); actionErr != nil {
		return actionErr
	}
	if err := s.repo.DeleteNotification(ctx, project.ID, input.NotificationID); err != nil {
		return flow.writeFailure(AlertReasonNotificationDeleteFailed, err)
	}
	flow.success()
	return nil
}

func (s *Service) TestNotification(ctx context.Context, input TestNotificationInput) (NotificationTestResult, error) {
	ctx, flow := s.startAlertFlow(ctx, "alert.notification.test", AlertActionTestNotification, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setNotificationID(input.NotificationID)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return NotificationTestResult{}, err
	}
	if actionErr := s.requireNotificationWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); actionErr != nil {
		return NotificationTestResult{}, actionErr
	}
	notification, err := s.repo.GetNotification(ctx, project.ID, input.NotificationID)
	if err != nil {
		return NotificationTestResult{}, flow.writeFailure(AlertReasonNotificationLookupFailed, err)
	}
	flow.setNotificationID(notification.ID)
	if s.notificationTester == nil {
		flow.businessResult(AlertReasonNotificationTesterUnavailable)
		return NotificationTestResult{Kind: "notification", Code: "tester_unavailable", Message: "notification tester is unavailable"}, nil
	}
	payload, err := testNotificationPayload(notification, time.Now().UTC())
	if err != nil {
		return NotificationTestResult{}, flow.writeFailure(AlertReasonNotificationTestFailed, err)
	}
	result := s.notificationTester.TestNotification(ctx, notification, payload)
	if result.Delivered && result.Message == "" {
		result.Message = "Test notification delivered."
	}
	flow.success()
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
