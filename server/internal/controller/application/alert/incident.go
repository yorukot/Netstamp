package alert

import (
	"context"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *Service) ListIncidents(ctx context.Context, input ListIncidentsInput) ([]domainalert.Incident, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListIncidents(ctx, project.ID, input.Status, input.Limit)
}

func (s *Service) GetIncident(ctx context.Context, input GetIncidentInput) (domainalert.Incident, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Incident{}, err
	}
	return s.repo.GetIncident(ctx, project.ID, input.IncidentID)
}
