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
	Ref           string `path:"ref" doc:"Project ID or slug." example:"vector-ix"`
	ProbeID       string `query:"probeId" required:"true" format:"uuid" doc:"Probe ID to query."`
	CheckID       string `query:"checkId" required:"true" format:"uuid" doc:"Check ID to query."`
	From          int64  `query:"from" doc:"Inclusive range start as epoch milliseconds."`
	To            int64  `query:"to" doc:"Exclusive range end as epoch milliseconds."`
	Metric        string `query:"metric" doc:"Ping metric to aggregate: rttAvgMs, lossPercent, or successRate." example:"rttAvgMs"`
	MaxDataPoints int32  `query:"maxDataPoints" doc:"Maximum target points per series." example:"600"`
}

type queryPingSeriesOutput struct {
	Body queryPingSeriesBody
}

type queryPingSeriesBody struct {
	Series []seriesBody      `json:"series"`
	Query  queryMetadataBody `json:"query"`
}

type seriesBody struct {
	Name   string            `json:"name" example:"rttAvgMs"`
	Labels map[string]string `json:"labels"`
	Unit   string            `json:"unit" example:"ms"`
	Points []pointTuple      `json:"points" doc:"Tuple points in the shape [epochMilliseconds, value]."`
}

type pointTuple [2]float64

type queryMetadataBody struct {
	FromMs        int64  `json:"from" example:"1778662800000"`
	ToMs          int64  `json:"to" example:"1778749200000"`
	MaxDataPoints int32  `json:"maxDataPoints" example:"600"`
	Resolution    string `json:"resolution" example:"lttb"`
	TotalPoints   int64  `json:"totalPoints" example:"487"`
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
