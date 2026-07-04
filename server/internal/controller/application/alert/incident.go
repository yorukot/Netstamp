package alert

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (s *Service) ListIncidents(ctx context.Context, input ListIncidentsInput) ([]domainalert.Incident, error) {
	ctx, span := alertTracer.Start(ctx, "alert.incident.list", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionListIncidents)),
		attrProjectRef.String(input.ProjectRef),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	incidents, err := s.repo.ListIncidents(ctx, project.ID, input.Status, input.Limit)
	if err != nil {
		return nil, recordAlertQueryFailure(span, AlertReasonIncidentListFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return incidents, nil
}

func (s *Service) GetIncident(ctx context.Context, input GetIncidentInput) (domainalert.Incident, error) {
	ctx, span := alertTracer.Start(ctx, "alert.incident.get", trace.WithAttributes(
		attrAlertAction.String(string(AlertActionGetIncident)),
		attrProjectRef.String(input.ProjectRef),
		attrAlertIncidentID.String(input.IncidentID),
	))
	defer span.End()

	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainalert.Incident{}, recordAlertQueryFailure(span, AlertReasonProjectLookupFailed, err)
	}
	span.SetAttributes(attrProjectID.String(project.ID))
	incident, err := s.repo.GetIncident(ctx, project.ID, input.IncidentID)
	if err != nil {
		return domainalert.Incident{}, recordAlertQueryFailure(span, AlertReasonIncidentLookupFailed, err)
	}
	span.SetAttributes(attrAlertOutcome.String(string(AlertOutcomeSuccess)))
	return incident, nil
}
