package notification

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type WorkerConfig struct {
	Enabled       bool
	Interval      time.Duration
	BatchSize     int32
	StaleTimeout  time.Duration
	RetryBackoffs []time.Duration
	Log           *zap.Logger
}

type Worker struct {
	repo   Repository
	sender NotificationSender
	cfg    WorkerConfig
	now    func() time.Time
	log    *zap.Logger
}

func NewWorker(repo Repository, sender NotificationSender, cfg WorkerConfig) *Worker {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 25
	}
	if cfg.StaleTimeout <= 0 {
		cfg.StaleTimeout = time.Minute
	}
	if len(cfg.RetryBackoffs) == 0 {
		cfg.RetryBackoffs = []time.Duration{30 * time.Second, 2 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	}
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{repo: repo, sender: sender, cfg: cfg, now: func() time.Time { return time.Now().UTC() }, log: log}
}

func (w *Worker) Run(ctx context.Context) error {
	if !w.cfg.Enabled {
		<-ctx.Done()
		return nil
	}
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		if err := w.RunOnce(ctx); err != nil {
			w.log.Error("background_worker_run_failed",
				zap.String("worker.name", "notification_outbox"),
				zap.String("worker.operation", "run_once"),
				zap.Error(err),
			)
			// Keep the worker alive; individual job errors are persisted in the outbox.
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	jobs, err := w.repo.ClaimOutbox(ctx, w.cfg.BatchSize, w.now().Add(-w.cfg.StaleTimeout))
	if err != nil {
		return err
	}
	var errs []error
	for _, job := range jobs {
		if err := w.deliver(ctx, job); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (w *Worker) deliver(ctx context.Context, job domainalert.NotificationOutboxJob) error {
	notification, err := w.repo.GetNotification(ctx, job.ProjectID, job.NotificationID)
	if err != nil {
		return w.repo.MarkOutboxRetry(ctx, job.ID, w.nextAttempt(job), "notification", "lookup_failed", "notification lookup failed")
	}
	if !notification.Enabled {
		return w.repo.MarkOutboxDiscarded(ctx, job.ID, "notification", "disabled", "notification is disabled")
	}

	result := DeliveryResult{Retryable: false, Kind: "notification", Code: "sender_unavailable", Message: "notification sender is unavailable"}
	if w.sender != nil {
		result = w.sender.SendNotification(ctx, notification, job.Payload)
	}

	switch {
	case result.Delivered:
		return w.repo.MarkOutboxDelivered(ctx, job.ID, w.now())
	case !result.Retryable:
		return w.repo.MarkOutboxDiscarded(ctx, job.ID, result.Kind, result.Code, result.Message)
	case job.AttemptCount+1 >= job.MaxAttempts:
		return w.repo.MarkOutboxFailed(ctx, job.ID, result.Kind, result.Code, result.Message)
	default:
		return w.repo.MarkOutboxRetry(ctx, job.ID, w.nextAttempt(job), result.Kind, result.Code, result.Message)
	}
}

func (w *Worker) nextAttempt(job domainalert.NotificationOutboxJob) time.Time {
	index := int(job.AttemptCount)
	if index >= len(w.cfg.RetryBackoffs) {
		index = len(w.cfg.RetryBackoffs) - 1
	}
	return w.now().Add(w.cfg.RetryBackoffs[index])
}
