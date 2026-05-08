package project

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (domainproject.Project, error) {
	name, err := normalize.RequiredString(input.Name, ErrInvalidInput)
	if err != nil {
		return domainproject.Project{}, err
	}
	slug, err := normalize.ProjectSlug(input.Slug, ErrInvalidInput)
	if err != nil {
		return domainproject.Project{}, err
	}

	return s.repo.CreateProjectWithOwner(ctx, CreateProjectStorageInput{
		Name:            name,
		Slug:            slug,
		CreatedByUserID: input.CurrentUserID,
	})
}

func (s *Service) ListProjects(ctx context.Context, input ListProjectsInput) ([]domainproject.Project, error) {
	return s.repo.ListProjectsForUser(ctx, input.CurrentUserID)
}

func (s *Service) GetProject(ctx context.Context, input GetProjectInput) (domainproject.Project, error) {
	return s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (domainproject.Project, error) {
	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Project{}, err
	}

	role, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Project{}, err
	}
	if !isManager(role) {
		return domainproject.Project{}, ErrForbidden
	}

	name := project.Name
	slug := project.Slug
	if input.Name != nil {
		name, err = normalize.RequiredString(*input.Name, ErrInvalidInput)
		if err != nil {
			return domainproject.Project{}, err
		}
	}
	if input.Slug != nil {
		slug, err = normalize.ProjectSlug(*input.Slug, ErrInvalidInput)
		if err != nil {
			return domainproject.Project{}, err
		}
	}
	if input.Name == nil && input.Slug == nil {
		return domainproject.Project{}, ErrInvalidInput
	}

	return s.repo.UpdateProject(ctx, UpdateProjectStorageInput{
		ProjectID: project.ID,
		Name:      name,
		Slug:      slug,
	})
}

func (s *Service) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}

	role, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return err
	}
	if role != domainproject.RoleOwner {
		return ErrForbidden
	}

	return s.repo.SoftDeleteProject(ctx, project.ID)
}

func (s *Service) ListMembers(ctx context.Context, input ListMembersInput) ([]domainproject.Member, error) {
	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID); err != nil {
		return nil, err
	}

	return s.repo.ListMembers(ctx, project.ID)
}

func (s *Service) AddMember(ctx context.Context, input AddMemberInput) (domainproject.Member, error) {
	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, err
	}

	actorRole, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, err
	}
	if !isManager(actorRole) {
		return domainproject.Member{}, ErrForbidden
	}
	if err := validateRole(input.Role); err != nil {
		return domainproject.Member{}, err
	}
	if !canAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, ErrForbidden
	}

	return s.repo.AddMember(ctx, AddMemberStorageInput{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
}

func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (domainproject.Member, error) {
	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, err
	}

	actorRole, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, err
	}
	if !isManager(actorRole) {
		return domainproject.Member{}, ErrForbidden
	}
	if err := validateRole(input.Role); err != nil {
		return domainproject.Member{}, err
	}
	if !canAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, ErrForbidden
	}

	member, err := s.repo.GetMember(ctx, project.ID, input.UserID)
	if err != nil {
		return domainproject.Member{}, err
	}
	if actorRole == domainproject.RoleAdmin && member.Role == domainproject.RoleOwner {
		return domainproject.Member{}, ErrForbidden
	}
	if member.Role == domainproject.RoleOwner && input.Role != domainproject.RoleOwner {
		owners, err := s.repo.CountOwners(ctx, project.ID)
		if err != nil {
			return domainproject.Member{}, err
		}
		if owners <= 1 {
			return domainproject.Member{}, ErrLastOwner
		}
	}

	return s.repo.UpdateMemberRole(ctx, UpdateMemberRoleStorageInput{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
}

func validateRole(role domainproject.Role) error {
	switch role {
	case domainproject.RoleOwner, domainproject.RoleAdmin, domainproject.RoleEditor, domainproject.RoleViewer:
		return nil
	default:
		return ErrInvalidRole
	}
}

func isManager(role domainproject.Role) bool {
	return role == domainproject.RoleOwner || role == domainproject.RoleAdmin
}

func canAssignRole(actorRole domainproject.Role, targetRole domainproject.Role) bool {
	switch actorRole {
	case domainproject.RoleOwner:
		return targetRole == domainproject.RoleAdmin || targetRole == domainproject.RoleEditor || targetRole == domainproject.RoleViewer
	case domainproject.RoleAdmin:
		return targetRole == domainproject.RoleEditor || targetRole == domainproject.RoleViewer
	default:
		return false
	}
}
