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
	Ref     string `path:"ref" doc:"Project ID or slug." example:"vector-ix"`
	ProbeID string `query:"probeId" required:"true" format:"uuid" doc:"Probe ID to query."`
	CheckID string `query:"checkId" required:"true" format:"uuid" doc:"Check ID to query."`
	From    int64  `query:"from" doc:"Inclusive range start as epoch milliseconds."`
	To      int64  `query:"to" doc:"Exclusive range end as epoch milliseconds."`
	Limit   int32  `query:"limit" doc:"Maximum traceroute runs to return." example:"100"`
	Cursor  int64  `query:"cursor" doc:"Pagination cursor as a run startedAt epoch millisecond timestamp."`
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
	FromMs     int64  `json:"from" example:"1778662800000"`
	ToMs       int64  `json:"to" example:"1778749200000"`
	Limit      int32  `json:"limit" example:"100"`
	NextCursor *int64 `json:"nextCursor,omitempty" example:"1778662800000"`
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
