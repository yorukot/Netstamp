package project

import (
	"context"

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

	input, err := normalizeCreateProjectInput(input)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventCreateFailure, ProjectReasonInvalidInput, err)
	}
	flow.setProjectSlug(input.Slug)

	project, err := s.repo.CreateProjectWithOwner(ctx, domainproject.Project{
		Name:            input.Name,
		Slug:            input.Slug,
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

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		err = invalidProjectField("projectRef", err.Error(), input.ProjectRef)
		return domainproject.Project{}, flow.businessFailure(ProjectEventGetFailure, ProjectReasonInvalidInput, err)
	}

	return s.loadProjectForRead(ctx, flow, projectRef, input.CurrentUserID, ProjectEventGetFailure)
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (domainproject.Project, error) {
	ctx, flow := s.startProjectFlow(ctx, "project.update", ProjectActionUpdate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeUpdateProjectInput(input)
	if err != nil {
		return domainproject.Project{}, flow.businessFailure(ProjectEventUpdateFailure, ProjectReasonInvalidInput, err)
	}
	flow.setProjectSlug(input.ProjectRef)

	// load user's project
	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventUpdateFailure)
	if err != nil {
		return domainproject.Project{}, err
	}

	// make sure user have access to this action
	_, err = s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateFailure, domainproject.ActionUpdateProject)
	if err != nil {
		return domainproject.Project{}, err
	}

	name := project.Name
	slug := project.Slug
	if input.Name != nil {
		name = *input.Name
	}
	if input.Slug != nil {
		slug = *input.Slug
		flow.setProjectSlug(slug)
	}

	project, err = s.repo.UpdateProject(ctx, domainproject.Project{
		ID:   project.ID,
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return domainproject.Project{}, flow.projectUpdateFailure(err)
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	ctx, flow := s.startProjectFlow(ctx, "project.delete", ProjectActionDelete, input.CurrentUserID)
	defer flow.end()

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		err = invalidProjectField("projectRef", err.Error(), input.ProjectRef)
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

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		err = invalidProjectField("projectRef", err.Error(), input.ProjectRef)
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

	input, err := normalizeAddMemberInput(input)
	if err != nil {
		return domainproject.Member{}, flow.businessFailure(ProjectEventAddMemberFailure, ProjectReasonInvalidInput, err)
	}
	flow.setTargetUser(input.UserID)
	flow.setRole(input.Role)

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventAddMemberFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventAddMemberFailure, domainproject.ActionManageMembers)
	if err != nil {
		return domainproject.Member{}, err
	}
	if !domainproject.CanAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, flow.businessFailure(ProjectEventAddMemberFailure, ProjectReasonForbidden, ErrForbidden)
	}

	member, err := s.repo.AddMember(ctx, domainproject.Member{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
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

	input, err := normalizeUpdateMemberRoleInput(input)
	if err != nil {
		return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonInvalidInput, err)
	}
	flow.setTargetUser(input.UserID)
	flow.setRole(input.Role)

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure)
	if err != nil {
		return domainproject.Member{}, err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventUpdateMemberRoleFailure, domainproject.ActionManageMembers)
	if err != nil {
		return domainproject.Member{}, err
	}
	if !domainproject.CanAssignRole(actorRole, input.Role) {
		return domainproject.Member{}, flow.businessFailure(ProjectEventAddMemberFailure, ProjectReasonForbidden, ErrForbidden)
	}

	member, err := s.repo.GetMember(ctx, project.ID, input.UserID)
	if err != nil {
		return domainproject.Member{}, flow.memberLookupFailure(ProjectEventUpdateMemberRoleFailure, err)
	}
	if actorRole == domainproject.RoleAdmin && member.Role == domainproject.RoleOwner {
		return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if member.Role == domainproject.RoleOwner && input.Role != domainproject.RoleOwner {
		var owners int
		owners, err = s.repo.CountOwners(ctx, project.ID)
		if err != nil {
			return domainproject.Member{}, flow.ownerCountFailure(ProjectEventUpdateMemberRoleFailure, err)
		}
		if owners <= 1 {
			return domainproject.Member{}, flow.businessFailure(ProjectEventUpdateMemberRoleFailure, ProjectReasonLastOwner, ErrLastOwner)
		}
	}

	member, err = s.repo.UpdateMemberRole(ctx, domainproject.Member{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      input.Role,
	})
	if err != nil {
		return domainproject.Member{}, flow.memberRoleUpdateFailure(err)
	}
	flow.success(ProjectEventUpdateMemberRoleSuccess)

	return member, nil
}

func (s *Service) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	ctx, flow := s.startProjectFlow(ctx, "project.member.remove", ProjectActionRemoveMember, input.CurrentUserID)
	defer flow.end()
	flow.setTargetUser(input.UserID)

	input, err := normalizeRemoveMemberInput(input)
	if err != nil {
		return flow.businessFailure(ProjectEventRemoveMemberFailure, ProjectReasonInvalidInput, err)
	}
	flow.setTargetUser(input.UserID)

	project, err := s.loadProjectForUser(ctx, flow, input.ProjectRef, input.CurrentUserID, ProjectEventRemoveMemberFailure)
	if err != nil {
		return err
	}
	actorRole, err := s.requireRole(ctx, flow, project.ID, input.CurrentUserID, ProjectEventRemoveMemberFailure, domainproject.ActionManageMembers)
	if err != nil {
		return err
	}

	member, err := s.repo.GetMember(ctx, project.ID, input.UserID)
	if err != nil {
		return flow.memberLookupFailure(ProjectEventRemoveMemberFailure, err)
	}
	flow.setRole(member.Role)

	isSelf := input.CurrentUserID == input.UserID
	if !domainproject.CanRemoveMember(actorRole, member.Role, isSelf) {
		return flow.businessFailure(ProjectEventRemoveMemberFailure, ProjectReasonForbidden, ErrForbidden)
	}
	if member.Role == domainproject.RoleOwner {
		owners, countErr := s.repo.CountOwners(ctx, project.ID)
		if countErr != nil {
			return flow.ownerCountFailure(ProjectEventRemoveMemberFailure, countErr)
		}
		if owners <= 1 {
			return flow.businessFailure(ProjectEventRemoveMemberFailure, ProjectReasonLastOwner, ErrLastOwner)
		}
	}

	if err := s.repo.DeleteMember(ctx, project.ID, input.UserID); err != nil {
		return flow.memberRemoveFailure(err)
	}
	flow.success(ProjectEventRemoveMemberSuccess)

	return nil
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
