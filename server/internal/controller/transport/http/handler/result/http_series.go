package result

import (
	"context"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

type queryHTTPSeriesInput struct {
	Ref, ProbeID, CheckID string
	From, To              int64
	Series                string
	MaxDataPoints         int32
}
type (
	queryHTTPSeriesOutput struct{ Body queryHTTPSeriesBody }
	queryHTTPSeriesBody   struct {
		Series map[string]seriesBody `json:"series"`
		Meta   queryMetadataBody     `json:"meta"`
	}
)

func (h *Handler) queryHTTPSeries(ctx context.Context, input *queryHTTPSeriesInput) (*queryHTTPSeriesOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	output, err := h.service.QueryHTTPSeries(ctx, appresult.QueryHTTPSeriesInput{CurrentUserID: userID, ProjectRef: input.Ref, ProbeID: input.ProbeID, CheckID: input.CheckID, FromMs: optionalInt64(input.From), ToMs: optionalInt64(input.To), Series: input.Series, MaxDataPoints: optionalInt32(input.MaxDataPoints)})
	if err != nil {
		return nil, mapResultError(err, "query http result series failed")
	}
	return &queryHTTPSeriesOutput{Body: queryHTTPSeriesBody{Series: newSeriesBodyMap(output.Series, func(value appresult.HTTPSeries) seriesBodySource[appresult.HTTPSeriesPoint] {
		return newSeriesBodySource(value.Name, value.Labels.ProbeID, value.Labels.CheckID, value.Labels.CheckType, value.Unit, value.Points)
	}, func(point appresult.HTTPSeriesPoint) (int64, float64) { return point.TimestampMs, point.Value }), Meta: newQueryMetadataBody(output.Meta)}}, nil
}
