package result

import (
	"context"

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
	Summary tcpInsightSummaryBody `json:"summary"`
	Meta    queryMetadataBody     `json:"meta"`
}

type tcpInsightSummaryBody struct {
	AverageConnectMs *float64 `json:"averageConnectMs,omitempty"`
	MaxConnectMs     *float64 `json:"maxConnectMs,omitempty"`
	FailurePercent   *float64 `json:"failurePercent,omitempty"`
	SuccessRate      *float64 `json:"successRate,omitempty"`
	Samples          int64    `json:"samples"`
}

func newQueryTCPInsightBody(output appresult.TCPInsightOutput) queryTCPInsightBody {
	return queryTCPInsightBody{
		Summary: newTCPInsightSummaryBody(output.Summary),
		Meta: queryMetadataBody{
			FromMs:        output.Meta.FromMs,
			ToMs:          output.Meta.ToMs,
			MaxDataPoints: output.Meta.MaxDataPoints,
			Source:        output.Meta.Source,
			Resolution:    output.Meta.Resolution,
			TotalPoints:   output.Meta.TotalPoints,
		},
	}
}

func newTCPInsightSummaryBody(summary appresult.TCPInsightSummary) tcpInsightSummaryBody {
	return tcpInsightSummaryBody{
		AverageConnectMs: summary.AverageConnectMs,
		MaxConnectMs:     summary.MaxConnectMs,
		FailurePercent:   summary.FailurePercent,
		SuccessRate:      summary.SuccessRate,
		Samples:          summary.Samples,
	}
}
