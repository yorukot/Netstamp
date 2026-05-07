package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/application/project"
)

func (h *Handler) deleteProject(ctx context.Context, input *projectRefInput) (*deleteProjectOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteProject(ctx, appproject.DeleteProjectInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	}); err != nil {
		return nil, mapProjectError(err, "delete project failed")
	}

	return &deleteProjectOutput{}, nil
}

type deleteProjectOutput struct{}
