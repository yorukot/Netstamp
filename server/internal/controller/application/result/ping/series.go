package ping

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/controller/application/pingquery"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func (s *Service) QuerySeries(ctx context.Context, input QuerySeriesInput) (SeriesOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.ping.series.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQuerySeriesInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return SeriesOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.base.ProjectRef),
		attrProbeID.String(normalized.base.ProbeID),
		attrCheckID.String(normalized.base.CheckID),
		attrSeries.String(strings.Join(seriesKeyStrings(normalized.series), ",")),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.base.ProjectRef, normalized.base.CurrentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.series == nil {
		configuredErr := errors.New("ping result repository is not configured")
		span.SetStatus(codes.Error, "ping repository missing")
		span.RecordError(configuredErr)
		return SeriesOutput{}, configuredErr
	}

	counts, err := s.series.CountPingSeriesPoints(ctx, domainping.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping series point count failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}
	plan := pingquery.SelectReadPlan(counts, normalized.maxDataPoints)

	series, err := s.series.ListPingSeries(ctx, domainping.SeriesReadQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.base.ProbeID,
		CheckID:       normalized.base.CheckID,
		From:          normalized.base.From,
		To:            normalized.base.To,
		Series:        seriesKeyStrings(normalized.series),
		MaxDataPoints: normalized.maxDataPoints,
		Mode:          plan.Mode,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping series query failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}

	return SeriesOutput{
		Series: newSeries(series, normalized.series, normalized.base.ProbeID, normalized.base.CheckID),
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

func newSeries(series map[string]domainping.SeriesData, requested []SeriesKey, probeID, checkID string) map[string]Series {
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

func unitForSeries(key SeriesKey) string {
	switch key {
	case SeriesLatencyAvg, SeriesLatencyMin, SeriesLatencyMax:
		return "ms"
	case SeriesLossPercent:
		return "percent"
	default:
		return ""
	}
}

func seriesKeyStrings(keys []SeriesKey) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, string(key))
	}
	return values
}
