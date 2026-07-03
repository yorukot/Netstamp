package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

const (
	testProjectID      = "11111111-1111-1111-1111-111111111111"
	testNotificationID = "22222222-2222-2222-2222-222222222222"
	testOutboxID       = "33333333-3333-3333-3333-333333333333"
)

func TestWorkerMarksDeliveredJobsDelivered(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	repo := &recordingRepository{
		jobs:         []domainalert.NotificationOutboxJob{testOutboxJob()},
		notification: testNotification(true),
	}
	sender := &recordingSender{result: DeliveryResult{Delivered: true}}
	worker := newTestWorker(repo, sender, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	if repo.method != "delivered" || repo.outboxID != testOutboxID || !repo.deliveredAt.Equal(now) {
		t.Fatalf("unexpected repository delivery call: %#v", repo)
	}
	if sender.calls != 1 || string(sender.payload) != `{"ok":true}` {
		t.Fatalf("unexpected sender call: %#v", sender)
	}
}

func TestWorkerDiscardsDisabledNotifications(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	repo := &recordingRepository{
		jobs:         []domainalert.NotificationOutboxJob{testOutboxJob()},
		notification: testNotification(false),
	}
	sender := &recordingSender{result: DeliveryResult{Delivered: true}}
	worker := newTestWorker(repo, sender, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	assertDiscarded(t, repo, "notification", "disabled", "notification is disabled")
	if sender.calls != 0 {
		t.Fatalf("expected sender not to be called, got %d calls", sender.calls)
	}
}

func TestWorkerDiscardsNonRetryableSenderResult(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	repo := &recordingRepository{
		jobs:         []domainalert.NotificationOutboxJob{testOutboxJob()},
		notification: testNotification(true),
	}
	sender := &recordingSender{
		result: DeliveryResult{Kind: "webhook", Code: "bad_request", Message: "payload rejected"},
	}
	worker := newTestWorker(repo, sender, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	assertDiscarded(t, repo, "webhook", "bad_request", "payload rejected")
}

func TestWorkerRetriesRetryableSenderResult(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	job := testOutboxJob()
	job.AttemptCount = 1
	job.MaxAttempts = 3
	repo := &recordingRepository{
		jobs:         []domainalert.NotificationOutboxJob{job},
		notification: testNotification(true),
	}
	sender := &recordingSender{
		result: DeliveryResult{Retryable: true, Kind: "webhook", Code: "timeout", Message: "request timed out"},
	}
	worker := newTestWorker(repo, sender, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	wantNextAttempt := now.Add(2 * time.Minute)
	if repo.method != "retry" || repo.outboxID != testOutboxID || !repo.nextAttemptAt.Equal(wantNextAttempt) {
		t.Fatalf("unexpected repository retry call: %#v", repo)
	}
	assertResultDetails(t, repo, "webhook", "timeout", "request timed out")
}

func TestWorkerFailsRetryableSenderResultAtMaxAttempts(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	job := testOutboxJob()
	job.AttemptCount = 2
	job.MaxAttempts = 3
	repo := &recordingRepository{
		jobs:         []domainalert.NotificationOutboxJob{job},
		notification: testNotification(true),
	}
	sender := &recordingSender{
		result: DeliveryResult{Retryable: true, Kind: "webhook", Code: "timeout", Message: "request timed out"},
	}
	worker := newTestWorker(repo, sender, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once to succeed: %v", err)
	}

	if repo.method != "failed" || repo.outboxID != testOutboxID {
		t.Fatalf("unexpected repository failure call: %#v", repo)
	}
	assertResultDetails(t, repo, "webhook", "timeout", "request timed out")
}

func TestDisabledWorkerWaitsForContextAndReturnsNil(t *testing.T) {
	worker := NewWorker(nil, nil, WorkerConfig{Enabled: false})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("expected worker to wait for context, returned early with %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected worker to return after context cancellation")
	}
}

func newTestWorker(repo Repository, sender NotificationSender, now time.Time) *Worker {
	worker := NewWorker(repo, sender, WorkerConfig{
		Enabled:       true,
		Interval:      time.Hour,
		BatchSize:     10,
		StaleTimeout:  time.Minute,
		RetryBackoffs: []time.Duration{30 * time.Second, 2 * time.Minute},
	})
	worker.now = func() time.Time { return now }
	return worker
}

func testOutboxJob() domainalert.NotificationOutboxJob {
	return domainalert.NotificationOutboxJob{
		ID:               testOutboxID,
		ProjectID:        testProjectID,
		NotificationID:   testNotificationID,
		NotificationType: domainalert.NotificationTypeWebhook,
		Payload:          []byte(`{"ok":true}`),
		AttemptCount:     0,
		MaxAttempts:      3,
	}
}

func testNotification(enabled bool) domainalert.Notification {
	return domainalert.Notification{
		ID:        testNotificationID,
		ProjectID: testProjectID,
		Type:      domainalert.NotificationTypeWebhook,
		Enabled:   enabled,
	}
}

func assertDiscarded(t *testing.T, repo *recordingRepository, kind, code, message string) {
	t.Helper()

	if repo.method != "discarded" || repo.outboxID != testOutboxID {
		t.Fatalf("unexpected repository discard call: %#v", repo)
	}
	assertResultDetails(t, repo, kind, code, message)
}

func assertResultDetails(t *testing.T, repo *recordingRepository, kind, code, message string) {
	t.Helper()

	if repo.kind != kind || repo.code != code || repo.message != message {
		t.Fatalf("unexpected result details: kind=%q code=%q message=%q", repo.kind, repo.code, repo.message)
	}
}

type recordingRepository struct {
	jobs          []domainalert.NotificationOutboxJob
	claimErr      error
	notification  domainalert.Notification
	lookupErr     error
	method        string
	outboxID      string
	deliveredAt   time.Time
	nextAttemptAt time.Time
	kind          string
	code          string
	message       string
}

func (r *recordingRepository) ClaimOutbox(_ context.Context, _ int32, _ time.Time) ([]domainalert.NotificationOutboxJob, error) {
	return r.jobs, r.claimErr
}

func (r *recordingRepository) GetNotification(_ context.Context, projectID, notificationID string) (domainalert.Notification, error) {
	if projectID != testProjectID || notificationID != testNotificationID {
		return domainalert.Notification{}, errors.New("unexpected notification lookup")
	}
	return r.notification, r.lookupErr
}

func (r *recordingRepository) MarkOutboxDelivered(_ context.Context, id string, at time.Time) error {
	r.method = "delivered"
	r.outboxID = id
	r.deliveredAt = at
	return nil
}

func (r *recordingRepository) MarkOutboxRetry(_ context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error {
	r.method = "retry"
	r.outboxID = id
	r.nextAttemptAt = nextAttemptAt
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

func (r *recordingRepository) MarkOutboxFailed(_ context.Context, id, kind, code, message string) error {
	r.method = "failed"
	r.outboxID = id
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

func (r *recordingRepository) MarkOutboxDiscarded(_ context.Context, id, kind, code, message string) error {
	r.method = "discarded"
	r.outboxID = id
	r.kind = kind
	r.code = code
	r.message = message
	return nil
}

type recordingSender struct {
	result       DeliveryResult
	calls        int
	notification domainalert.Notification
	payload      []byte
}

func (s *recordingSender) SendNotification(_ context.Context, notification domainalert.Notification, payload []byte) DeliveryResult {
	s.calls++
	s.notification = notification
	s.payload = payload
	return s.result
}
