package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
)

func (h *Handler) getProject(ctx context.Context, input *projectRefInput) (*projectOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := h.service.GetProject(ctx, appproject.GetProjectInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapProjectError(err, "get project failed")
	}

	return &projectOutput{Body: projectOutputBody{Project: project}}, nil
}

type projectRefInput struct {
	Ref string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project UUID or slug." example:"engineering"`
}
