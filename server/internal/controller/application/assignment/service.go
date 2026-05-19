package assignment

import (
	"context"
	"math"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	events        EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		events:        events,
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

func (s *Service) PreviewSelector(ctx context.Context, input PreviewSelectorInput) (SelectorPreviewOutput, error) {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.selector_preview", AssignmentActionPreview)
	defer flow.end()

	normalized, err := normalizePreviewSelectorInput(input)
	if err != nil {
		return SelectorPreviewOutput{}, flow.businessFailure(AssignmentEventPreviewFailure, AssignmentReasonInvalidInput, err)
	}

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		return SelectorPreviewOutput{}, flow.projectLookupFailure(AssignmentEventPreviewFailure, err)
	}
	flow.setProjectID(project.ID)

	probes, err := s.repo.ListSelectorPreviewProbes(ctx, project.ID, normalized.selector)
	if err != nil {
		return SelectorPreviewOutput{}, flow.technicalFailure(AssignmentEventPreviewFailure, AssignmentReasonPreviewFailed, err)
	}

	return SelectorPreviewOutput{
		Selector:     normalized.selector.CanonicalJSON(),
		MatchedCount: matchedProbeCount(probes),
		Probes:       probes,
	}, nil
}

func (s *Service) ListProjectAssignments(ctx context.Context, input ListProjectAssignmentsInput) (ProjectAssignmentsOutput, error) {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.list", AssignmentActionList)
	defer flow.end()

	normalized, err := normalizeListProjectAssignmentsInput(input)
	if err != nil {
		return ProjectAssignmentsOutput{}, flow.businessFailure(AssignmentEventListFailure, AssignmentReasonInvalidInput, err)
	}
	if normalized.probeID != "" {
		flow.setProbeID(normalized.probeID)
	}
	if normalized.checkID != "" {
		flow.setCheckID(normalized.checkID)
	}

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		return ProjectAssignmentsOutput{}, flow.projectLookupFailure(AssignmentEventListFailure, err)
	}
	flow.setProjectID(project.ID)

	assignments, err := s.repo.ListProjectAssignments(ctx, domainassignment.Query{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
	})
	if err != nil {
		return ProjectAssignmentsOutput{}, flow.technicalFailure(AssignmentEventListFailure, AssignmentReasonListFailed, err)
	}

	return ProjectAssignmentsOutput{Assignments: assignments}, nil
}

func matchedProbeCount(probes []domainprobe.Probe) int32 {
	if len(probes) > math.MaxInt32 {
		return math.MaxInt32
	}
	//nolint:gosec // len(probes) is bounded above before narrowing to the OpenAPI int32 field.
	return int32(len(probes))
}
