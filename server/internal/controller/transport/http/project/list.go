package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (h *Handler) listProjects(ctx context.Context, _ *listProjectsInput) (*listProjectsOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := h.service.ListProjects(ctx, appproject.ListProjectsInput{CurrentUserID: currentUserID})
	if err != nil {
		return nil, mapProjectError(err, "list projects failed")
	}

	return &listProjectsOutput{Body: listProjectsOutputBody{Projects: newProjectResponses(projects)}}, nil
}

type listProjectsInput struct{}

type listProjectsOutput struct {
	Body listProjectsOutputBody
}

type listProjectsOutputBody struct {
	Projects []projectResponse `json:"projects"`
}

func newProjectResponses(projects []domainproject.Project) []projectResponse {
	responses := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		responses = append(responses, newProjectResponse(project))
	}

	return responses
}
