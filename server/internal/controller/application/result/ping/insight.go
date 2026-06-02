package ping

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/controller/application/pingquery"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func (s *Service) QueryInsight(ctx context.Context, input QueryInsightInput) (InsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.ping.insight.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryInsightInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return InsightOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.base.ProjectRef),
		attrProbeID.String(normalized.base.ProbeID),
		attrCheckID.String(normalized.base.CheckID),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.base.ProjectRef, normalized.base.CurrentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.series == nil {
		configuredErr := errors.New("ping result repository is not configured")
		span.SetStatus(codes.Error, "ping repository missing")
		span.RecordError(configuredErr)
		return InsightOutput{}, configuredErr
	}

	rawPoints, rollupPoints, err := s.series.CountPingSeriesPoints(ctx, domainping.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping insight point count failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	plan := pingquery.SelectReadPlan(rawPoints, rollupPoints, normalized.maxDataPoints)

	summary, err := s.series.GetPingInsightSummary(ctx, domainping.InsightSummaryQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
		Source:    plan.Source,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping insight query failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}

	return InsightOutput{
		Summary: newInsightSummary(summary),
		Meta: resultshared.QueryMetadata{
			FromMs:        normalized.base.From.UnixMilli(),
			ToMs:          normalized.base.To.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Source:        string(plan.Source),
			Resolution:    string(plan.Resolution),
			TotalPoints:   plan.TotalPoints,
		},
	}, nil
}

func newInsightSummary(summary domainping.InsightSummary) InsightSummary {
	return InsightSummary{
		AverageRttMs: summary.AverageRttMs,
		MaxRttMs:     summary.MaxRttMs,
		LossPercent:  summary.LossPercent,
		SuccessRate:  summary.SuccessRate,
		Samples:      summary.Samples,
	}
}
