package tcp

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
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

	result, err := s.insights.ListTCPInsight(ctx, domaintcp.InsightQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.base.ProbeID,
		CheckID:       normalized.base.CheckID,
		From:          normalized.base.From,
		To:            normalized.base.To,
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "tcp insight query failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}

	return InsightOutput{
		Buckets: newInsightBuckets(result.Buckets),
		Summary: newInsightSummary(result.Summary),
		Query: resultshared.QueryMetadata{
			FromMs:        normalized.base.From.UnixMilli(),
			ToMs:          normalized.base.To.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

func newInsightBuckets(buckets []domaintcp.InsightBucket) []InsightBucket {
	values := make([]InsightBucket, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, InsightBucket{
			TimestampMs:     bucket.Timestamp.UTC().UnixMilli(),
			ResultCount:     bucket.ResultCount,
			DurationAvgMs:   bucket.DurationAvgMs,
			ConnectMinMs:    bucket.ConnectMinMs,
			ConnectAvgMs:    bucket.ConnectAvgMs,
			ConnectMedianMs: bucket.ConnectMedianMs,
			ConnectMaxMs:    bucket.ConnectMaxMs,
			ConnectStddevMs: bucket.ConnectStddevMs,
			SuccessRate:     bucket.SuccessRate,
			TimeoutCount:    bucket.TimeoutCount,
			ErrorCount:      bucket.ErrorCount,
		})
	}
	return values
}

func newInsightSummary(summary domaintcp.InsightSummary) InsightSummary {
	return InsightSummary{
		TotalResults:      summary.TotalResults,
		SuccessfulCount:   summary.SuccessfulCount,
		TimeoutCount:      summary.TimeoutCount,
		ErrorCount:        summary.ErrorCount,
		AvgConnectMs:      summary.AvgConnectMs,
		MedianConnectMs:   summary.MedianConnectMs,
		MaxConnectMs:      summary.MaxConnectMs,
		P95ConnectMs:      summary.P95ConnectMs,
		P99ConnectMs:      summary.P99ConnectMs,
		LatestStatus:      statusString(summary.LatestStatus),
		LatestStartedAtMs: resultshared.TimePtrMillis(summary.LatestStartedAt),
		LatestConnectMs:   summary.LatestConnectMs,
		LatestResolvedIP:  summary.LatestResolvedIP,
	}
}

func statusString(value *domaintcp.Status) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}
