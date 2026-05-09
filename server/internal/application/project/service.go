package project

import (
	"context"
	"errors"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo   Repository
	events EventRecorder
}

func NewService(repo Repository, events EventRecorder) *Service {
	return &Service{
		repo:   repo,
		events: events,
	}
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.create", ProjectActionCreate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeCreateProjectInput(input)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventCreateFailure, ProjectReasonInvalidInput, err)
	}
	flow.setProjectSlug(normalized.slug)

	project, err := s.repo.CreateProjectWithOwner(ctx, domainproject.CreateProjectStorageInput{
		Name:            normalized.name,
		Slug:            normalized.slug,
		CreatedByUserID: input.CurrentUserID,
	})
	if err != nil {
		return domainproject.Project{}, flow.projectCreateFailure(err)
	}
	flow.setProject(project)
	flow.success(ProjectEventCreateSuccess)

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, input ListProjectsInput) ([]domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.list", ProjectActionList, input.CurrentUserID)
	defer flow.end()

	projects, err := s.repo.ListProjectsForUser(ctx, input.CurrentUserID)
	if err != nil {
		return nil, flow.projectListFailure(err)
	}

	return projects, nil
}

func (s *Service) GetProject(ctx context.Context, input GetProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.get", ProjectActionGet, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventGetFailure, ProjectReasonInvalidInput, err)
	}

	return s.loadProjectForRead(ctx, flow, projectRef, input.CurrentUserID, ProjectEventGetFailure)
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.update", ProjectActionUpdate, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, err)
	}
	normalized, err := normalizeUpdateProjectInput(input)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, err)
	}

	project, err := s.loadProjectForUser(ctx, flow, projectRef, input.CurrentUserID, ProjectEventUpdateFailure)
	if err != nil {
		return domainproject.Project{}, err
	}
	_, err = s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateFailure, domainproject.ActionUpdateProject)
	if err != nil {
		return domainproject.Project{}, err
	}

	name := project.Name
	slug := project.Slug
	if normalized.name != nil {
		name = *normalized.name
	}
	if normalized.slug != nil {
		slug = *normalized.slug
		flow.setProjectSlug(slug)
	}

	project, err = s.repo.UpdateProject(ctx, domainproject.UpdateProjectStorageInput{
		ProjectID: project.ID,
		Name:      name,
		Slug:      slug,
	})
	if err != nil {
		return domainproject.Project{}, flow.projectUpdateFailure(err)
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	ctx, flow := s.startProjectFlow(ctx, "project.delete", ProjectActionDelete, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return flow.businessFailure(ProjectEventDeleteFailure, ProjectReasonInvalidInput, err)
	}

	project, err := s.loadProjectForUser(ctx, flow, projectRef, input.CurrentUserID, ProjectEventDeleteFailure)
	if err != nil {
		return err
	}
	_, err = s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventDeleteFailure, domainproject.ActionDeleteProject)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDeleteProject(ctx, project.ID); err != nil {
		return flow.projectDeleteFailure(err)
	}
	flow.success(ProjectEventDeleteSuccess)

	return nil
}

