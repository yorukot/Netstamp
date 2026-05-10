package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
)

func (h *Handler) createProject(ctx context.Context, input *createProjectInput) (*projectOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := h.service.CreateProject(ctx, appproject.CreateProjectInput{
		CurrentUserID: currentUserID,
		Name:          input.Body.Name,
		Slug:          input.Body.Slug,
	})
	if err != nil {
		return nil, mapProjectError(err, "create project failed")
	}

	return &projectOutput{Body: projectOutputBody{Project: newProjectResponse(project)}}, nil
}

type createProjectInput struct {
	Body createProjectInputBody
}

type createProjectInputBody struct {
	Name string `json:"name,omitempty" doc:"Project display name." example:"Engineering"`
	Slug string `json:"slug,omitempty" patternDescription:"lowercase letters, numbers, and dashes" doc:"Stable project slug." example:"engineering"`
}
