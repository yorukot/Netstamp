package result

import (
	"context"
	"net/netip"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryTCPInsight(ctx context.Context, input *queryTCPInsightInput) (*queryTCPInsightOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryTCPInsight(ctx, appresult.QueryTCPInsightInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	})
	if err != nil {
		return nil, mapResultError(err, "query tcp insight failed")
	}

	return &queryTCPInsightOutput{Body: newQueryTCPInsightBody(output)}, nil
}

type queryTCPInsightInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	MaxDataPoints int32
}

type queryTCPInsightOutput struct {
	Body queryTCPInsightBody
}

type queryTCPInsightBody struct {
	Buckets []tcpInsightBucketBody `json:"buckets"`
	Summary tcpInsightSummaryBody  `json:"summary"`
	Query   queryMetadataBody      `json:"query"`
}

type tcpInsightBucketBody struct {
	TimestampMs     int64    `json:"timestampMs"`
	ResultCount     int64    `json:"resultCount"`
	DurationAvgMs   *float64 `json:"durationAvgMs,omitempty"`
	ConnectMinMs    *float64 `json:"connectMinMs,omitempty"`
	ConnectAvgMs    *float64 `json:"connectAvgMs,omitempty"`
	ConnectMedianMs *float64 `json:"connectMedianMs,omitempty"`
	ConnectMaxMs    *float64 `json:"connectMaxMs,omitempty"`
	ConnectStddevMs *float64 `json:"connectStddevMs,omitempty"`
	SuccessRate     *float64 `json:"successRate,omitempty"`
	TimeoutCount    int64    `json:"timeoutCount"`
	ErrorCount      int64    `json:"errorCount"`
}

type tcpInsightSummaryBody struct {
	TotalResults      int64       `json:"totalResults"`
	SuccessfulCount   int64       `json:"successfulCount"`
	TimeoutCount      int64       `json:"timeoutCount"`
	ErrorCount        int64       `json:"errorCount"`
	AvgConnectMs      *float64    `json:"avgConnectMs,omitempty"`
	MedianConnectMs   *float64    `json:"medianConnectMs,omitempty"`
	MaxConnectMs      *float64    `json:"maxConnectMs,omitempty"`
	P95ConnectMs      *float64    `json:"p95ConnectMs,omitempty"`
	P99ConnectMs      *float64    `json:"p99ConnectMs,omitempty"`
	LatestStatus      *string     `json:"latestStatus,omitempty"`
	LatestStartedAtMs *int64      `json:"latestStartedAtMs,omitempty"`
	LatestConnectMs   *float64    `json:"latestConnectMs,omitempty"`
	LatestResolvedIP  *netip.Addr `json:"latestResolvedIp,omitempty"`
}

func newQueryTCPInsightBody(output appresult.TCPInsightOutput) queryTCPInsightBody {
	return queryTCPInsightBody{
		Buckets: newTCPInsightBucketsBody(output.Buckets),
		Summary: newTCPInsightSummaryBody(output.Summary),
		Query: queryMetadataBody{
			FromMs:        output.Query.FromMs,
			ToMs:          output.Query.ToMs,
			MaxDataPoints: output.Query.MaxDataPoints,
			Resolution:    output.Query.Resolution,
			TotalPoints:   output.Query.TotalPoints,
		},
	}
}

func newTCPInsightBucketsBody(buckets []appresult.TCPInsightBucket) []tcpInsightBucketBody {
	values := make([]tcpInsightBucketBody, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, tcpInsightBucketBody{
			TimestampMs:     bucket.TimestampMs,
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

func newTCPInsightSummaryBody(summary appresult.TCPInsightSummary) tcpInsightSummaryBody {
	return tcpInsightSummaryBody{
		TotalResults:      summary.TotalResults,
		SuccessfulCount:   summary.SuccessfulCount,
		TimeoutCount:      summary.TimeoutCount,
		ErrorCount:        summary.ErrorCount,
		AvgConnectMs:      summary.AvgConnectMs,
		MedianConnectMs:   summary.MedianConnectMs,
		MaxConnectMs:      summary.MaxConnectMs,
		P95ConnectMs:      summary.P95ConnectMs,
		P99ConnectMs:      summary.P99ConnectMs,
		LatestStatus:      summary.LatestStatus,
		LatestStartedAtMs: summary.LatestStartedAtMs,
		LatestConnectMs:   summary.LatestConnectMs,
		LatestResolvedIP:  summary.LatestResolvedIP,
	}
}
