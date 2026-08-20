package updatecheck

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/yorukot/netstamp/internal/controller/application/workerloop"
	appversion "github.com/yorukot/netstamp/internal/platform/version"
)

const Interval = 6 * time.Hour

var updateCheckTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/updatecheck")

type Worker struct {
	cache    *Cache
	releases ReleaseSource
	settings SettingsReader
	log      *zap.Logger
	now      func() time.Time
}

func NewWorker(cache *Cache, releases ReleaseSource, settings SettingsReader, log *zap.Logger) *Worker {
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{
		cache: cache, releases: releases, settings: settings, log: log,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (w *Worker) Run(ctx context.Context) error {
	return workerloop.Run(ctx, true, Interval, w.RunOnce, func(err error) {
		w.log.Error("background_worker_run_failed",
			zap.String("worker.name", "update_check"),
			zap.String("worker.operation", "run_once"),
			zap.Error(err),
		)
	})
}

func (w *Worker) RunOnce(ctx context.Context) error {
	ctx, span := updateCheckTracer.Start(ctx, "updatecheck.run_once")
	defer span.End()

	checkedAt := w.now()
	if w.cache == nil {
		err := errors.New("update check cache is unavailable")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if w.settings == nil {
		return w.fail(span, checkedAt, errors.New("update check settings reader is unavailable"))
	}
	enabled, err := w.settings.UpdateCheckEnabled(ctx)
	if err != nil {
		return w.fail(span, checkedAt, fmt.Errorf("read update check setting: %w", err))
	}
	if !enabled {
		w.cache.Clear(appversion.Product)
		return nil
	}
	if w.releases == nil {
		return w.fail(span, checkedAt, errors.New("update release source is unavailable"))
	}
	release, err := w.releases.LatestRelease(ctx)
	if err != nil {
		return w.fail(span, checkedAt, fmt.Errorf("get latest release: %w", err))
	}
	comparison, err := appversion.Compare(appversion.Product, release.TagName)
	if err != nil {
		return w.fail(span, checkedAt, fmt.Errorf("compare release versions: %w", err))
	}
	w.cache.StoreSuccess(Status{
		CurrentVersion:  appversion.Product,
		LatestVersion:   release.TagName,
		UpdateAvailable: comparison < 0,
		ReleaseURL:      release.URL,
		PublishedAt:     release.PublishedAt,
		LastCheckedAt:   checkedAt,
	})
	return nil
}

func (w *Worker) fail(span trace.Span, checkedAt time.Time, err error) error {
	w.cache.StoreFailure(appversion.Product, checkedAt, err)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
