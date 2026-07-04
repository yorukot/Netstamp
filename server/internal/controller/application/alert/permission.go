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
