package result

import (
	"context"
	"errors"
	"strings"

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
		attrSeries.String(strings.Join(pingSeriesKeyStrings(normalized.series), ",")),
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
		Series:        pingSeriesKeyStrings(normalized.series),
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping series query failed")
		span.RecordError(err)
		return PingSeriesOutput{}, err
	}

	return PingSeriesOutput{
		Series: newPingSeries(result.Series, normalized.series, normalized.probeID, normalized.checkID),
		Meta: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Source:        string(result.Source),
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
		Summary: newPingInsightSummary(result.Summary),
		Meta: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Source:        string(result.Source),
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

func newPingSeries(series map[string]domainping.SeriesData, requested []PingSeriesKey, probeID, checkID string) map[string]Series {
	values := make(map[string]Series, len(requested))
	for _, key := range requested {
		name := string(key)
		data := series[name]
		values[name] = Series{
			Name: name,
			Labels: SeriesLabels{
				ProbeID:   probeID,
				CheckID:   checkID,
				CheckType: "ping",
			},
			Unit:   unitForSeries(key),
			Points: newSeriesPoints(data.Points),
		}
	}
	return values
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

func newPingInsightSummary(summary domainping.InsightSummary) PingInsightSummary {
	return PingInsightSummary{
		AverageRttMs: summary.AverageRttMs,
		MaxRttMs:     summary.MaxRttMs,
		LossPercent:  summary.LossPercent,
		SuccessRate:  summary.SuccessRate,
		Samples:      summary.Samples,
	}
}

func unitForSeries(key PingSeriesKey) string {
	switch key {
	case PingSeriesLatencyAvg, PingSeriesLatencyMin, PingSeriesLatencyMax:
		return "ms"
	case PingSeriesLossPercent:
		return "percent"
	default:
		return ""
	}
}

func pingSeriesKeyStrings(keys []PingSeriesKey) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, string(key))
	}
	return values
}
