package proberuntime

import (
	"context"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func (h *Handler) listAssignments(ctx context.Context, input *listAssignmentsInput) (*assignmentsOutput, error) {
	auth, err := runtimeAuthInput(input.ProbeID, input.Authorization)
	if err != nil {
		return nil, err
	}
	output, err := h.service.ListAssignments(ctx, auth)
	if err != nil {
		return nil, mapRuntimeError(err, "list probe runtime assignments failed")
	}

	assignments := make([]assignmentResponse, 0, len(output.Assignments))
	for _, assignment := range output.Assignments {
		assignments = append(assignments, newAssignmentResponse(assignment))
	}

	return &assignmentsOutput{Body: assignmentsOutputBody{Assignments: assignments}}, nil
}

type listAssignmentsInput struct {
	ProbeID       string `path:"probe_id" doc:"Probe ID."`
	Authorization string `header:"Authorization" doc:"Probe authorization header. Use 'Probe <secret>'."`
}

type assignmentsOutput struct {
	Body assignmentsOutputBody
}

type assignmentsOutputBody struct {
	Assignments []assignmentResponse `json:"assignments"`
}

type assignmentResponse struct {
	ID              string             `json:"assignmentId" format:"uuid"`
	ProjectID       string             `json:"projectId" format:"uuid"`
	ProbeID         string             `json:"probeId" format:"uuid"`
	CheckID         string             `json:"checkId" format:"uuid"`
	CheckVersion    string             `json:"checkVersion"`
	SelectorVersion string             `json:"selectorVersion"`
	Type            string             `json:"type" enum:"ping"`
	Target          string             `json:"target"`
	IntervalSeconds int32              `json:"intervalSeconds"`
	PingConfig      pingConfigResponse `json:"pingConfig"`
}

type pingConfigResponse struct {
	PacketCount     int32   `json:"packetCount"`
	PacketSizeBytes int32   `json:"packetSizeBytes"`
	TimeoutMs       int32   `json:"timeoutMs"`
	IPFamily        *string `json:"ipFamily,omitempty" enum:"inet,inet6"`
}

func newAssignmentResponse(assignment domainassignment.Assignment) assignmentResponse {
	check := assignment.Check
	pingConfig := domainping.DefaultConfig()
	if check != nil && check.PingConfig != nil {
		pingConfig = *check.PingConfig
	}

	var checkType string
	var target string
	var intervalSeconds int32
	if check != nil {
		checkType = string(check.Type)
		target = check.Target
		intervalSeconds = check.IntervalSeconds
	}

	return assignmentResponse{
		ID:              assignment.ID,
		ProjectID:       assignment.ProjectID,
		ProbeID:         assignment.ProbeID,
		CheckID:         assignment.CheckID,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Type:            checkType,
		Target:          target,
		IntervalSeconds: intervalSeconds,
		PingConfig: pingConfigResponse{
			PacketCount:     pingConfig.PacketCount,
			PacketSizeBytes: pingConfig.PacketSizeBytes,
			TimeoutMs:       pingConfig.TimeoutMs,
			IPFamily:        ipFamilyString(pingConfig.IPFamily),
		},
	}
}

func ipFamilyString(ipFamily *domainnetwork.IPFamily) *string {
	if ipFamily == nil {
		return nil
	}

	value := string(*ipFamily)
	return &value
}
