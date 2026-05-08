package project

import (
	"context"
	"errors"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
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
	defer flow.End()

	name, err := normalize.RequiredString(input.Name, ErrInvalidInput)
	if err != nil {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventCreateFailure, ProjectReasonInvalidInput, err)
	}
	slug, err := normalize.ProjectSlug(input.Slug, ErrInvalidInput)
	if err != nil {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventCreateFailure, ProjectReasonInvalidInput, err)
	}
	flow.SetProjectSlug(slug)

	project, err := s.repo.CreateProjectWithOwner(ctx, domainproject.CreateProjectStorageInput{
		Name:            name,
		Slug:            slug,
		CreatedByUserID: input.CurrentUserID,
	})
	if errors.Is(err, ErrProjectSlugAlreadyExists) {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventCreateFailure, ProjectReasonSlugAlreadyExists, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventCreateFailure, ProjectReasonUserNotFound, err)
	}
	if err != nil {
		return domainproject.Project{}, flow.TechnicalFailure(ProjectEventCreateFailure, ProjectReasonProjectCreateFailed, err)
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, input ListProjectsInput) ([]domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.list", ProjectActionList, input.CurrentUserID)
	defer flow.End()

	projects, err := s.repo.ListProjectsForUser(ctx, input.CurrentUserID)
	if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, flow.TechnicalFailure(ProjectEventListFailure, ProjectReasonProjectListFailed, err)
	}

	return projects, nil
}

func (s *Service) GetProject(ctx context.Context, input GetProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.get", ProjectActionGet, input.CurrentUserID)
	defer flow.End()

	return s.loadProjectForRead(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventGetFailure)
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.update", ProjectActionUpdate, input.CurrentUserID)
	defer flow.End()

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventUpdateFailure)
	if err != nil {
		return domainproject.Project{}, err
	}
	if _, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateFailure, isManager); err != nil {
		return domainproject.Project{}, err
	}

	name := project.Name
	slug := project.Slug
	if input.Name != nil {
		name, err = normalize.RequiredString(*input.Name, ErrInvalidInput)
		if err != nil {
			return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, err)
		}
	}
	if input.Slug != nil {
		slug, err = normalize.ProjectSlug(*input.Slug, ErrInvalidInput)
		if err != nil {
			return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, err)
		}
		flow.SetProjectSlug(slug)
	}
	if input.Name == nil && input.Slug == nil {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, ErrInvalidInput)
	}

	project, err = s.repo.UpdateProject(ctx, domainproject.UpdateProjectStorageInput{
		ProjectID: project.ID,
		Name:      name,
		Slug:      slug,
	})
	if errors.Is(err, ErrProjectNotFound) {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrProjectSlugAlreadyExists) {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonSlugAlreadyExists, err)
	}
	if err != nil {
		return domainproject.Project{}, flow.TechnicalFailure(ProjectEventUpdateFailure, ProjectReasonProjectUpdateFailed, err)
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	ctx, flow := s.startProjectFlow(ctx, "project.delete", ProjectActionDelete, input.CurrentUserID)
	defer flow.End()

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventDeleteFailure)
	if err != nil {
		return err
	}
	if _, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventDeleteFailure, isOwner); err != nil {
		return err
	}

	if err := s.repo.SoftDeleteProject(ctx, project.ID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return flow.BusinessFailure(ProjectEventDeleteFailure, ProjectReasonProjectNotFound, err)
		}
		return flow.TechnicalFailure(ProjectEventDeleteFailure, ProjectReasonProjectDeleteFailed, err)
	}
	flow.Success(ProjectEventDeleteSuccess)

	return nil
}

