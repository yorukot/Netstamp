package proberuntime

import (
	"context"
	"net/netip"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
)

func (h *Handler) submitResults(ctx context.Context, input *submitResultsInput) (*submitResultsOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.SubmitResults(ctx, newSubmitResultsInput(auth, input.Body))
	if err != nil {
		return nil, mapRuntimeError(err, "submit probe runtime results failed")
	}

	return &submitResultsOutput{Body: submitResultsOutputBody{
		Accepted:   output.Accepted,
		ServerTime: output.ServerTime,
	}}, nil
}

type submitResultsInput struct {
	ProbeID string `path:"probe_id" doc:"Probe ID."`
	Body    submitResultsBody
}

type submitResultsBody struct {
	Results []runtimeResultGroupBody `json:"results" minItems:"1" maxItems:"100" doc:"Result groups keyed by check and check type."`
}

type runtimeResultGroupBody struct {
	CheckID    string                        `json:"checkId" format:"uuid" doc:"Assigned check ID." example:"44444444-4444-4444-4444-444444444444"`
	Type       string                        `json:"type" enum:"ping,traceroute" doc:"Check result type. Must match the assigned check type." example:"ping"`
	Ping       []pingResultBody              `json:"ping,omitempty" doc:"Ping result payloads for this check."`
	Traceroute []runtimeTracerouteResultBody `json:"traceroute,omitempty" doc:"Traceroute result payloads for this check."`
}

type pingResultBody struct {
	StartedAt     time.Time   `json:"startedAt" doc:"UTC time when the ping check started." example:"2026-05-13T10:00:00Z"`
	FinishedAt    time.Time   `json:"finishedAt" doc:"UTC time when the ping check finished." example:"2026-05-13T10:00:01Z"`
	DurationMs    int32       `json:"durationMs" minimum:"0" doc:"Total check duration in milliseconds." example:"1000"`
	Status        string      `json:"status" enum:"successful,timeout,error" doc:"Ping result status." example:"successful"`
	SentCount     int32       `json:"sentCount" minimum:"0" doc:"Packets sent." example:"4"`
	ReceivedCount int32       `json:"receivedCount" minimum:"0" doc:"Packets received." example:"4"`
	LossPercent   float64     `json:"lossPercent" minimum:"0" maximum:"100" doc:"Packet loss percentage." example:"0"`
	RttMinMs      *float64    `json:"rttMinMs,omitempty" minimum:"0" doc:"Minimum RTT in milliseconds." example:"10.1"`
	RttAvgMs      *float64    `json:"rttAvgMs,omitempty" minimum:"0" doc:"Average RTT in milliseconds." example:"12.3"`
	RttMedianMs   *float64    `json:"rttMedianMs,omitempty" minimum:"0" doc:"Median RTT in milliseconds." example:"12"`
	RttMaxMs      *float64    `json:"rttMaxMs,omitempty" minimum:"0" doc:"Maximum RTT in milliseconds." example:"15.6"`
	RttStddevMs   *float64    `json:"rttStddevMs,omitempty" minimum:"0" doc:"RTT standard deviation in milliseconds." example:"1.7"`
	RttSamplesMs  []float64   `json:"rttSamplesMs,omitempty" doc:"RTT sample values in milliseconds."`
	ResolvedIP    *netip.Addr `json:"resolvedIp,omitempty" doc:"Resolved IP address used for the ping." example:"1.1.1.1"`
	IPFamily      *string     `json:"ipFamily,omitempty" enum:"inet,inet6" doc:"IP family used for the check." example:"inet"`
	ErrorCode     *string     `json:"errorCode,omitempty" doc:"Optional machine-readable error code." example:"icmp_timeout"`
	ErrorMessage  *string     `json:"errorMessage,omitempty" doc:"Optional executor error message." example:"request timed out"`
}

type runtimeTracerouteResultBody struct {
	StartedAt          time.Time                  `json:"startedAt" doc:"UTC time when the traceroute check started." example:"2026-05-13T10:00:00Z"`
	FinishedAt         time.Time                  `json:"finishedAt" doc:"UTC time when the traceroute check finished." example:"2026-05-13T10:00:04Z"`
	DurationMs         int32                      `json:"durationMs" minimum:"0" doc:"Total check duration in milliseconds." example:"4000"`
	Status             string                     `json:"status" enum:"successful,timeout,error,partial" doc:"Traceroute result status." example:"partial"`
	ResolvedIP         *netip.Addr                `json:"resolvedIp,omitempty" doc:"Resolved destination IP address." example:"93.184.216.34"`
	IPFamily           *string                    `json:"ipFamily,omitempty" enum:"inet,inet6" doc:"IP family used for the check." example:"inet"`
	DestinationReached bool                       `json:"destinationReached" doc:"Whether the traceroute reached the destination." example:"false"`
	HopCount           int32                      `json:"hopCount" minimum:"0" doc:"Observed hop count." example:"12"`
	Hops               []runtimeTracerouteHopBody `json:"hops,omitempty" doc:"Per-hop traceroute samples."`
	ErrorCode          *string                    `json:"errorCode,omitempty" doc:"Optional machine-readable error code." example:"destination_unreached"`
	ErrorMessage       *string                    `json:"errorMessage,omitempty" doc:"Optional executor error message." example:"destination was not reached before max hops"`
}

