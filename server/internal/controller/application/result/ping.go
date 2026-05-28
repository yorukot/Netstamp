package result

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func (s *Service) QueryPingSeries(ctx context.Context, input QueryPingSeriesInput) (PingSeriesOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.ping.series.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryPingSeriesInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return PingSeriesOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
		attrMetric.String(string(normalized.metric)),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return PingSeriesOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.pings == nil {
		configuredErr := errors.New("ping result repository is not configured")
		span.SetStatus(codes.Error, "ping repository missing")
		span.RecordError(configuredErr)
		return PingSeriesOutput{}, configuredErr
	}

	result, err := s.pings.ListPingSeries(ctx, domainping.SeriesQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		CheckID:       normalized.checkID,
		From:          normalized.from,
		To:            normalized.to,
		Metric:        string(normalized.metric),
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping series query failed")
		span.RecordError(err)
		return PingSeriesOutput{}, err
	}

	return PingSeriesOutput{
		Series: []Series{{
			Name: string(normalized.metric),
			Labels: SeriesLabels{
				ProbeID:   normalized.probeID,
				CheckID:   normalized.checkID,
				CheckType: "ping",
			},
			Unit:   unitForMetric(normalized.metric),
			Points: newSeriesPoints(result.Points),
		}},
		Query: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

func (s *Service) QueryPingInsight(ctx context.Context, input QueryPingInsightInput) (PingInsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.ping.insight.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryPingInsightInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return PingInsightOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return PingInsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.pings == nil {
		configuredErr := errors.New("ping result repository is not configured")
		span.SetStatus(codes.Error, "ping repository missing")
		span.RecordError(configuredErr)
		return PingInsightOutput{}, configuredErr
	}

	result, err := s.pings.ListPingInsight(ctx, domainping.InsightQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		CheckID:       normalized.checkID,
		From:          normalized.from,
		To:            normalized.to,
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping insight query failed")
		span.RecordError(err)
		return PingInsightOutput{}, err
	}

	return PingInsightOutput{
		Buckets:       newPingInsightBuckets(result.Buckets),
		SampleDensity: newPingSampleDensity(result.SampleDensity),
		Summary:       newPingInsightSummary(result.Summary),
		Query: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

func newSeriesPoints(points []domainping.SeriesPoint) []SeriesPoint {
	values := make([]SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, SeriesPoint{
			TimestampMs: point.Timestamp.UTC().UnixMilli(),
			Value:       point.Value,
		})
	}
	return values
}

func newPingInsightBuckets(buckets []domainping.InsightBucket) []PingInsightBucket {
	values := make([]PingInsightBucket, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, PingInsightBucket{
			TimestampMs:   bucket.Timestamp.UTC().UnixMilli(),
			ResultCount:   bucket.ResultCount,
			DurationAvgMs: bucket.DurationAvgMs,
			RttMinMs:      bucket.RttMinMs,
			RttAvgMs:      bucket.RttAvgMs,
			RttMedianMs:   bucket.RttMedianMs,
			RttMaxMs:      bucket.RttMaxMs,
			RttStddevMs:   bucket.RttStddevMs,
			LossPercent:   bucket.LossPercent,
			SuccessRate:   bucket.SuccessRate,
			SentCount:     bucket.SentCount,
			ReceivedCount: bucket.ReceivedCount,
			TimeoutCount:  bucket.TimeoutCount,
			ErrorCount:    bucket.ErrorCount,
		})
	}
	return values
}

func newPingSampleDensity(cells []domainping.SampleDensityCell) []PingSampleDensityCell {
	values := make([]PingSampleDensityCell, 0, len(cells))
	for _, cell := range cells {
		values = append(values, PingSampleDensityCell{
			TimestampMs:      cell.Timestamp.UTC().UnixMilli(),
			RttBucketStartMs: cell.RttBucketStartMs,
			RttBucketEndMs:   cell.RttBucketEndMs,
			SampleCount:      cell.SampleCount,
		})
	}
	return values
}

func newPingInsightSummary(summary domainping.InsightSummary) PingInsightSummary {
	return PingInsightSummary{
		TotalResults:      summary.TotalResults,
		SuccessfulCount:   summary.SuccessfulCount,
		TimeoutCount:      summary.TimeoutCount,
		ErrorCount:        summary.ErrorCount,
		SentCount:         summary.SentCount,
		ReceivedCount:     summary.ReceivedCount,
		AvgLossPercent:    summary.AvgLossPercent,
		AvgRttMs:          summary.AvgRttMs,
		MedianRttMs:       summary.MedianRttMs,
		MaxRttMs:          summary.MaxRttMs,
		P95RttMs:          summary.P95RttMs,
		P99RttMs:          summary.P99RttMs,
		LatestStatus:      pingStatusString(summary.LatestStatus),
		LatestStartedAtMs: timePtrMillis(summary.LatestStartedAt),
		LatestRttAvgMs:    summary.LatestRttAvgMs,
		LatestLossPercent: summary.LatestLossPercent,
		LatestResolvedIP:  summary.LatestResolvedIP,
	}
}

func pingStatusString(value *domainping.Status) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}

func unitForMetric(metric PingMetric) string {
	switch metric {
	case PingMetricRTTAvgMS:
		return "ms"
	case PingMetricLossPercent, PingMetricSuccessRate:
		return "percent"
	default:
		return ""
	}
}
