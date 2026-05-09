package proberuntime

import (
	"context"
	"encoding/json"
	"net/netip"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func (h *Handler) submitResults(ctx context.Context, input *submitResultsInput) (*submitResultsOutput, error) {
	auth, err := runtimeAuthInput(input.ProbeID, input.Authorization)
	if err != nil {
		return nil, err
	}
	pingResults, err := newPingResultInputs(input.Body.Ping)
	if err != nil {
		return nil, err
	}
	dnsResults, err := newUnsupportedResultInputs(input.Body.DNS)
	if err != nil {
		return nil, err
	}
	tracerouteResults, err := newUnsupportedResultInputs(input.Body.Traceroute)
	if err != nil {
		return nil, err
	}
	err = h.service.SubmitResults(ctx, appproberuntime.SubmitResultsInput{
		RuntimeAuthInput: auth,
		Ping:             pingResults,
		DNS:              dnsResults,
		Traceroute:       tracerouteResults,
	})
	if err != nil {
		return nil, mapRuntimeError(err, "submit probe results failed")
	}

	return &submitResultsOutput{}, nil
}

type submitResultsInput struct {
	ProbeID       string `path:"probe_id" format:"uuid" doc:"Probe ID."`
	Authorization string `header:"Authorization" doc:"Probe authorization header. Use 'Probe <secret>'."`
	Body          submitResultsInputBody
}

type submitResultsInputBody struct {
	Ping       []pingResultInputBody `json:"ping,omitempty" maxItems:"500"`
	DNS        []map[string]any      `json:"dns,omitempty" maxItems:"500"`
	Traceroute []map[string]any      `json:"traceroute,omitempty" maxItems:"500"`
}

type pingResultInputBody struct {
	CheckID       string         `json:"checkId" format:"uuid" required:"true"`
	StartedAt     time.Time      `json:"startedAt" required:"true"`
	FinishedAt    time.Time      `json:"finishedAt" required:"true"`
	DurationMs    int32          `json:"durationMs" minimum:"0" required:"true"`
	Status        string         `json:"status" enum:"successful,timeout,error" required:"true"`
	SentCount     int32          `json:"sentCount" minimum:"0" required:"true"`
	ReceivedCount int32          `json:"receivedCount" minimum:"0" required:"true"`
	LossPercent   float64        `json:"lossPercent" minimum:"0" maximum:"100" required:"true"`
	RttMinMs      *float64       `json:"rttMinMs,omitempty" minimum:"0"`
	RttAvgMs      *float64       `json:"rttAvgMs,omitempty" minimum:"0"`
	RttMedianMs   *float64       `json:"rttMedianMs,omitempty" minimum:"0"`
	RttMaxMs      *float64       `json:"rttMaxMs,omitempty" minimum:"0"`
	RttStddevMs   *float64       `json:"rttStddevMs,omitempty" minimum:"0"`
	RttSamplesMs  []float64      `json:"rttSamplesMs,omitempty"`
	ResolvedIP    *string        `json:"resolvedIp,omitempty"`
	IPFamily      *string        `json:"ipFamily,omitempty" enum:"inet,inet6"`
	Raw           map[string]any `json:"raw,omitempty"`
	ErrorCode     *string        `json:"errorCode,omitempty" maxLength:"100"`
	ErrorMessage  *string        `json:"errorMessage,omitempty" maxLength:"500"`
}

type submitResultsOutput struct{}

func newPingResultInputs(results []pingResultInputBody) ([]appproberuntime.PingResultInput, error) {
	mapped := make([]appproberuntime.PingResultInput, 0, len(results))
	for _, result := range results {
		resolvedIP, err := parseOptionalResultIP(result.ResolvedIP)
		if err != nil {
			return nil, err
		}
		ipFamily, err := domainnetwork.ParseOptionalIPFamily(result.IPFamily)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid IP family")
		}
		raw, err := rawMessage(result.Raw)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid raw result")
		}
		mapped = append(mapped, appproberuntime.PingResultInput{
			CheckID:       result.CheckID,
			StartedAt:     result.StartedAt,
			FinishedAt:    result.FinishedAt,
			DurationMs:    result.DurationMs,
			Status:        result.Status,
			SentCount:     result.SentCount,
			ReceivedCount: result.ReceivedCount,
			LossPercent:   result.LossPercent,
			RttMinMs:      result.RttMinMs,
			RttAvgMs:      result.RttAvgMs,
			RttMedianMs:   result.RttMedianMs,
			RttMaxMs:      result.RttMaxMs,
			RttStddevMs:   result.RttStddevMs,
			RttSamplesMs:  result.RttSamplesMs,
			ResolvedIP:    resolvedIP,
			IPFamily:      ipFamily,
			Raw:           raw,
			ErrorCode:     result.ErrorCode,
			ErrorMessage:  result.ErrorMessage,
		})
	}

	return mapped, nil
}

func newUnsupportedResultInputs(results []map[string]any) ([]appproberuntime.UnsupportedResultInput, error) {
	if len(results) == 0 {
		return nil, nil
	}

	mapped := make([]appproberuntime.UnsupportedResultInput, 0, len(results))
	for _, result := range results {
		raw, err := json.Marshal(result)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid unsupported result")
		}
		mapped = append(mapped, appproberuntime.UnsupportedResultInput{Raw: raw})
	}

	return mapped, nil
}

func parseOptionalResultIP(value *string) (*netip.Addr, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Omitted result IP fields are represented as nil.
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, huma.Error422UnprocessableEntity("invalid resolved IP")
	}
	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return nil, huma.Error422UnprocessableEntity("invalid resolved IP")
	}

	return &addr, nil
}

func rawMessage(raw map[string]any) (json.RawMessage, error) {
	if raw == nil {
		return nil, nil
	}

	return json.Marshal(raw)
}
