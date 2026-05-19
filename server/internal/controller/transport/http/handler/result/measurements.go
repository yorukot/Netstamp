package result

import (
	"context"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryMeasurements(ctx context.Context, input *queryMeasurementsInput) (*queryMeasurementsOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryMeasurements(ctx, appresult.QueryMeasurementsInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		Type:          input.Type,
		Status:        input.Status,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		Limit:         optionalInt32(input.Limit),
		CursorMs:      optionalInt64(input.Cursor),
	})
	if err != nil {
		return nil, mapResultError(err, "query measurements failed")
	}

	return &queryMeasurementsOutput{Body: newQueryMeasurementsBody(output)}, nil
}

type queryMeasurementsInput struct {
	Ref     string
	ProbeID string
	CheckID string
	Type    string
	Status  string
	From    int64
	To      int64
	Limit   int32
	Cursor  int64
}

type queryMeasurementsOutput struct {
	Body queryMeasurementsBody
}

type queryMeasurementsBody struct {
	Measurements []measurementBody            `json:"measurements"`
	Query        measurementQueryMetadataBody `json:"query"`
}

type measurementBody struct {
	Type         string    `json:"type"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
	ProbeID      string    `json:"probeId"`
	CheckID      string    `json:"checkId"`
	Status       string    `json:"status"`
	DurationMs   int32     `json:"durationMs"`
	LatencyMs    *float64  `json:"latencyMs,omitempty"`
	LossPercent  *float64  `json:"lossPercent,omitempty"`
	Metadata     *string   `json:"metadata,omitempty"`
	ErrorCode    *string   `json:"errorCode,omitempty"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
}

type measurementQueryMetadataBody struct {
	FromMs     int64  `json:"from"`
	ToMs       int64  `json:"to"`
	Limit      int32  `json:"limit"`
	NextCursor *int64 `json:"nextCursor,omitempty"`
}

func newQueryMeasurementsBody(output appresult.MeasurementsOutput) queryMeasurementsBody {
	measurements := make([]measurementBody, 0, len(output.Measurements))
	for _, measurement := range output.Measurements {
		measurements = append(measurements, measurementBody{
			Type:         measurement.Type,
			StartedAt:    measurement.StartedAt,
			FinishedAt:   measurement.FinishedAt,
			ProbeID:      measurement.ProbeID,
			CheckID:      measurement.CheckID,
			Status:       measurement.Status,
			DurationMs:   measurement.DurationMs,
			LatencyMs:    measurement.LatencyMs,
			LossPercent:  measurement.LossPercent,
			Metadata:     measurement.Metadata,
			ErrorCode:    measurement.ErrorCode,
			ErrorMessage: measurement.ErrorMessage,
		})
	}

	return queryMeasurementsBody{
		Measurements: measurements,
		Query: measurementQueryMetadataBody{
			FromMs:     output.Query.FromMs,
			ToMs:       output.Query.ToMs,
			Limit:      output.Query.Limit,
			NextCursor: output.Query.NextCursor,
		},
	}
}
