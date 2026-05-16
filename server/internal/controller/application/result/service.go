package result

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type Service struct {
	pings         PingSeriesRepository
	projectAccess ProjectAccess
}

func NewService(pings PingSeriesRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:         pings,
		projectAccess: projectAccess,
	}
}

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
