package proberuntime

import (
	"context"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
)

func (h *Handler) listAssignments(ctx context.Context, _ *listAssignmentsInput) (*assignmentsOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}
	output, err := h.service.ListAssignments(ctx, auth)
	if err != nil {
		return nil, mapRuntimeError(err, "list probe runtime assignments failed")
	}

	return &assignmentsOutput{Body: assignmentsOutputBody{Assignments: output.Assignments}}, nil
}

type listAssignmentsInput struct {
	ProbeID string `path:"probe_id" doc:"Probe ID."`
}

type assignmentsOutput struct {
	Body assignmentsOutputBody
}

type assignmentsOutputBody struct {
	Assignments []domainassignment.Assignment `json:"assignments"`
}
