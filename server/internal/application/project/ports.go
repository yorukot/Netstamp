package project

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	CreateProjectWithOwner(ctx context.Context, input domainproject.CreateProjectStorageInput) (domainproject.Project, error)
	ListProjectsForUser(ctx context.Context, userID string) ([]domainproject.Project, error)
	GetProjectForUser(ctx context.Context, projectRef string, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID string, userID string) (domainproject.Role, error)
	UpdateProject(ctx context.Context, input domainproject.UpdateProjectStorageInput) (domainproject.Project, error)
	SoftDeleteProject(ctx context.Context, projectID string) error
	ListMembers(ctx context.Context, projectID string) ([]domainproject.Member, error)
	GetMember(ctx context.Context, projectID string, userID string) (domainproject.Member, error)
	AddMember(ctx context.Context, input domainproject.AddMemberStorageInput) (domainproject.Member, error)
	UpdateMemberRole(ctx context.Context, input domainproject.UpdateMemberRoleStorageInput) (domainproject.Member, error)
	CountOwners(ctx context.Context, projectID string) (int, error)
}
