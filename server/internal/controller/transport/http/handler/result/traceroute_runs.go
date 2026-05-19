package result

import (
	"context"
	"net/netip"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryTracerouteRuns(ctx context.Context, input *queryTracerouteRunsInput) (*queryTracerouteRunsOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryTracerouteRuns(ctx, appresult.QueryTracerouteRunsInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		Limit:         optionalInt32(input.Limit),
		CursorMs:      optionalInt64(input.Cursor),
	})
	if err != nil {
		return nil, mapResultError(err, "query traceroute result runs failed")
	}

	return &queryTracerouteRunsOutput{Body: newQueryTracerouteRunsBody(output)}, nil
}

type queryTracerouteRunsInput struct {
	Ref     string
	ProbeID string
	CheckID string
	From    int64
	To      int64
	Limit   int32
	Cursor  int64
}

type queryTracerouteRunsOutput struct {
	Body queryTracerouteRunsBody
}

type queryTracerouteRunsBody struct {
	Runs  []tracerouteResultRunBody      `json:"runs"`
	Query tracerouteRunQueryMetadataBody `json:"query"`
}

type tracerouteResultRunBody struct {
	StartedAt          time.Time                 `json:"startedAt"`
	FinishedAt         time.Time                 `json:"finishedAt"`
	DurationMs         int32                     `json:"durationMs"`
	Status             string                    `json:"status"`
	ResolvedIP         *netip.Addr               `json:"resolvedIp,omitempty"`
	IPFamily           *string                   `json:"ipFamily,omitempty"`
	DestinationReached bool                      `json:"destinationReached"`
	HopCount           int32                     `json:"hopCount"`
	ErrorCode          *string                   `json:"errorCode,omitempty"`
	ErrorMessage       *string                   `json:"errorMessage,omitempty"`
	Hops               []tracerouteResultHopBody `json:"hops"`
}

type tracerouteResultHopBody struct {
	HopIndex      int32       `json:"hopIndex"`
	Address       *netip.Addr `json:"address,omitempty"`
	Hostname      *string     `json:"hostname,omitempty"`
	SentCount     int32       `json:"sentCount"`
	ReceivedCount int32       `json:"receivedCount"`
	LossPercent   float64     `json:"lossPercent"`
	RttMinMs      *float64    `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64    `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64    `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64    `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64    `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64   `json:"rttSamplesMs,omitempty"`
	ErrorCode     *string     `json:"errorCode,omitempty"`
	ErrorMessage  *string     `json:"errorMessage,omitempty"`
}

type tracerouteRunQueryMetadataBody struct {
	FromMs     int64  `json:"from"`
	ToMs       int64  `json:"to"`
	Limit      int32  `json:"limit"`
	NextCursor *int64 `json:"nextCursor,omitempty"`
}

func newQueryTracerouteRunsBody(output appresult.TracerouteRunsOutput) queryTracerouteRunsBody {
	runs := make([]tracerouteResultRunBody, 0, len(output.Runs))
	for _, run := range output.Runs {
		runs = append(runs, tracerouteResultRunBody{
			StartedAt:          run.StartedAt,
			FinishedAt:         run.FinishedAt,
			DurationMs:         run.DurationMs,
			Status:             run.Status,
			ResolvedIP:         run.ResolvedIP,
			IPFamily:           run.IPFamily,
			DestinationReached: run.DestinationReached,
			HopCount:           run.HopCount,
			ErrorCode:          run.ErrorCode,
			ErrorMessage:       run.ErrorMessage,
			Hops:               newTracerouteHopBodies(run.Hops),
		})
	}

	return queryTracerouteRunsBody{
		Runs: runs,
		Query: tracerouteRunQueryMetadataBody{
			FromMs:     output.Query.FromMs,
			ToMs:       output.Query.ToMs,
			Limit:      output.Query.Limit,
			NextCursor: output.Query.NextCursor,
		},
	}
}

func newTracerouteHopBodies(hops []appresult.TracerouteHop) []tracerouteResultHopBody {
	values := make([]tracerouteResultHopBody, 0, len(hops))
	for _, hop := range hops {
		values = append(values, tracerouteResultHopBody{
			HopIndex:      hop.HopIndex,
			Address:       hop.Address,
			Hostname:      hop.Hostname,
			SentCount:     hop.SentCount,
			ReceivedCount: hop.ReceivedCount,
			LossPercent:   hop.LossPercent,
			RttMinMs:      hop.RttMinMs,
			RttAvgMs:      hop.RttAvgMs,
			RttMedianMs:   hop.RttMedianMs,
			RttMaxMs:      hop.RttMaxMs,
			RttStddevMs:   hop.RttStddevMs,
			RttSamplesMs:  append([]float64(nil), hop.RttSamplesMs...),
			ErrorCode:     hop.ErrorCode,
			ErrorMessage:  hop.ErrorMessage,
		})
	}
	return values
}
