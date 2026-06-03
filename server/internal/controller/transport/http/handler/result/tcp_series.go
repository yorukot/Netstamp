package result

import (
	"context"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryTCPSeries(ctx context.Context, input *queryTCPSeriesInput) (*queryTCPSeriesOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryTCPSeries(ctx, appresult.QueryTCPSeriesInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		Series:        input.Series,
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	})
	if err != nil {
		return nil, mapResultError(err, "query tcp result series failed")
	}

	return &queryTCPSeriesOutput{Body: newQueryTCPSeriesBody(output)}, nil
}

type queryTCPSeriesInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	Series        string
	MaxDataPoints int32
}

type queryTCPSeriesOutput struct {
	Body queryTCPSeriesBody
}

type queryTCPSeriesBody struct {
	Series map[string]seriesBody `json:"series"`
	Meta   queryMetadataBody     `json:"meta"`
}

func newQueryTCPSeriesBody(output appresult.TCPSeriesOutput) queryTCPSeriesBody {
	series := make(map[string]seriesBody, len(output.Series))
	for name, value := range output.Series {
		points := make([]pointTuple, 0, len(value.Points))
		for _, point := range value.Points {
			points = append(points, pointTuple{float64(point.TimestampMs), point.Value})
		}
		series[name] = seriesBody{
			Name: value.Name,
			Labels: map[string]string{
				"probeId":   value.Labels.ProbeID,
				"checkId":   value.Labels.CheckID,
				"checkType": value.Labels.CheckType,
			},
			Unit:   value.Unit,
			Points: points,
		}
	}

	return queryTCPSeriesBody{
		Series: series,
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