type runtimeTracerouteHopBody struct {
	HopIndex      int32       `json:"hopIndex" minimum:"1" maximum:"64" doc:"One-based hop index." example:"1"`
	Address       *netip.Addr `json:"address,omitempty" doc:"Hop IP address." example:"192.0.2.1"`
	Hostname      *string     `json:"hostname,omitempty" doc:"Hop reverse DNS hostname." example:"gateway.local"`
	SentCount     int32       `json:"sentCount" minimum:"0" doc:"Probe attempts sent for this hop." example:"3"`
	ReceivedCount int32       `json:"receivedCount" minimum:"0" doc:"Probe attempts received for this hop." example:"3"`
	LossPercent   float64     `json:"lossPercent" minimum:"0" maximum:"100" doc:"Probe loss percentage for this hop." example:"0"`
	RttMinMs      *float64    `json:"rttMinMs,omitempty" minimum:"0" doc:"Minimum RTT in milliseconds." example:"1.5"`
	RttAvgMs      *float64    `json:"rttAvgMs,omitempty" minimum:"0" doc:"Average RTT in milliseconds." example:"1.7"`
	RttMedianMs   *float64    `json:"rttMedianMs,omitempty" minimum:"0" doc:"Median RTT in milliseconds." example:"1.7"`
	RttMaxMs      *float64    `json:"rttMaxMs,omitempty" minimum:"0" doc:"Maximum RTT in milliseconds." example:"1.9"`
	RttStddevMs   *float64    `json:"rttStddevMs,omitempty" minimum:"0" doc:"RTT standard deviation in milliseconds." example:"0.2"`
	RttSamplesMs  []float64   `json:"rttSamplesMs,omitempty" doc:"RTT sample values in milliseconds."`
	ErrorCode     *string     `json:"errorCode,omitempty" doc:"Optional machine-readable hop error code." example:"hop_timeout"`
	ErrorMessage  *string     `json:"errorMessage,omitempty" doc:"Optional hop error message." example:"request timed out"`
}

type submitResultsOutput struct {
	Body submitResultsOutputBody
}

type submitResultsOutputBody struct {
	Accepted   int       `json:"accepted"`
	ServerTime time.Time `json:"serverTime"`
}

func newSubmitResultsInput(auth appproberuntime.RuntimeAuthInput, body submitResultsBody) appproberuntime.SubmitResultsInput {
	results := make([]appproberuntime.RuntimeResultGroupInput, 0, len(body.Results))
	for _, group := range body.Results {
		results = append(results, appproberuntime.RuntimeResultGroupInput{
			CheckID:    group.CheckID,
			Type:       group.Type,
			Ping:       newPingResultInputs(group.Ping),
			Traceroute: newTracerouteResultInputs(group.Traceroute),
		})
	}

	return appproberuntime.SubmitResultsInput{
		RuntimeAuthInput: auth,
		Results:          results,
	}
}

func newTracerouteResultInputs(values []runtimeTracerouteResultBody) []appproberuntime.TracerouteResultInput {
	results := make([]appproberuntime.TracerouteResultInput, 0, len(values))
	for _, value := range values {
		results = append(results, appproberuntime.TracerouteResultInput{
			StartedAt:          value.StartedAt,
			FinishedAt:         value.FinishedAt,
			DurationMs:         value.DurationMs,
			Status:             value.Status,
			ResolvedIP:         cloneAddr(value.ResolvedIP),
			IPFamily:           value.IPFamily,
			DestinationReached: value.DestinationReached,
			HopCount:           value.HopCount,
			Hops:               newTracerouteHopInputs(value.Hops),
			ErrorCode:          value.ErrorCode,
			ErrorMessage:       value.ErrorMessage,
		})
	}

	return results
}

func newTracerouteHopInputs(values []runtimeTracerouteHopBody) []appproberuntime.TracerouteHopInput {
	hops := make([]appproberuntime.TracerouteHopInput, 0, len(values))
	for _, value := range values {
		hops = append(hops, appproberuntime.TracerouteHopInput{
			HopIndex:      value.HopIndex,
			Address:       cloneAddr(value.Address),
			Hostname:      value.Hostname,
			SentCount:     value.SentCount,
			ReceivedCount: value.ReceivedCount,
			LossPercent:   value.LossPercent,
			RttMinMs:      value.RttMinMs,
			RttAvgMs:      value.RttAvgMs,
			RttMedianMs:   value.RttMedianMs,
			RttMaxMs:      value.RttMaxMs,
			RttStddevMs:   value.RttStddevMs,
			RttSamplesMs:  append([]float64(nil), value.RttSamplesMs...),
			ErrorCode:     value.ErrorCode,
			ErrorMessage:  value.ErrorMessage,
		})
	}

	return hops
}

func newPingResultInputs(values []pingResultBody) []appproberuntime.PingResultInput {
	results := make([]appproberuntime.PingResultInput, 0, len(values))
	for _, value := range values {
		results = append(results, appproberuntime.PingResultInput{
			StartedAt:     value.StartedAt,
			FinishedAt:    value.FinishedAt,
			DurationMs:    value.DurationMs,
			Status:        value.Status,
			SentCount:     value.SentCount,
			ReceivedCount: value.ReceivedCount,
			LossPercent:   value.LossPercent,
			RttMinMs:      value.RttMinMs,
			RttAvgMs:      value.RttAvgMs,
			RttMedianMs:   value.RttMedianMs,
			RttMaxMs:      value.RttMaxMs,
			RttStddevMs:   value.RttStddevMs,
			RttSamplesMs:  append([]float64(nil), value.RttSamplesMs...),
			ResolvedIP:    cloneAddr(value.ResolvedIP),
			IPFamily:      value.IPFamily,
			ErrorCode:     value.ErrorCode,
			ErrorMessage:  value.ErrorMessage,
		})
	}

	return results
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}
