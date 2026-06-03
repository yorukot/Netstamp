package tcp

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	"github.com/yorukot/netstamp/internal/controller/application/tcpquery"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type Service struct {
	insights      InsightRepository
	projectAccess resultshared.ProjectAccess
}

func NewService(insights InsightRepository, projectAccess resultshared.ProjectAccess) *Service {
	return &Service{
		insights:      insights,
		projectAccess: projectAccess,
	}
}

func (s *Service) QuerySeries(ctx context.Context, input QuerySeriesInput) (SeriesOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.tcp.series.query", trace.WithAttributes(
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

	if s.insights == nil {
		configuredErr := errors.New("tcp result repository is not configured")
		span.SetStatus(codes.Error, "tcp repository missing")
		span.RecordError(configuredErr)
		return SeriesOutput{}, configuredErr
	}

	rawPoints, err := s.insights.CountTCPSeriesPoints(ctx, domaintcp.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "tcp series point count failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}
	plan := tcpquery.SelectReadPlan(rawPoints, normalized.base.From, normalized.base.Now, normalized.maxDataPoints)

	series, err := s.insights.ListTCPSeries(ctx, domaintcp.SeriesReadQuery{
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
		span.SetStatus(codes.Error, "tcp series query failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}

	totalPoints := plan.TotalPoints
	if plan.Source == domaintcp.SeriesSourceAggregate {
		totalPoints = maxSeriesPointCount(series)
	}

	return SeriesOutput{
		Series: newSeries(series, normalized.series, normalized.base.ProbeID, normalized.base.CheckID),
		Meta: resultshared.QueryMetadata{
			FromMs:        normalized.base.From.UnixMilli(),
			ToMs:          normalized.base.To.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Source:        string(plan.Source),
			Resolution:    string(plan.Resolution),
			TotalPoints:   totalPoints,
		},
	}, nil
}

func (s *Service) QueryInsight(ctx context.Context, input QueryInsightInput) (InsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.tcp.insight.query", trace.WithAttributes(
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

	if s.insights == nil {
		configuredErr := errors.New("tcp result repository is not configured")
		span.SetStatus(codes.Error, "tcp repository missing")
		span.RecordError(configuredErr)
		return InsightOutput{}, configuredErr
	}

	rawPoints, err := s.insights.CountTCPSeriesPoints(ctx, domaintcp.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "tcp insight point count failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	plan := tcpquery.SelectReadPlan(rawPoints, normalized.base.From, normalized.base.Now, normalized.maxDataPoints)

	summary, err := s.insights.GetTCPInsightSummary(ctx, domaintcp.InsightSummaryQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
		Source:    plan.Source,
	})
	if err != nil {
		span.SetStatus(codes.Error, "tcp insight query failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	totalPoints := plan.TotalPoints
	if plan.Source == domaintcp.SeriesSourceAggregate || summary.TotalResults > 0 {
		totalPoints = summary.TotalResults
	}

	return InsightOutput{
		Summary: newInsightSummary(summary),
		Meta: resultshared.QueryMetadata{
			FromMs:        normalized.base.From.UnixMilli(),
			ToMs:          normalized.base.To.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Source:        string(plan.Source),
			Resolution:    string(plan.Resolution),
			TotalPoints:   totalPoints,
		},
	}, nil
}

func maxSeriesPointCount(series map[string]domaintcp.SeriesData) int64 {
	var maxCount int
	for _, data := range series {
		if len(data.Points) > maxCount {
			maxCount = len(data.Points)
		}
	}
	return int64(maxCount)
}

func newSeries(series map[string]domaintcp.SeriesData, requested []SeriesKey, probeID, checkID string) map[string]Series {
	values := make(map[string]Series, len(requested))
	for _, key := range requested {
		name := string(key)
		data := series[name]
		values[name] = Series{
			Name: name,
			Labels: SeriesLabels{
				ProbeID:   probeID,
				CheckID:   checkID,
				CheckType: "tcp",
			},
			Unit:   unitForSeries(key),
			Points: newSeriesPoints(data.Points),
		}
	}
	return values
}

func newSeriesPoints(points []domaintcp.SeriesPoint) []SeriesPoint {
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
	case SeriesConnectAvg, SeriesConnectMin, SeriesConnectMax:
		return "ms"
	case SeriesFailurePercent:
		return "percent"
	default:
		return ""
	}
}

func newInsightSummary(summary domaintcp.InsightSummary) InsightSummary {
	return InsightSummary{
		AverageConnectMs: summary.AverageConnectMs,
		MaxConnectMs:     summary.MaxConnectMs,
		FailurePercent:   summary.FailurePercent,
		SuccessRate:      summary.SuccessRate,
		Samples:          summary.Samples,
	}
}

func seriesKeyStrings(keys []SeriesKey) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, string(key))
	}
	return values
}
