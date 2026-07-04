package publicstatus

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (s *Service) loadProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	return s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
}

func (s *Service) loadWritableProject(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	project, err := s.loadProject(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	err = s.requireProjectWrite(ctx, project.ID, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	return project, nil
}

func (s *Service) requireProjectWrite(ctx context.Context, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) loadProjectForFlow(ctx context.Context, flow *publicStatusFlow, projectRef, userID string) (domainproject.Project, error) {
	project, err := s.loadProject(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(err)
	}
	flow.setProject(project)
	return project, nil
}

func (s *Service) requireProjectWriteForFlow(ctx context.Context, flow *publicStatusFlow, projectID, userID string) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return flow.roleLookupFailure(err)
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return flow.businessFailure(PublicStatusReasonForbidden, ErrForbidden)
	}
	return nil
}
