package result

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func (s *Service) QueryTCPInsight(ctx context.Context, input QueryTCPInsightInput) (TCPInsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.tcp.insight.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryTCPInsightInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return TCPInsightOutput{}, err
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
		return TCPInsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.tcps == nil {
		configuredErr := errors.New("tcp result repository is not configured")
		span.SetStatus(codes.Error, "tcp repository missing")
		span.RecordError(configuredErr)
		return TCPInsightOutput{}, configuredErr
	}

	result, err := s.tcps.ListTCPInsight(ctx, domaintcp.InsightQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		CheckID:       normalized.checkID,
		From:          normalized.from,
		To:            normalized.to,
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "tcp insight query failed")
		span.RecordError(err)
		return TCPInsightOutput{}, err
	}

	return TCPInsightOutput{
		Buckets: newTCPInsightBuckets(result.Buckets),
		Summary: newTCPInsightSummary(result.Summary),
		Query: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

func newTCPInsightBuckets(buckets []domaintcp.InsightBucket) []TCPInsightBucket {
	values := make([]TCPInsightBucket, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, TCPInsightBucket{
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

func newTCPInsightSummary(summary domaintcp.InsightSummary) TCPInsightSummary {
	return TCPInsightSummary{
		TotalResults:      summary.TotalResults,
		SuccessfulCount:   summary.SuccessfulCount,
		TimeoutCount:      summary.TimeoutCount,
		ErrorCount:        summary.ErrorCount,
		AvgConnectMs:      summary.AvgConnectMs,
		MedianConnectMs:   summary.MedianConnectMs,
		MaxConnectMs:      summary.MaxConnectMs,
		P95ConnectMs:      summary.P95ConnectMs,
		P99ConnectMs:      summary.P99ConnectMs,
		LatestStatus:      tcpStatusString(summary.LatestStatus),
		LatestStartedAtMs: timePtrMillis(summary.LatestStartedAt),
		LatestConnectMs:   summary.LatestConnectMs,
		LatestResolvedIP:  summary.LatestResolvedIP,
	}
}

func tcpStatusString(value *domaintcp.Status) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}
