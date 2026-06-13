package proberuntime

import (
	"context"
	"time"

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

	return &assignmentsOutput{Body: assignmentsOutputBody{
		ServerTime:  output.ServerTime,
		Assignments: output.Assignments,
	}}, nil
}

type listAssignmentsInput struct {
	ProbeID string
}

type assignmentsOutput struct {
	Body assignmentsOutputBody
}

type assignmentsOutputBody struct {
	ServerTime  time.Time                     `json:"serverTime"`
	Assignments []domainassignment.Assignment `json:"assignments"`
}
