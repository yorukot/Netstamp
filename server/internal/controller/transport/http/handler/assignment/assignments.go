package assignment

import (
	"context"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
)

func (h *Handler) listProjectAssignments(ctx context.Context, input *listProjectAssignmentsInput) (*listProjectAssignmentsOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.ListProjectAssignments(ctx, appassignment.ListProjectAssignmentsInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
	})
	if err != nil {
		return nil, mapAssignmentError(err, "list project assignments failed")
	}

	return &listProjectAssignmentsOutput{Body: listProjectAssignmentsOutputBody{
		Assignments: output.Assignments,
	}}, nil
}

type listProjectAssignmentsInput struct {
	Ref     string
	ProbeID string
	CheckID string
}

type listProjectAssignmentsOutput struct {
	Body listProjectAssignmentsOutputBody
}

type listProjectAssignmentsOutputBody struct {
	Assignments []domainassignment.Assignment `json:"assignments"`
}
