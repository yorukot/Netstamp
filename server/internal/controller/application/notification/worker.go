package notification

import (
	"context"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type WorkerConfig struct {
	Enabled       bool
	Interval      time.Duration
	BatchSize     int32
	StaleTimeout  time.Duration
	RetryBackoffs []time.Duration
}

type Worker struct {
	repo    Repository
	webhook WebhookSender
	cfg     WorkerConfig
	now     func() time.Time
}

func NewWorker(repo Repository, webhook WebhookSender, cfg WorkerConfig) *Worker {
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
	return &Worker{repo: repo, webhook: webhook, cfg: cfg, now: func() time.Time { return time.Now().UTC() }}
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
			// Keep the worker alive; individual job errors are persisted in the outbox.
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
	for _, job := range jobs {
		w.deliver(ctx, job)
	}
	return nil
}

func (w *Worker) deliver(ctx context.Context, job domainalert.NotificationOutboxJob) {
	channel, err := w.repo.GetChannel(ctx, job.ProjectID, job.ChannelID)
	if err != nil {
		_ = w.repo.MarkOutboxRetry(ctx, job.ID, w.nextAttempt(job), "channel", "lookup_failed", "notification channel lookup failed")
		return
	}
	if !channel.Enabled {
		_ = w.repo.MarkOutboxDiscarded(ctx, job.ID, "channel", "disabled", "notification channel is disabled")
		return
	}

	var result DeliveryResult
	switch job.ChannelType {
	case domainalert.ChannelTypeWebhook:
		result = w.webhook.SendWebhook(ctx, channel, job.Payload)
	default:
		result = DeliveryResult{Retryable: false, Kind: "channel", Code: "unsupported_type", Message: "unsupported notification channel type"}
	}

	switch {
	case result.Delivered:
		_ = w.repo.MarkOutboxDelivered(ctx, job.ID, w.now())
	case !result.Retryable:
		_ = w.repo.MarkOutboxDiscarded(ctx, job.ID, result.Kind, result.Code, result.Message)
	case job.AttemptCount+1 >= job.MaxAttempts:
		_ = w.repo.MarkOutboxFailed(ctx, job.ID, result.Kind, result.Code, result.Message)
	default:
		_ = w.repo.MarkOutboxRetry(ctx, job.ID, w.nextAttempt(job), result.Kind, result.Code, result.Message)
	}
}

func (w *Worker) nextAttempt(job domainalert.NotificationOutboxJob) time.Time {
	index := int(job.AttemptCount)
	if index >= len(w.cfg.RetryBackoffs) {
		index = len(w.cfg.RetryBackoffs) - 1
	}
	return w.now().Add(w.cfg.RetryBackoffs[index])
}
