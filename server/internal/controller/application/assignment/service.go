package assignment

import "context"

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

func (s *Service) RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.probe.refresh", AssignmentActionRefreshProbe)
	defer flow.end()

	projectID, probeID, err := normalizeProbeTarget(projectID, probeID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshProbeFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setProbeID(probeID)

	if err := s.repo.RefreshProbeCheckAssignmentsForProbe(ctx, projectID, probeID); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshProbeFailure, err)
	}
	flow.success(AssignmentEventRefreshProbeSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.check.refresh", AssignmentActionRefreshCheck)
	defer flow.end()

	projectID, checkID, err := normalizeCheckTarget(projectID, checkID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshCheckFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setCheckID(checkID)

	if err := s.repo.RefreshProbeCheckAssignmentsForCheck(ctx, projectID, checkID); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshCheckFailure, err)
	}
	flow.success(AssignmentEventRefreshCheckSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.label.refresh", AssignmentActionRefreshLabel)
	defer flow.end()

	projectID, labelID, err := normalizeLabelTarget(projectID, labelID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshLabelFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setLabelID(labelID)

	if err := s.repo.RefreshProbeCheckAssignmentsForLabel(ctx, projectID, labelID); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshLabelFailure, err)
	}
	flow.success(AssignmentEventRefreshLabelSuccess)

	return nil
}

func (s *Service) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.probe.delete", AssignmentActionDeleteProbe)
	defer flow.end()

	projectID, probeID, err := normalizeProbeTarget(projectID, probeID)
	if err != nil {
		return flow.businessFailure(AssignmentEventDeleteProbeFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setProbeID(probeID)

	if err := s.repo.DeleteProbeCheckAssignmentsForProbe(ctx, projectID, probeID); err != nil {
		return flow.deleteFailure(AssignmentEventDeleteProbeFailure, err)
	}
	flow.success(AssignmentEventDeleteProbeSuccess)

	return nil
}

func (s *Service) DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.check.delete", AssignmentActionDeleteCheck)
	defer flow.end()

	projectID, checkID, err := normalizeCheckTarget(projectID, checkID)
	if err != nil {
		return flow.businessFailure(AssignmentEventDeleteCheckFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setCheckID(checkID)

	if err := s.repo.DeleteProbeCheckAssignmentsForCheck(ctx, projectID, checkID); err != nil {
		return flow.deleteFailure(AssignmentEventDeleteCheckFailure, err)
	}
	flow.success(AssignmentEventDeleteCheckSuccess)

	return nil
}
