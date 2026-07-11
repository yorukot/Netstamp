package result

import (
	"context"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

type queryMetadataBody struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Source        string `json:"source,omitempty"`
	Resolution    string `json:"resolution"`
	TotalPoints   int64  `json:"totalPoints"`
}

func newQueryMetadataBody(meta appresult.QueryMetadata) queryMetadataBody {
	return queryMetadataBody{
		FromMs:        meta.FromMs,
		ToMs:          meta.ToMs,
		MaxDataPoints: meta.MaxDataPoints,
		Source:        meta.Source,
		Resolution:    meta.Resolution,
		TotalPoints:   meta.TotalPoints,
	}
}

type seriesBody struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Unit   string            `json:"unit"`
	Points []pointTuple      `json:"points"`
}

type pointTuple [2]float64

type seriesBodySource[P any] struct {
	Name   string
	Labels seriesLabelsSource
	Unit   string
	Points []P
}

type seriesLabelsSource struct {
	ProbeID   string
	CheckID   string
	CheckType string
}

func newSeriesBodySource[P any](name, probeID, checkID, checkType, unit string, points []P) seriesBodySource[P] {
	return seriesBodySource[P]{
		Name: name,
		Labels: seriesLabelsSource{
			ProbeID:   probeID,
			CheckID:   checkID,
			CheckType: checkType,
		},
		Unit:   unit,
		Points: points,
	}
}

func newSeriesBodyMap[S, P any](
	values map[string]S,
	source func(S) seriesBodySource[P],
	pointValues func(P) (int64, float64),
) map[string]seriesBody {
	series := make(map[string]seriesBody, len(values))
	for name, value := range values {
		seriesSource := source(value)
		points := make([]pointTuple, 0, len(seriesSource.Points))
		for _, point := range seriesSource.Points {
			timestampMs, value := pointValues(point)
			points = append(points, pointTuple{float64(timestampMs), value})
		}
		series[name] = seriesBody{
			Name: seriesSource.Name,
			Labels: map[string]string{
				"probeId":   seriesSource.Labels.ProbeID,
				"checkId":   seriesSource.Labels.CheckID,
				"checkType": seriesSource.Labels.CheckType,
			},
			Unit:   seriesSource.Unit,
			Points: points,
		}
	}
	return series
}

type queryInsightInput struct {
	Ref           string
	ProbeID       string
	CheckID       string
	From          int64
	To            int64
	MaxDataPoints int32
}

type (
	queryPingInsightInput = queryInsightInput
	queryTCPInsightInput  = queryInsightInput
	queryHTTPInsightInput = queryInsightInput
)

type queryInsightOutput[B any] struct {
	Body B
}

type queryInsightBody[S any] struct {
	Summary S                 `json:"summary"`
	Meta    queryMetadataBody `json:"meta"`
}

type (
	queryPingInsightOutput = queryInsightOutput[queryPingInsightBody]
	queryTCPInsightOutput  = queryInsightOutput[queryTCPInsightBody]
	queryHTTPInsightOutput = queryInsightOutput[queryHTTPInsightBody]
	queryPingInsightBody   = queryInsightBody[pingInsightSummaryBody]
	queryTCPInsightBody    = queryInsightBody[tcpInsightSummaryBody]
	queryHTTPInsightBody   = queryInsightBody[httpInsightSummaryBody]
)

func newQueryInsightBody[S any](summary S, meta appresult.QueryMetadata) queryInsightBody[S] {
	return queryInsightBody[S]{
		Summary: summary,
		Meta:    newQueryMetadataBody(meta),
	}
}

func (h *Handler) queryPingInsight(ctx context.Context, input *queryPingInsightInput) (*queryPingInsightOutput, error) {
	return queryInsight(ctx, input, newQueryInsightServiceInput[appresult.QueryPingInsightInput], h.service.QueryPingInsight, newQueryPingInsightBody, "query ping insight failed")
}

func (h *Handler) queryTCPInsight(ctx context.Context, input *queryTCPInsightInput) (*queryTCPInsightOutput, error) {
	return queryInsight(ctx, input, newQueryInsightServiceInput[appresult.QueryTCPInsightInput], h.service.QueryTCPInsight, newQueryTCPInsightBody, "query tcp insight failed")
}

func (h *Handler) queryHTTPInsight(ctx context.Context, input *queryHTTPInsightInput) (*queryHTTPInsightOutput, error) {
	return queryInsight(ctx, input, newQueryInsightServiceInput[appresult.QueryHTTPInsightInput], h.service.QueryHTTPInsight, newQueryHTTPInsightBody, "query http insight failed")
}

type queryInsightServiceInput interface {
	~struct {
		CurrentUserID string
		ProjectRef    string
		ProbeID       string
		CheckID       string
		FromMs        *int64
		ToMs          *int64
		MaxDataPoints *int32
		Now           time.Time
	}
}

func newQueryInsightServiceInput[I queryInsightServiceInput](userID string, input *queryInsightInput) I {
	return I{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		MaxDataPoints: optionalInt32(input.MaxDataPoints),
	}
}

func queryInsight[I, O, B any](
	ctx context.Context,
	input *queryInsightInput,
	newInput func(string, *queryInsightInput) I,
	query func(context.Context, I) (O, error),
	newBody func(O) B,
	fallback string,
) (*queryInsightOutput[B], error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := query(ctx, newInput(userID, input))
	if err != nil {
		return nil, mapResultError(err, fallback)
	}

	return &queryInsightOutput[B]{Body: newBody(output)}, nil
}
