package result

import (
	"context"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryPingSeries(ctx context.Context, input *queryPingSeriesInput) (*queryPingSeriesOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryPingSeries(ctx, appresult.QueryPingSeriesInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		Metric:        input.Metric,
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	})
	if err != nil {
		return nil, mapResultError(err, "query ping result series failed")
	}

	return &queryPingSeriesOutput{Body: newQueryPingSeriesBody(output)}, nil
}

type queryPingSeriesInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	Metric        string
	MaxDataPoints int32
}

type queryPingSeriesOutput struct {
	Body queryPingSeriesBody
}

type queryPingSeriesBody struct {
	Series []seriesBody      `json:"series"`
	Query  queryMetadataBody `json:"query"`
}

type seriesBody struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Unit   string            `json:"unit"`
	Points []pointTuple      `json:"points"`
}

type pointTuple [2]float64

type queryMetadataBody struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Resolution    string `json:"resolution"`
	TotalPoints   int64  `json:"totalPoints"`
}

func newQueryPingSeriesBody(output appresult.PingSeriesOutput) queryPingSeriesBody {
	series := make([]seriesBody, 0, len(output.Series))
	for _, value := range output.Series {
		points := make([]pointTuple, 0, len(value.Points))
		for _, point := range value.Points {
			points = append(points, pointTuple{float64(point.TimestampMs), point.Value})
		}
		series = append(series, seriesBody{
			Name: value.Name,
			Labels: map[string]string{
				"probeId":   value.Labels.ProbeID,
				"checkId":   value.Labels.CheckID,
				"checkType": value.Labels.CheckType,
			},
			Unit:   value.Unit,
			Points: points,
		})
	}

	return queryPingSeriesBody{
		Series: series,
		Query: queryMetadataBody{
			FromMs:        output.Query.FromMs,
			ToMs:          output.Query.ToMs,
			MaxDataPoints: output.Query.MaxDataPoints,
			Resolution:    output.Query.Resolution,
			TotalPoints:   output.Query.TotalPoints,
		},
	}
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func optionalInt32(value int32) *int32 {
	if value == 0 {
		return nil
	}
	return &value
}
