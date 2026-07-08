package alert

import (
	"context"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *Service) ListIncidents(ctx context.Context, input ListIncidentsInput) ([]domainalert.Incident, error) {
	return getAlertList(ctx, s, "alert.incident.list", AlertActionListIncidents, input.ProjectRef, input.CurrentUserID, AlertReasonIncidentListFailed, func(ctx context.Context, projectID string) ([]domainalert.Incident, error) {
		return s.repo.ListIncidents(ctx, projectID, input.Status, input.Limit)
	})
}

func (s *Service) GetIncident(ctx context.Context, input GetIncidentInput) (domainalert.Incident, error) {
	return getAlertResource(ctx, s, "alert.incident.get", AlertActionGetIncident, input.ProjectRef, input.CurrentUserID, attrAlertIncidentID.String(input.IncidentID), AlertReasonIncidentLookupFailed, func(ctx context.Context, projectID string) (domainalert.Incident, error) {
		return s.repo.GetIncident(ctx, projectID, input.IncidentID)
	})
}
