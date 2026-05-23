package result

import (
	"context"
	"net/netip"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryPingInsight(ctx context.Context, input *queryPingInsightInput) (*queryPingInsightOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryPingInsight(ctx, appresult.QueryPingInsightInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	})
	if err != nil {
		return nil, mapResultError(err, "query ping insight failed")
	}

	return &queryPingInsightOutput{Body: newQueryPingInsightBody(output)}, nil
}

type queryPingInsightInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	MaxDataPoints int32
}

type queryPingInsightOutput struct {
	Body queryPingInsightBody
}

type queryPingInsightBody struct {
	Buckets       []pingInsightBucketBody     `json:"buckets"`
	SampleDensity []pingSampleDensityCellBody `json:"sampleDensity"`
	Summary       pingInsightSummaryBody      `json:"summary"`
	Query         queryMetadataBody           `json:"query"`
}

type pingInsightBucketBody struct {
	TimestampMs   int64    `json:"timestampMs"`
	ResultCount   int64    `json:"resultCount"`
	DurationAvgMs *float64 `json:"durationAvgMs,omitempty"`
	RttMinMs      *float64 `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64 `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64 `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64 `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64 `json:"rttStddevMs,omitempty"`
	LossPercent   *float64 `json:"lossPercent,omitempty"`
	SuccessRate   *float64 `json:"successRate,omitempty"`
	SentCount     int64    `json:"sentCount"`
	ReceivedCount int64    `json:"receivedCount"`
	TimeoutCount  int64    `json:"timeoutCount"`
	ErrorCount    int64    `json:"errorCount"`
}

type pingSampleDensityCellBody struct {
	TimestampMs      int64   `json:"timestampMs"`
	RttBucketStartMs float64 `json:"rttBucketStartMs"`
	RttBucketEndMs   float64 `json:"rttBucketEndMs"`
	SampleCount      int64   `json:"sampleCount"`
}

type pingInsightSummaryBody struct {
	TotalResults      int64       `json:"totalResults"`
	SuccessfulCount   int64       `json:"successfulCount"`
	TimeoutCount      int64       `json:"timeoutCount"`
	ErrorCount        int64       `json:"errorCount"`
	SentCount         int64       `json:"sentCount"`
	ReceivedCount     int64       `json:"receivedCount"`
	AvgLossPercent    *float64    `json:"avgLossPercent,omitempty"`
	AvgRttMs          *float64    `json:"avgRttMs,omitempty"`
	MedianRttMs       *float64    `json:"medianRttMs,omitempty"`
	MaxRttMs          *float64    `json:"maxRttMs,omitempty"`
	P95RttMs          *float64    `json:"p95RttMs,omitempty"`
	P99RttMs          *float64    `json:"p99RttMs,omitempty"`
	LatestStatus      *string     `json:"latestStatus,omitempty"`
	LatestStartedAtMs *int64      `json:"latestStartedAtMs,omitempty"`
	LatestRttAvgMs    *float64    `json:"latestRttAvgMs,omitempty"`
	LatestLossPercent *float64    `json:"latestLossPercent,omitempty"`
	LatestResolvedIP  *netip.Addr `json:"latestResolvedIp,omitempty"`
}

func newQueryPingInsightBody(output appresult.PingInsightOutput) queryPingInsightBody {
	return queryPingInsightBody{
		Buckets:       newPingInsightBucketsBody(output.Buckets),
		SampleDensity: newPingSampleDensityBody(output.SampleDensity),
		Summary:       newPingInsightSummaryBody(output.Summary),
		Query: queryMetadataBody{
			FromMs:        output.Query.FromMs,
			ToMs:          output.Query.ToMs,
			MaxDataPoints: output.Query.MaxDataPoints,
			Resolution:    output.Query.Resolution,
			TotalPoints:   output.Query.TotalPoints,
		},
	}
}

func newPingInsightBucketsBody(buckets []appresult.PingInsightBucket) []pingInsightBucketBody {
	values := make([]pingInsightBucketBody, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, pingInsightBucketBody{
			TimestampMs:   bucket.TimestampMs,
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

func newPingSampleDensityBody(cells []appresult.PingSampleDensityCell) []pingSampleDensityCellBody {
	values := make([]pingSampleDensityCellBody, 0, len(cells))
	for _, cell := range cells {
		values = append(values, pingSampleDensityCellBody{
			TimestampMs:      cell.TimestampMs,
			RttBucketStartMs: cell.RttBucketStartMs,
			RttBucketEndMs:   cell.RttBucketEndMs,
			SampleCount:      cell.SampleCount,
		})
	}
	return values
}

func newPingInsightSummaryBody(summary appresult.PingInsightSummary) pingInsightSummaryBody {
	return pingInsightSummaryBody{
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
		LatestStatus:      summary.LatestStatus,
		LatestStartedAtMs: summary.LatestStartedAtMs,
		LatestRttAvgMs:    summary.LatestRttAvgMs,
		LatestLossPercent: summary.LatestLossPercent,
		LatestResolvedIP:  summary.LatestResolvedIP,
	}
}
