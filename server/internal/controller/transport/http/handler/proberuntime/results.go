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
	ProbeID string
	Body    submitResultsBody
}

type submitResultsBody struct {
	Results []runtimeResultGroupBody `json:"results"`
}

type runtimeResultGroupBody struct {
	CheckID    string                        `json:"checkId"`
	Type       string                        `json:"type"`
	Ping       []pingResultBody              `json:"ping,omitempty"`
	TCP        []runtimeTCPResultBody        `json:"tcp,omitempty"`
	Traceroute []runtimeTracerouteResultBody `json:"traceroute,omitempty"`
	HTTP       []runtimeHTTPResultBody       `json:"http,omitempty"`
}

type pingResultBody struct {
	StartedAt     time.Time   `json:"startedAt"`
	FinishedAt    time.Time   `json:"finishedAt"`
	DurationMs    int32       `json:"durationMs"`
	Status        string      `json:"status"`
	SentCount     int32       `json:"sentCount"`
	ReceivedCount int32       `json:"receivedCount"`
	LossPercent   float64     `json:"lossPercent"`
	RttMinMs      *float64    `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64    `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64    `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64    `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64    `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64   `json:"rttSamplesMs,omitempty"`
	ResolvedIP    *netip.Addr `json:"resolvedIp,omitempty"`
	IPFamily      *string     `json:"ipFamily,omitempty"`
	ErrorCode     *string     `json:"errorCode,omitempty"`
	ErrorMessage  *string     `json:"errorMessage,omitempty"`
}

type runtimeTCPResultBody struct {
	StartedAt         time.Time   `json:"startedAt"`
	FinishedAt        time.Time   `json:"finishedAt"`
	DurationMs        int32       `json:"durationMs"`
	Status            string      `json:"status"`
	ConnectDurationMs *float64    `json:"connectDurationMs,omitempty"`
	ResolvedIP        *netip.Addr `json:"resolvedIp,omitempty"`
	IPFamily          *string     `json:"ipFamily,omitempty"`
	ErrorCode         *string     `json:"errorCode,omitempty"`
	ErrorMessage      *string     `json:"errorMessage,omitempty"`
}

type runtimeHTTPResultBody struct {
	StartedAt            time.Time   `json:"startedAt"`
	FinishedAt           time.Time   `json:"finishedAt"`
	DurationMs           int32       `json:"durationMs"`
	Status               string      `json:"status"`
	DNSDurationMs        *float64    `json:"dnsDurationMs,omitempty"`
	ConnectDurationMs    *float64    `json:"connectDurationMs,omitempty"`
	TLSDurationMs        *float64    `json:"tlsDurationMs,omitempty"`
	TTFBDurationMs       *float64    `json:"ttfbDurationMs,omitempty"`
	ResolvedIP           *netip.Addr `json:"resolvedIp,omitempty"`
	IPFamily             *string     `json:"ipFamily,omitempty"`
	StatusCode           *int32      `json:"statusCode,omitempty"`
	FinalURL             *string     `json:"finalUrl,omitempty"`
	RedirectCount        int32       `json:"redirectCount"`
	ResponseBytes        *int64      `json:"responseBytes,omitempty"`
	ResponseTruncated    bool        `json:"responseTruncated"`
	BodyMatched          *bool       `json:"bodyMatched,omitempty"`
	TLSVersion           *string     `json:"tlsVersion,omitempty"`
	TLSCipherSuite       *string     `json:"tlsCipherSuite,omitempty"`
	CertificateNotBefore *time.Time  `json:"certificateNotBefore,omitempty"`
	CertificateNotAfter  *time.Time  `json:"certificateNotAfter,omitempty"`
	ErrorCode            *string     `json:"errorCode,omitempty"`
	ErrorMessage         *string     `json:"errorMessage,omitempty"`
}

type runtimeTracerouteResultBody struct {
	StartedAt          time.Time                  `json:"startedAt"`
	FinishedAt         time.Time                  `json:"finishedAt"`
	DurationMs         int32                      `json:"durationMs"`
	Status             string                     `json:"status"`
	ResolvedIP         *netip.Addr                `json:"resolvedIp,omitempty"`
	IPFamily           *string                    `json:"ipFamily,omitempty"`
	DestinationReached bool                       `json:"destinationReached"`
	HopCount           int32                      `json:"hopCount"`
	Hops               []runtimeTracerouteHopBody `json:"hops,omitempty"`
	ErrorCode          *string                    `json:"errorCode,omitempty"`
	ErrorMessage       *string                    `json:"errorMessage,omitempty"`
}

type runtimeTracerouteHopBody struct {
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
			TCP:        newTCPResultInputs(group.TCP),
			Traceroute: newTracerouteResultInputs(group.Traceroute),
			HTTP:       newHTTPResultInputs(group.HTTP),
		})
	}

	return appproberuntime.SubmitResultsInput{
		RuntimeAuthInput: auth,
		Results:          results,
	}
}

func newHTTPResultInputs(values []runtimeHTTPResultBody) []appproberuntime.HTTPResultInput {
	results := make([]appproberuntime.HTTPResultInput, 0, len(values))
	for _, value := range values {
		results = append(results, appproberuntime.HTTPResultInput{
			StartedAt: value.StartedAt, FinishedAt: value.FinishedAt, DurationMs: value.DurationMs,
			Status: value.Status, DNSDurationMs: value.DNSDurationMs, ConnectDurationMs: value.ConnectDurationMs,
			TLSDurationMs: value.TLSDurationMs, TTFBDurationMs: value.TTFBDurationMs,
			ResolvedIP: cloneAddr(value.ResolvedIP), IPFamily: value.IPFamily, StatusCode: value.StatusCode,
			FinalURL: value.FinalURL, RedirectCount: value.RedirectCount, ResponseBytes: value.ResponseBytes,
			ResponseTruncated: value.ResponseTruncated, BodyMatched: value.BodyMatched,
			TLSVersion: value.TLSVersion, TLSCipherSuite: value.TLSCipherSuite,
			CertificateNotBefore: value.CertificateNotBefore, CertificateNotAfter: value.CertificateNotAfter,
			ErrorCode: value.ErrorCode, ErrorMessage: value.ErrorMessage,
		})
	}
	return results
}

func newTCPResultInputs(values []runtimeTCPResultBody) []appproberuntime.TCPResultInput {
	results := make([]appproberuntime.TCPResultInput, 0, len(values))
	for _, value := range values {
		results = append(results, appproberuntime.TCPResultInput{
			StartedAt:         value.StartedAt,
			FinishedAt:        value.FinishedAt,
			DurationMs:        value.DurationMs,
			Status:            value.Status,
			ConnectDurationMs: value.ConnectDurationMs,
			ResolvedIP:        cloneAddr(value.ResolvedIP),
			IPFamily:          value.IPFamily,
			ErrorCode:         value.ErrorCode,
			ErrorMessage:      value.ErrorMessage,
		})
	}

	return results
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
