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
	if events == nil {
		events = noopProjectEventRecorder{}
	}

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
	if err != nil {
		return domainproject.Project{}, flowFailure(flow, ProjectEventCreateFailure, ProjectReasonProjectCreateFailed, err)
	}
	flow.SetProject(project)
	flow.Success(ProjectEventCreateSuccess)

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, input ListProjectsInput) ([]domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.list", ProjectActionList, input.CurrentUserID)
	defer flow.End()

	projects, err := s.repo.ListProjectsForUser(ctx, input.CurrentUserID)
	if err != nil {
		return nil, readFailure(flow, ProjectEventListFailure, ProjectReasonProjectListFailed, err)
	}

	return projects, nil
}

func (s *Service) GetProject(ctx context.Context, input GetProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.get", ProjectActionGet, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Project{}, readFailure(flow, ProjectEventGetFailure, ProjectReasonProjectLookupFailed, err)
	}

	return project, nil
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.update", ProjectActionUpdate, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Project{}, flowFailure(flow, ProjectEventUpdateFailure, ProjectReasonProjectLookupFailed, err)
	}
	flow.SetProject(project)

	role, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Project{}, flowFailure(flow, ProjectEventUpdateFailure, ProjectReasonRoleLookupFailed, err)
	}
	if !isManager(role) {
		return domainproject.Project{}, flow.BusinessFailure(ProjectEventUpdateFailure, ProjectReasonForbidden, ErrForbidden)
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
	if err != nil {
		return domainproject.Project{}, flowFailure(flow, ProjectEventUpdateFailure, ProjectReasonProjectUpdateFailed, err)
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	ctx, flow := s.startProjectFlow(ctx, "project.delete", ProjectActionDelete, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return flowFailure(flow, ProjectEventDeleteFailure, ProjectReasonProjectLookupFailed, err)
	}
	flow.SetProject(project)

	role, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return flowFailure(flow, ProjectEventDeleteFailure, ProjectReasonRoleLookupFailed, err)
	}
	if role != domainproject.RoleOwner {
		return flow.BusinessFailure(ProjectEventDeleteFailure, ProjectReasonForbidden, ErrForbidden)
	}

	if err := s.repo.SoftDeleteProject(ctx, project.ID); err != nil {
		return flowFailure(flow, ProjectEventDeleteFailure, ProjectReasonProjectDeleteFailed, err)
	}
	flow.Success(ProjectEventDeleteSuccess)

	return nil
}

func (s *Service) ListMembers(ctx context.Context, input ListMembersInput) ([]domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.members.list", ProjectActionListMembers, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, readFailure(flow, ProjectEventListMembersFailure, ProjectReasonProjectLookupFailed, err)
	}
	flow.SetProject(project)

	if _, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID); err != nil {
		return nil, readFailure(flow, ProjectEventListMembersFailure, ProjectReasonRoleLookupFailed, err)
	}

	members, err := s.repo.ListMembers(ctx, project.ID)
	if err != nil {
		return nil, readFailure(flow, ProjectEventListMembersFailure, ProjectReasonMembersListFailed, err)
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, input AddMemberInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.add", ProjectActionAddMember, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)
	flow.SetTargetUser(input.UserID)
	flow.SetRole(input.Role)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventAddMemberFailure, ProjectReasonProjectLookupFailed, err)
	}
	flow.SetProject(project)

	actorRole, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventAddMemberFailure, ProjectReasonRoleLookupFailed, err)
	}
	if !isManager(actorRole) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if err := validateRole(input.Role); err != nil {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonInvalidRole, err)
	}
	if !canAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventAddMemberFailure, ProjectReasonForbidden, ErrForbidden)
	}

	member, err := s.repo.AddMember(ctx, domainproject.AddMemberStorageInput{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventAddMemberFailure, ProjectReasonMemberAddFailed, err)
	}
	flow.Success(ProjectEventAddMemberSuccess)

	return member, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (domainproject.Member, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.member.role_update", ProjectActionUpdateMemberRole, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)
	flow.SetTargetUser(input.UserID)
	flow.SetRole(input.Role)

	project, err := s.repo.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventUpdateMemberRoleFailure, ProjectReasonProjectLookupFailed, err)
	}
	flow.SetProject(project)

	actorRole, err := s.repo.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventUpdateMemberRoleFailure, ProjectReasonRoleLookupFailed, err)
	}
	if !isManager(actorRole) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if err := validateRole(input.Role); err != nil {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonInvalidRole, err)
	}
	if !canAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}

	member, err := s.repo.GetMember(ctx, project.ID, input.UserID)
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberLookupFailed, err)
	}
	if actorRole == domainproject.RoleAdmin && member.Role == domainproject.RoleOwner {
		return domainproject.Member{}, flow.BusinessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if member.Role == domainproject.RoleOwner && input.Role != domainproject.RoleOwner {
		owners, err := s.repo.CountOwners(ctx, project.ID)
		if err != nil {
			return domainproject.Member{}, flowFailure(flow, ProjectEventUpdateMemberRoleFailure, ProjectReasonOwnerCountFailed, err)
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
	if err != nil {
		return domainproject.Member{}, flowFailure(flow, ProjectEventUpdateMemberRoleFailure, ProjectReasonMemberRoleUpdateFailed, err)
	}
	flow.Success(ProjectEventUpdateMemberRoleSuccess)

	return member, nil
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

func flowFailure(flow *projectFlow, name ProjectEventName, technicalReason ProjectEventReason, err error) error {
	if reason, ok := projectBusinessReason(err); ok {
		return flow.BusinessFailure(name, reason, err)
	}

	return flow.TechnicalFailure(name, technicalReason, err)
}

func readFailure(flow *projectFlow, name ProjectEventName, technicalReason ProjectEventReason, err error) error {
	if _, ok := projectBusinessReason(err); ok {
		return err
	}

	return flow.TechnicalFailure(name, technicalReason, err)
}

func projectBusinessReason(err error) (ProjectEventReason, bool) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return ProjectReasonInvalidInput, true
	case errors.Is(err, ErrInvalidRole):
		return ProjectReasonInvalidRole, true
	case errors.Is(err, ErrForbidden):
		return ProjectReasonForbidden, true
	case errors.Is(err, ErrProjectNotFound):
		return ProjectReasonProjectNotFound, true
	case errors.Is(err, ErrProjectSlugAlreadyExists):
		return ProjectReasonSlugAlreadyExists, true
	case errors.Is(err, ErrMemberAlreadyExists):
		return ProjectReasonMemberAlreadyExists, true
	case errors.Is(err, ErrMemberNotFound):
		return ProjectReasonMemberNotFound, true
	case errors.Is(err, ErrUserNotFound):
		return ProjectReasonUserNotFound, true
	case errors.Is(err, ErrLastOwner):
		return ProjectReasonLastOwner, true
	default:
		return "", false
	}
}

type noopProjectEventRecorder struct{}

func (noopProjectEventRecorder) RecordProjectEvent(context.Context, ProjectEvent) {}