func (s *Service) ListMembers(ctx context.Context, input ListMembersInput) ([]domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.members.list", ProjectActionListMembers, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return nil, flow.businessFailure(ProjectEventListMembersFailure, ProjectReasonInvalidInput, err)
	}

	project, err := s.loadProjectForRead(ctx, flow, projectRef, input.CurrentUserID, ProjectEventListMembersFailure)
	if err != nil {
		return nil, err
	}
	_, err = s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventListMembersFailure, domainproject.ActionReadProject)
	if err != nil {
		return nil, err
	}

	members, err := s.repo.ListMembers(ctx, project.ID)
	if err != nil {
		return nil, flow.membersListFailure(err)
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, input AddMemberInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.add", ProjectActionAddMember, input.CurrentUserID)
	defer flow.end()
	flow.setTargetUser(input.UserID)
	flow.setRole(input.Role)

	normalized, err := normalizeAddMemberInput(input)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			return domainproject.Member{}, flow.assignableRoleFailure(ProjectEventAddMemberFailure, err)
		}
		return domainproject.Member{}, flow.businessFailure(ProjectEventAddMemberFailure, ProjectReasonInvalidInput, err)
	}
	flow.setTargetUser(normalized.userID)
	flow.setRole(normalized.role)

	project, err := s.loadProjectForUser(ctx, flow, normalized.projectRef, input.CurrentUserID, ProjectEventAddMemberFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventAddMemberFailure, domainproject.ActionManageMembers)
	if err != nil {
		return domainproject.Member{}, err
	}
	err = validateAssignableRole(actorRole, normalized.role)
	if err != nil {
		return domainproject.Member{}, flow.assignableRoleFailure(ProjectEventAddMemberFailure, err)
	}

	member, err := s.repo.AddMember(ctx, domainproject.AddMemberStorageInput{
		ProjectID: project.ID,
		UserID:    normalized.userID,
		Role:      normalized.role,
	})
	if err != nil {
		return domainproject.Member{}, flow.memberAddFailure(err)
	}
	flow.success(ProjectEventAddMemberSuccess)

	return member, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.role_update", ProjectActionUpdateMemberRole, input.CurrentUserID)
	defer flow.end()
	flow.setTargetUser(input.UserID)
	flow.setRole(input.Role)

	normalized, err := normalizeUpdateMemberRoleInput(input)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			return domainproject.Member{}, flow.assignableRoleFailure(ProjectEventUpdateMemberRoleFailure, err)
		}
		return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonInvalidInput, err)
	}
	flow.setTargetUser(normalized.userID)
	flow.setRole(normalized.role)

	project, err := s.loadProjectForUser(ctx, flow, normalized.projectRef, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure, domainproject.ActionManageMembers)
	if err != nil {
		return domainproject.Member{}, err
	}
	err = validateAssignableRole(actorRole, normalized.role)
	if err != nil {
		return domainproject.Member{}, flow.assignableRoleFailure(ProjectEventUpdateMemberRoleFailure, err)
	}

	member, err := s.repo.GetMember(ctx, project.ID, normalized.userID)
	if err != nil {
		return domainproject.Member{}, flow.memberLookupFailure(ProjectEventUpdateMemberRoleFailure, err)
	}
	if actorRole == domainproject.RoleAdmin && member.Role == domainproject.RoleOwner {
		return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if member.Role == domainproject.RoleOwner && normalized.role != domainproject.RoleOwner {
		var owners int
		owners, err = s.repo.CountOwners(ctx, project.ID)
		if err != nil {
			return domainproject.Member{}, flow.ownerCountFailure(ProjectEventUpdateMemberRoleFailure, err)
		}
		if owners <= 1 {
			return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonLastOwner, ErrLastOwner)
		}
	}

	member, err = s.repo.UpdateMemberRole(ctx, domainproject.UpdateMemberRoleStorageInput{
		ProjectID: project.ID,
		UserID:    normalized.userID,
		Role:      normalized.role,
	})
	if err != nil {
		return domainproject.Member{}, flow.memberRoleUpdateFailure(err)
	}
	flow.success(ProjectEventUpdateMemberRoleSuccess)

	return member, nil
}

func (s *Service) loadProjectForUser(ctx context.Context, flow *projectFlow, projectRef, userID string, failureEvent ProjectEventName) (domainproject.Project, error) {
	flow.setProjectRef(projectRef)

	project, err := s.repo.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(failureEvent, err)
	}

	flow.setProject(project)
	return project, nil
}

func (s *Service) loadProjectForRead(ctx context.Context, flow *projectFlow, projectRef, userID string, failureEvent ProjectEventName) (domainproject.Project, error) {
	flow.setProjectRef(projectRef)

	project, err := s.repo.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectReadLookupFailure(failureEvent, err)
	}

	flow.setProject(project)
	return project, nil
}

func (s *Service) requireRole(ctx context.Context, flow *projectFlow, projectID, userID string, failureEvent ProjectEventName, action domainproject.Action) (domainproject.Role, error) {
	role, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return "", flow.roleLookupFailure(failureEvent, err)
	}
	if !domainproject.Can(role, action) {
		return "", flow.businessFailure(failureEvent, ProjectReasonForbidden, ErrForbidden)
	}

	return role, nil
}

func validateRole(role domainproject.Role) error {
	if domainproject.IsValidRole(role) {
		return nil
	}

	return ErrInvalidRole
}

func validateAssignableRole(actorRole, targetRole domainproject.Role) error {
	if err := validateRole(targetRole); err != nil {
		return err
	}
	if !domainproject.CanAssignRole(actorRole, targetRole) {
		return ErrForbidden
	}

	return nil
}
