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
	return queryTCPSeriesBody{
		Series: newSeriesBodyMap(output.Series, newTCPSeriesBodySource, tcpSeriesPointValues),
		Meta:   newQueryMetadataBody(output.Meta),
	}
}

func newTCPSeriesBodySource(value appresult.TCPSeries) seriesBodySource[appresult.TCPSeriesPoint] {
	return newSeriesBodySource(value.Name, value.Labels.ProbeID, value.Labels.CheckID, value.Labels.CheckType, value.Unit, value.Points)
}

func tcpSeriesPointValues(point appresult.TCPSeriesPoint) (int64, float64) {
	return point.TimestampMs, point.Value
}
