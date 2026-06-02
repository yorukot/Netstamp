package result

import (
	"context"

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
	Summary pingInsightSummaryBody `json:"summary"`
	Meta    queryMetadataBody      `json:"meta"`
}

type pingInsightSummaryBody struct {
	AverageRttMs *float64 `json:"averageRttMs,omitempty"`
	MaxRttMs     *float64 `json:"maxRttMs,omitempty"`
	LossPercent  *float64 `json:"lossPercent,omitempty"`
	SuccessRate  *float64 `json:"successRate,omitempty"`
	Samples      int64    `json:"samples"`
}

func newQueryPingInsightBody(output appresult.PingInsightOutput) queryPingInsightBody {
	return queryPingInsightBody{
		Summary: newPingInsightSummaryBody(output.Summary),
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

func newPingInsightSummaryBody(summary appresult.PingInsightSummary) pingInsightSummaryBody {
	return pingInsightSummaryBody{
		AverageRttMs: summary.AverageRttMs,
		MaxRttMs:     summary.MaxRttMs,
		LossPercent:  summary.LossPercent,
		SuccessRate:  summary.SuccessRate,
		Samples:      summary.Samples,
	}
}
