package notification

import (
	"context"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type Repository interface {
	ClaimOutbox(ctx context.Context, limit int32, staleBefore time.Time) ([]domainalert.NotificationOutboxJob, error)
	GetNotification(ctx context.Context, projectID, notificationID string) (domainalert.Notification, error)
	MarkOutboxDelivered(ctx context.Context, id string, at time.Time) error
	MarkOutboxRetry(ctx context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error
	MarkOutboxFailed(ctx context.Context, id, kind, code, message string) error
	MarkOutboxDiscarded(ctx context.Context, id, kind, code, message string) error
}

type NotificationSender interface {
	SendNotification(ctx context.Context, notification domainalert.Notification, payload []byte) DeliveryResult
}

type DeliveryResult struct {
	Delivered bool
	Retryable bool
	Kind      string
	Code      string
	Message   string
}
