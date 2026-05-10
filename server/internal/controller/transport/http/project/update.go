package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
)

func (h *Handler) updateProject(ctx context.Context, input *updateProjectInput) (*projectOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := h.service.UpdateProject(ctx, appproject.UpdateProjectInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Name:          input.Body.Name,
		Slug:          input.Body.Slug,
	})
	if err != nil {
		return nil, mapProjectError(err, "update project failed")
	}

	return &projectOutput{Body: projectOutputBody{Project: newProjectResponse(project)}}, nil
}

type updateProjectInput struct {
	Ref  string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	Body updateProjectInputBody
}

type updateProjectInputBody struct {
	Name *string `json:"name,omitempty" doc:"Project display name." example:"Engineering"`
	Slug *string `json:"slug,omitempty" patternDescription:"lowercase letters, numbers, and dashes" doc:"Stable project slug." example:"engineering"`
}
