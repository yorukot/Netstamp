package project

import (
	"time"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type projectOutput struct {
	Body projectOutputBody
}

type projectOutputBody struct {
	Project projectResponse `json:"project"`
}

type memberOutput struct {
	Body memberOutputBody
}

type memberOutputBody struct {
	Member projectMemberResponse `json:"member"`
}

type projectResponse struct {
	ID              string    `json:"id" format:"uuid"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	CreatedByUserID string    `json:"createdByUserId" format:"uuid"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type projectMemberResponse struct {
	ID        string    `json:"id" format:"uuid"`
	ProjectID string    `json:"projectId" format:"uuid"`
	UserID    string    `json:"userId" format:"uuid"`
	Email     string    `json:"email" format:"email"`
	Role      string    `json:"role" enum:"owner,admin,editor,viewer"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newProjectResponse(project domainproject.Project) projectResponse {
	return projectResponse{
		ID:              project.ID,
		Name:            project.Name,
		Slug:            project.Slug,
		CreatedByUserID: project.CreatedByUserID,
		CreatedAt:       project.CreatedAt,
		UpdatedAt:       project.UpdatedAt,
	}
}

func newProjectMemberResponse(member domainproject.Member) projectMemberResponse {
	return projectMemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Email:     member.Email,
		Role:      string(member.Role),
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}
}
