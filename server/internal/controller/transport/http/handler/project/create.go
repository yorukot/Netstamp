package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) createProject(ctx context.Context, input *createProjectInput) (*projectOutput, error) {
	if !h.creationEnabled {
		return nil, httpx.Forbidden("project creation is disabled")
	}

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

	return &projectOutput{Body: projectOutputBody{Project: project}}, nil
}

type createProjectInput struct {
	Body createProjectInputBody
}

type createProjectInputBody struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}
