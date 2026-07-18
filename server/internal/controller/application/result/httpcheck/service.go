package httpcheck

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/controller/application/httpquery"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

type Service struct {
	repo     Repository
	projects resultshared.ProjectAccess
}

func NewService(repo Repository, projects resultshared.ProjectAccess) *Service {
	return &Service{repo: repo, projects: projects}
}

func (s *Service) QueryLatest(ctx context.Context, input QueryLatestInput) (LatestResultsOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.http.latest.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryLatestInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid latest HTTP result query input")
		return LatestResultsOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
	)

	project, err := s.projects.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return LatestResultsOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.repo == nil {
		configuredErr := errors.New("http result repository is not configured")
		span.SetStatus(codes.Error, "http repository missing")
		span.RecordError(configuredErr)
		return LatestResultsOutput{}, configuredErr
	}

	latest, err := s.repo.ListLatestHTTPResults(ctx, domainhttp.LatestResultQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
	})
	if err != nil {
		span.SetStatus(codes.Error, "latest HTTP result query failed")
		span.RecordError(err)
		return LatestResultsOutput{}, err
	}

	return LatestResultsOutput{Results: newLatestResults(latest.Results)}, nil
}

func (s *Service) QuerySeries(ctx context.Context, input QuerySeriesInput) (SeriesOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.http.series.query", trace.WithAttributes(
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

	project, err := s.projects.GetProjectForUser(ctx, normalized.base.ProjectRef, normalized.base.CurrentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.repo == nil {
		configuredErr := errors.New("http result repository is not configured")
		span.SetStatus(codes.Error, "http repository missing")
		span.RecordError(configuredErr)
		return SeriesOutput{}, configuredErr
	}

	rawPoints, err := s.repo.CountHTTPSeriesPoints(ctx, domainhttp.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "http series point count failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}
	plan := httpquery.SelectReadPlan(rawPoints, normalized.base.From, normalized.base.Now, normalized.maxDataPoints)

	series, err := s.repo.ListHTTPSeries(ctx, domainhttp.SeriesReadQuery{
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
		span.SetStatus(codes.Error, "http series query failed")
		span.RecordError(err)
		return SeriesOutput{}, err
	}

	totalPoints := plan.TotalPoints
	if plan.Source == domainhttp.SeriesSourceAggregate {
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
	ctx, span := resultTracer.Start(ctx, "result.http.insight.query", trace.WithAttributes(
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

	project, err := s.projects.GetProjectForUser(ctx, normalized.base.ProjectRef, normalized.base.CurrentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.repo == nil {
		configuredErr := errors.New("http result repository is not configured")
		span.SetStatus(codes.Error, "http repository missing")
		span.RecordError(configuredErr)
		return InsightOutput{}, configuredErr
	}

	rawPoints, err := s.repo.CountHTTPSeriesPoints(ctx, domainhttp.SeriesPointCountQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
	})
	if err != nil {
		span.SetStatus(codes.Error, "http insight point count failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}
	plan := httpquery.SelectReadPlan(rawPoints, normalized.base.From, normalized.base.Now, normalized.maxDataPoints)

	summary, err := s.repo.GetHTTPInsightSummary(ctx, domainhttp.InsightSummaryQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
		Source:    plan.Source,
	})
	if err != nil {
		span.SetStatus(codes.Error, "http insight query failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}

	totalPoints := plan.TotalPoints
	if plan.Source == domainhttp.SeriesSourceAggregate || summary.TotalResults > 0 {
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

func maxSeriesPointCount(series map[string]domainhttp.SeriesData) int64 {
	var maxCount int
	for _, data := range series {
		if len(data.Points) > maxCount {
			maxCount = len(data.Points)
		}
	}
	return int64(maxCount)
}

func newSeries(series map[string]domainhttp.SeriesData, requested []SeriesKey, probeID, checkID string) map[string]Series {
	values := make(map[string]Series, len(requested))
	for _, key := range requested {
		name := string(key)
		values[name] = Series{
			Name: name,
			Labels: SeriesLabels{
				ProbeID:   probeID,
				CheckID:   checkID,
				CheckType: "http",
			},
			Unit:   unitForSeries(key),
			Points: newSeriesPoints(series[name].Points),
		}
	}
	return values
}

func newSeriesPoints(points []domainhttp.SeriesPoint) []SeriesPoint {
	values := make([]SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, SeriesPoint{TimestampMs: point.Timestamp.UTC().UnixMilli(), Value: point.Value})
	}
	return values
}

func unitForSeries(key SeriesKey) string {
	if key == SeriesFailurePercent {
		return "percent"
	}
	return "ms"
}

func newInsightSummary(summary domainhttp.InsightSummary) InsightSummary {
	return InsightSummary{
		AverageTotalMs:           summary.AverageTotalMs,
		MaxTotalMs:               summary.MaxTotalMs,
		AverageTTFBMs:            summary.AverageTTFBMs,
		MaxTTFBMs:                summary.MaxTTFBMs,
		FailurePercent:           summary.FailurePercent,
		SuccessRate:              summary.SuccessRate,
		CertificateDaysRemaining: summary.CertificateDaysRemaining,
		Samples:                  summary.Samples,
	}
}

func newLatestResults(results []domainhttp.LatestResult) []LatestResult {
	values := make([]LatestResult, 0, len(results))
	for _, result := range results {
		values = append(values, LatestResult{
			ProbeID: result.ProbeID,
			CheckID: result.CheckID,
			Result:  result.Result,
		})
	}
	return values
}

func seriesKeyStrings(keys []SeriesKey) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, string(key))
	}
	return values
}