func (s *Service) ListMembers(ctx context.Context, input ListMembersInput) ([]domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.members.list", ProjectActionListMembers, input.CurrentUserID)
	defer flow.End()

	project, err := s.loadProjectForRead(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventListMembersFailure)
	if err != nil {
		return nil, err
	}
	if err := s.requireRoleForRead(ctx, flow, project.ID, input.CurrentUserID, ProjectEventListMembersFailure); err != nil {
		return nil, err
	}

	members, err := s.repo.ListMembers(ctx, project.ID)
	if errors.Is(err, ErrProjectNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, flow.TechnicalFailure(ProjectEventListMembersFailure, ProjectReasonMembersListFailed, err)
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, input AddMemberInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.add", ProjectActionAddMember, input.CurrentUserID)
	defer flow.End()
	flow.SetTargetUser(input.UserID)
	flow.SetRole(input.Role)

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventAddMemberFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventAddMemberFailure, isManager)
	if err != nil {
		return domainproject.Member{}, err
	}
	if err := validateAssignableRole(actorRole, input.Role); err != nil {
		return domainproject.Member{}, assignableRoleFailure(flow, ProjectEventAddMemberFailure, err)
	}

	member, err := s.repo.AddMember(ctx, domainproject.AddMemberStorageInput{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
	if errors.Is(err, ErrMemberAlreadyExists) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonMemberAlreadyExists, err)
	}
	if errors.Is(err, ErrProjectNotFound) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonUserNotFound, err)
	}
	if err != nil {
		return domainproject.Member{}, flow.TechnicalFailure(ProjectEventAddMemberFailure, ProjectReasonMemberAddFailed, err)
	}
	flow.Success(ProjectEventAddMemberSuccess)

	return member, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.role_update", ProjectActionUpdateMemberRole, input.CurrentUserID)
	defer flow.End()
	flow.SetTargetUser(input.UserID)
	flow.SetRole(input.Role)

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure, isManager)
	if err != nil {
		return domainproject.Member{}, err
	}
	if err := validateAssignableRole(actorRole, input.Role); err != nil {
		return domainproject.Member{}, assignableRoleFailure(flow, ProjectEventUpdateMemberRoleFailure, err)
	}

	member, err := s.repo.GetMember(ctx, project.ID, input.UserID)
	if errors.Is(err, ErrMemberNotFound) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberNotFound, err)
	}
	if err != nil {
		return domainproject.Member{}, flow.TechnicalFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberLookupFailed, err)
	}
	if actorRole == domainproject.RoleAdmin && member.Role == domainproject.RoleOwner {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if member.Role == domainproject.RoleOwner && input.Role != domainproject.RoleOwner {
		owners, err := s.repo.CountOwners(ctx, project.ID)
		if errors.Is(err, ErrProjectNotFound) {
			return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonProjectNotFound, err)
		}
		if err != nil {
			return domainproject.Member{}, flow.TechnicalFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonOwnerCountFailed, err)
		}
		if owners <= 1 {
			return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonLastOwner, ErrLastOwner)
		}
	}

	member, err = s.repo.UpdateMemberRole(ctx, domainproject.UpdateMemberRoleStorageInput{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
	if errors.Is(err, ErrMemberNotFound) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberNotFound, err)
	}
	if err != nil {
		return domainproject.Member{}, flow.TechnicalFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberRoleUpdateFailed, err)
	}
	flow.Success(ProjectEventUpdateMemberRoleSuccess)

	return member, nil
}

func (s *Service) loadProjectForUser(ctx context.Context, flow *projectFlow, projectRef string, userID string, failureEvent ProjectEventName) (domainproject.Project, error) {
	flow.SetProjectRef(projectRef)

	project, err := s.repo.GetProjectForUser(ctx, projectRef, userID)
	if errors.Is(err, ErrProjectNotFound) {
		return domainproject.Project{}, flow.BusinessFailure(failureEvent, ProjectReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return domainproject.Project{}, flow.BusinessFailure(failureEvent, ProjectReasonUserNotFound, err)
	}
	if err != nil {
		return domainproject.Project{}, flow.TechnicalFailure(failureEvent, ProjectReasonProjectLookupFailed, err)
	}

	flow.SetProject(project)
	return project, nil
}

func (s *Service) loadProjectForRead(ctx context.Context, flow *projectFlow, projectRef string, userID string, failureEvent ProjectEventName) (domainproject.Project, error) {
	flow.SetProjectRef(projectRef)

	project, err := s.repo.GetProjectForUser(ctx, projectRef, userID)
	if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrUserNotFound) {
		return domainproject.Project{}, err
	}
	if err != nil {
		return domainproject.Project{}, flow.TechnicalFailure(failureEvent, ProjectReasonProjectLookupFailed, err)
	}

	flow.SetProject(project)
	return project, nil
}

func (s *Service) requireRole(ctx context.Context, flow *projectFlow, projectID string, userID string, failureEvent ProjectEventName, allow func(domainproject.Role) bool) (domainproject.Role, error) {
	role, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if errors.Is(err, ErrProjectNotFound) {
		return "", flow.BusinessFailure(failureEvent, ProjectReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return "", flow.BusinessFailure(failureEvent, ProjectReasonUserNotFound, err)
	}
	if err != nil {
		return "", flow.TechnicalFailure(failureEvent, ProjectReasonRoleLookupFailed, err)
	}
	if !allow(role) {
		return "", flow.BusinessFailure(failureEvent, ProjectReasonForbidden, ErrForbidden)
	}

	return role, nil
}

func (s *Service) requireRoleForRead(ctx context.Context, flow *projectFlow, projectID string, userID string, failureEvent ProjectEventName) error {
	_, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrUserNotFound) {
		return err
	}
	if err != nil {
		return flow.TechnicalFailure(failureEvent, ProjectReasonRoleLookupFailed, err)
	}

	return nil
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

func isOwner(role domainproject.Role) bool {
	return role == domainproject.RoleOwner
}

func validateAssignableRole(actorRole domainproject.Role, targetRole domainproject.Role) error {
	if err := validateRole(targetRole); err != nil {
		return err
	}
	if !canAssignRole(actorRole, targetRole) {
		return ErrForbidden
	}

	return nil
}

func assignableRoleFailure(flow *projectFlow, event ProjectEventName, err error) error {
	if errors.Is(err, ErrInvalidRole) {
		return flow.BusinessFailure(event, ProjectReasonInvalidRole, err)
	}

	return flow.BusinessFailure(event, ProjectReasonForbidden, err)
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
