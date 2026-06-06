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
		Series:        input.Series,
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
	Series        string
	MaxDataPoints int32
}

type queryPingSeriesOutput struct {
	Body queryPingSeriesBody
}

type queryPingSeriesBody struct {
	Series map[string]seriesBody `json:"series"`
	Meta   queryMetadataBody     `json:"meta"`
}

func newQueryPingSeriesBody(output appresult.PingSeriesOutput) queryPingSeriesBody {
	return queryPingSeriesBody{
		Series: newSeriesBodyMap(output.Series, newPingSeriesBodySource, pingSeriesPointValues),
		Meta:   newQueryMetadataBody(output.Meta),
	}
}

func newPingSeriesBodySource(value appresult.PingSeries) seriesBodySource[appresult.PingSeriesPoint] {
	return newSeriesBodySource(value.Name, value.Labels.ProbeID, value.Labels.CheckID, value.Labels.CheckType, value.Unit, value.Points)
}

func pingSeriesPointValues(point appresult.PingSeriesPoint) (int64, float64) {
	return point.TimestampMs, point.Value
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
