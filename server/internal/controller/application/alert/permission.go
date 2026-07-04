package alert

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (s *Service) loadProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	return s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
}

func (s *Service) requireProjectAction(ctx context.Context, projectID, userID string, action domainproject.Action) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !domainproject.Can(role, action) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireNotificationWrite(ctx context.Context, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if role != domainproject.RoleOwner && role != domainproject.RoleAdmin {
		return ErrForbidden
	}
	return nil
}

func (s *Service) loadProjectForFlow(ctx context.Context, flow *alertFlow, projectRef, userID string) (domainproject.Project, error) {
	project, err := s.loadProject(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(err)
	}
	flow.setProject(project)
	return project, nil
}

func (s *Service) requireProjectActionForFlow(ctx context.Context, flow *alertFlow, projectID, userID string, action domainproject.Action) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return flow.roleLookupFailure(err)
	}
	if !domainproject.Can(role, action) {
		return flow.businessFailure(AlertReasonForbidden, ErrForbidden)
	}
	return nil
}

func (s *Service) requireNotificationWriteForFlow(ctx context.Context, flow *alertFlow, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return flow.roleLookupFailure(err)
	}
	if role != domainproject.RoleOwner && role != domainproject.RoleAdmin {
		return flow.businessFailure(AlertReasonForbidden, ErrForbidden)
	}
	return nil
}
