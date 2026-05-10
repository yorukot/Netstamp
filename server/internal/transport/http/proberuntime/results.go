package proberuntime

import (
	"context"
	"encoding/json"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func (h *Handler) submitResults(ctx context.Context, input *submitResultsInput) (*submitResultsOutput, error) {
	auth, err := runtimeAuthInput(input.ProbeID, input.Authorization)
	if err != nil {
		return nil, err
	}
	groups, err := newResultGroupInputs(input.Body)
	if err != nil {
		return nil, err
	}
	output, err := h.service.SubmitResults(ctx, appproberuntime.SubmitResultsInput{
		RuntimeAuthInput: auth,
		Groups:           groups,
	})
	if err != nil {
		return nil, mapRuntimeError(err, "submit probe results failed")
	}

	return &submitResultsOutput{Body: newSubmitResultsOutputBody(output)}, nil
}

type submitResultsInput struct {
	ProbeID       string `path:"probe_id" doc:"Probe ID."`
	Authorization string `header:"Authorization" doc:"Probe authorization header. Use 'Probe <secret>'."`
	Body          submitResultsInputBody
}

type submitResultsInputBody map[string]resultGroupInputBody

type resultGroupInputBody struct {
	Type    string                     `json:"type,omitempty"`
	Detail  resultGroupDetailInputBody `json:"detail,omitempty"`
	Results []pingResultInputBody      `json:"results,omitempty"`
}

type resultGroupDetailInputBody struct {
	AssignmentID    string `json:"assignmentId,omitempty" format:"uuid"`
	CheckVersion    string `json:"checkVersion,omitempty"`
	SelectorVersion string `json:"selectorVersion,omitempty"`
}

type pingResultInputBody struct {
	StartedAt     time.Time      `json:"startedAt,omitempty"`
	FinishedAt    time.Time      `json:"finishedAt,omitempty"`
	DurationMs    int32          `json:"durationMs,omitempty"`
	Status        string         `json:"status,omitempty"`
	SentCount     int32          `json:"sentCount,omitempty"`
	ReceivedCount int32          `json:"receivedCount,omitempty"`
	LossPercent   float64        `json:"lossPercent,omitempty"`
	RttMinMs      *float64       `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64       `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64       `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64       `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64       `json:"rttStddevMs,omitempty"`
	RttSamplesMs  []float64      `json:"rttSamplesMs,omitempty"`
	ResolvedIP    *string        `json:"resolvedIp,omitempty"`
	IPFamily      *string        `json:"ipFamily,omitempty"`
	Raw           map[string]any `json:"raw,omitempty"`
	ErrorCode     *string        `json:"errorCode,omitempty"`
	ErrorMessage  *string        `json:"errorMessage,omitempty"`
}

type submitResultsOutput struct {
	Body submitResultsOutputBody
}

type submitResultsOutputBody struct {
	Accepted     bool                 `json:"accepted"`
	ResyncNeeded bool                 `json:"resyncNeeded"`
	StaleChecks  []string             `json:"staleChecks"`
	Assignments  []assignmentResponse `json:"assignments"`
}

func newResultGroupInputs(groups submitResultsInputBody) ([]appproberuntime.ResultGroupInput, error) {
	checkIDs := make([]string, 0, len(groups))
	for checkID := range groups {
		checkIDs = append(checkIDs, checkID)
	}
	slices.Sort(checkIDs)

	mapped := make([]appproberuntime.ResultGroupInput, 0, len(groups))
	for _, checkID := range checkIDs {
		group := groups[checkID]
		pingResults, err := newPingResultInputs(group.Results)
		if err != nil {
			return nil, err
		}
		mapped = append(mapped, appproberuntime.ResultGroupInput{
			CheckID:         checkID,
			Type:            domaincheck.Type(strings.TrimSpace(group.Type)),
			AssignmentID:    group.Detail.AssignmentID,
			CheckVersion:    group.Detail.CheckVersion,
			SelectorVersion: group.Detail.SelectorVersion,
			PingResults:     pingResults,
		})
	}

	return mapped, nil
}

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

func newSubmitResultsOutputBody(output appproberuntime.SubmitResultsOutput) submitResultsOutputBody {
	assignments := make([]assignmentResponse, 0, len(output.Assignments))
	for _, assignment := range output.Assignments {
		assignments = append(assignments, newAssignmentResponse(assignment))
	}

	return submitResultsOutputBody{
		Accepted:     output.Accepted,
		ResyncNeeded: output.ResyncNeeded,
		StaleChecks:  output.StaleChecks,
		Assignments:  assignments,
	}
}
