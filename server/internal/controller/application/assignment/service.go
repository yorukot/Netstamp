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

func (s *Service) RefreshProbeCheckAssignmentsForProject(ctx context.Context, projectID string) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.project.refresh", AssignmentActionRefreshProject)
	defer flow.end()

	projectID, err := normalizeProjectTarget(projectID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshProjectFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)

	if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
		ProjectID: projectID,
		Type:      domainassignment.RefreshTargetProject,
		TargetID:  projectID,
	}); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshProjectFailure, err)
	}
	if err := s.repo.RefreshProbeCheckAssignmentsForProject(ctx, projectID); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshProjectFailure, err)
	}
	flow.success(AssignmentEventRefreshProjectSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return s.runAssignmentTargetMutation(ctx, projectID, probeID, assignmentTargetMutation{
		spanName:     "assignment.probe.refresh",
		action:       AssignmentActionRefreshProbe,
		targetType:   domainassignment.RefreshTargetProbe,
		successEvent: AssignmentEventRefreshProbeSuccess,
		failureEvent: AssignmentEventRefreshProbeFailure,
		normalize:    normalizeProbeTarget,
		setTarget:    (*assignmentFlow).setProbeID,
		run:          s.repo.RefreshProbeCheckAssignmentsForProbe,
		mapFailure:   (*assignmentFlow).refreshFailure,
	})
}

func (s *Service) RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return s.runAssignmentTargetMutation(ctx, projectID, checkID, assignmentTargetMutation{
		spanName:     "assignment.check.refresh",
		action:       AssignmentActionRefreshCheck,
		targetType:   domainassignment.RefreshTargetCheck,
		successEvent: AssignmentEventRefreshCheckSuccess,
		failureEvent: AssignmentEventRefreshCheckFailure,
		normalize:    normalizeCheckTarget,
		setTarget:    (*assignmentFlow).setCheckID,
		run:          s.repo.RefreshProbeCheckAssignmentsForCheck,
		mapFailure:   (*assignmentFlow).refreshFailure,
	})
}

func (s *Service) RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error {
	return s.runAssignmentTargetMutation(ctx, projectID, labelID, assignmentTargetMutation{
		spanName:     "assignment.label.refresh",
		action:       AssignmentActionRefreshLabel,
		targetType:   domainassignment.RefreshTargetLabel,
		successEvent: AssignmentEventRefreshLabelSuccess,
		failureEvent: AssignmentEventRefreshLabelFailure,
		normalize:    normalizeLabelTarget,
		setTarget:    (*assignmentFlow).setLabelID,
		run:          s.repo.RefreshProbeCheckAssignmentsForLabel,
		mapFailure:   (*assignmentFlow).refreshFailure,
	})
}

func (s *Service) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return s.runAssignmentTargetMutation(ctx, projectID, probeID, assignmentTargetMutation{
		spanName:     "assignment.probe.delete",
		action:       AssignmentActionDeleteProbe,
		targetType:   domainassignment.RefreshTargetProbe,
		successEvent: AssignmentEventDeleteProbeSuccess,
		failureEvent: AssignmentEventDeleteProbeFailure,
		normalize:    normalizeProbeTarget,
		setTarget:    (*assignmentFlow).setProbeID,
		run:          s.repo.DeleteProbeCheckAssignmentsForProbe,
		mapFailure:   (*assignmentFlow).deleteFailure,
	})
}

func (s *Service) DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return s.runAssignmentTargetMutation(ctx, projectID, checkID, assignmentTargetMutation{
		spanName:     "assignment.check.delete",
		action:       AssignmentActionDeleteCheck,
		targetType:   domainassignment.RefreshTargetCheck,
		successEvent: AssignmentEventDeleteCheckSuccess,
		failureEvent: AssignmentEventDeleteCheckFailure,
		normalize:    normalizeCheckTarget,
		setTarget:    (*assignmentFlow).setCheckID,
		run:          s.repo.DeleteProbeCheckAssignmentsForCheck,
		mapFailure:   (*assignmentFlow).deleteFailure,
	})
}

type assignmentTargetMutation struct {
	spanName     string
	action       AssignmentEventAction
	targetType   domainassignment.RefreshTargetType
	successEvent AssignmentEventName
	failureEvent AssignmentEventName
	normalize    func(string, string) (string, string, error)
	setTarget    func(*assignmentFlow, string)
	run          func(context.Context, string, string) error
	mapFailure   func(*assignmentFlow, AssignmentEventName, error) error
}

func (s *Service) runAssignmentTargetMutation(ctx context.Context, projectID, targetID string, mutation assignmentTargetMutation) error {
	ctx, flow := s.startAssignmentFlow(ctx, mutation.spanName, mutation.action)
	defer flow.end()

	projectID, targetID, err := mutation.normalize(projectID, targetID)
	if err != nil {
		return flow.businessFailure(mutation.failureEvent, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	mutation.setTarget(flow, targetID)

	if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
		ProjectID: projectID,
		Type:      mutation.targetType,
		TargetID:  targetID,
	}); err != nil {
		return mutation.mapFailure(flow, mutation.failureEvent, err)
	}
	if err := mutation.run(ctx, projectID, targetID); err != nil {
		return mutation.mapFailure(flow, mutation.failureEvent, err)
	}
	flow.success(mutation.successEvent)

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

func (s *Service) enqueueRefreshJob(ctx context.Context, target domainassignment.RefreshTarget) error {
	return s.repo.EnqueueRefreshJob(ctx, target, domainassignment.DefaultRefreshJobMaxAttempts)
}
