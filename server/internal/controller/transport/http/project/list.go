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

	return &listProjectsOutput{Body: listProjectsOutputBody{Projects: projects}}, nil
}

type listProjectsInput struct{}

type listProjectsOutput struct {
	Body listProjectsOutputBody
}

type listProjectsOutputBody struct {
	Projects []domainproject.Project `json:"projects"`
}
