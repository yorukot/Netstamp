package notification

import (
	"context"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type Repository interface {
	ClaimOutbox(ctx context.Context, limit int32, staleBefore time.Time) ([]domainalert.NotificationOutboxJob, error)
	GetChannel(ctx context.Context, projectID, channelID string) (domainalert.NotificationChannel, error)
	MarkOutboxDelivered(ctx context.Context, id string, at time.Time) error
	MarkOutboxRetry(ctx context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error
	MarkOutboxFailed(ctx context.Context, id, kind, code, message string) error
	MarkOutboxDiscarded(ctx context.Context, id, kind, code, message string) error
}

type ChannelSender interface {
	SendChannel(ctx context.Context, channel domainalert.NotificationChannel, payload []byte) DeliveryResult
}

type DeliveryResult struct {
	Delivered bool
	Retryable bool
	Kind      string
	Code      string
	Message   string
}
