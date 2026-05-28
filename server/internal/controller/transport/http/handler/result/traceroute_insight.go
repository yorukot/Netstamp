package result

import (
	"context"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryTracerouteInsight(ctx context.Context, input *queryTracerouteInsightInput) (*queryTracerouteInsightOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryTracerouteInsight(ctx, appresult.QueryTracerouteInsightInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	})
	if err != nil {
		return nil, mapResultError(err, "query traceroute insight failed")
	}

	return &queryTracerouteInsightOutput{Body: newQueryTracerouteInsightBody(output)}, nil
}

type queryTracerouteInsightInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	MaxDataPoints int32
}

type queryTracerouteInsightOutput struct {
	Body queryTracerouteInsightBody
}

type queryTracerouteInsightBody struct {
	Points []tracerouteInsightPointBody       `json:"points"`
	Query  tracerouteInsightQueryMetadataBody `json:"query"`
}

type tracerouteInsightPointBody struct {
	TimestampMs        int64      `json:"timestampMs"`
	BucketFromMs       int64      `json:"bucketFromMs"`
	BucketToMs         int64      `json:"bucketToMs"`
	RunStartedAt       *time.Time `json:"runStartedAt,omitempty"`
	ResultCount        int64      `json:"resultCount"`
	FinalRttAvgMs      *float64   `json:"finalRttAvgMs,omitempty"`
	FinalLossPercent   *float64   `json:"finalLossPercent,omitempty"`
	HasLoss            bool       `json:"hasLoss"`
	HasRouteChange     bool       `json:"hasRouteChange"`
	DestinationReached bool       `json:"destinationReached"`
}

type tracerouteInsightQueryMetadataBody struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Resolution    string `json:"resolution"`
	TotalRuns     int64  `json:"totalRuns"`
}

func newQueryTracerouteInsightBody(output appresult.TracerouteInsightOutput) queryTracerouteInsightBody {
	points := make([]tracerouteInsightPointBody, 0, len(output.Points))
	for _, point := range output.Points {
		points = append(points, tracerouteInsightPointBody{
			TimestampMs:        point.TimestampMs,
			BucketFromMs:       point.BucketFromMs,
			BucketToMs:         point.BucketToMs,
			RunStartedAt:       point.RunStartedAt,
			ResultCount:        point.ResultCount,
			FinalRttAvgMs:      point.FinalRttAvgMs,
			FinalLossPercent:   point.FinalLossPercent,
			HasLoss:            point.HasLoss,
			HasRouteChange:     point.HasRouteChange,
			DestinationReached: point.DestinationReached,
		})
	}

	return queryTracerouteInsightBody{
		Points: points,
		Query: tracerouteInsightQueryMetadataBody{
			FromMs:        output.Query.FromMs,
			ToMs:          output.Query.ToMs,
			MaxDataPoints: output.Query.MaxDataPoints,
			Resolution:    output.Query.Resolution,
			TotalRuns:     output.Query.TotalRuns,
		},
	}
}
