package assignment

import (
	"context"
	"math"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	tx            Transactor
	events        EventRecorder
}

type noopTransactor struct{}

func (noopTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, transactors ...Transactor) *Service {
	tx := Transactor(noopTransactor{})
	if len(transactors) > 0 && transactors[0] != nil {
		tx = transactors[0]
	}

	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		tx:            tx,
		events:        events,
	}
}

type serviceRefreshRunner struct {
	service *Service
}

func NewWorkerRefreshRunner(service *Service) RefreshRunner {
	if service == nil {
		return nil
	}
	return serviceRefreshRunner{service: service}
}

func (r serviceRefreshRunner) RefreshProbeCheckAssignmentsForProject(ctx context.Context, projectID string) error {
	return r.service.refreshProbeCheckAssignmentsForProject(ctx, projectID, false)
}

func (r serviceRefreshRunner) RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return r.service.refreshProbeCheckAssignmentsForProbe(ctx, projectID, probeID, false)
}

func (r serviceRefreshRunner) RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return r.service.refreshProbeCheckAssignmentsForCheck(ctx, projectID, checkID, false)
}

func (r serviceRefreshRunner) RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error {
	return r.service.refreshProbeCheckAssignmentsForLabel(ctx, projectID, labelID, false)
}

func (r serviceRefreshRunner) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return r.service.deleteProbeCheckAssignmentsForProbe(ctx, projectID, probeID, false)
}

func (r serviceRefreshRunner) DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return r.service.deleteProbeCheckAssignmentsForCheck(ctx, projectID, checkID, false)
}

func (s *Service) RefreshProbeCheckAssignmentsForProject(ctx context.Context, projectID string) error {
	return s.refreshProbeCheckAssignmentsForProject(ctx, projectID, true)
}

func (s *Service) refreshProbeCheckAssignmentsForProject(ctx context.Context, projectID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.project.refresh", AssignmentActionRefreshProject)
	defer flow.end()

	projectID, err := normalizeProjectTarget(projectID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshProjectFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetProject,
			TargetID:  projectID,
		}); err != nil {
			return flow.refreshFailure(AssignmentEventRefreshProjectFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		probes, err := s.repo.ListProbeRefreshCandidatesForProject(ctx, projectID)
		if err != nil {
			return err
		}
		checks, err := s.repo.ListCheckRefreshCandidatesForProject(ctx, projectID)
		if err != nil {
			return err
		}
		for _, probe := range probes {
			if err := s.refreshProbeCandidate(ctx, probe, checks); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshProjectFailure, err)
	}
	flow.success(AssignmentEventRefreshProjectSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return s.refreshProbeCheckAssignmentsForProbe(ctx, projectID, probeID, true)
}

func (s *Service) refreshProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.probe.refresh", AssignmentActionRefreshProbe)
	defer flow.end()

	projectID, probeID, err := normalizeProbeTarget(projectID, probeID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshProbeFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setProbeID(probeID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetProbe,
			TargetID:  probeID,
		}); err != nil {
			return flow.refreshFailure(AssignmentEventRefreshProbeFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		probe, err := s.repo.GetProbeRefreshCandidate(ctx, projectID, probeID)
		if err != nil {
			return err
		}
		checks, err := s.repo.ListCheckRefreshCandidatesForProject(ctx, projectID)
		if err != nil {
			return err
		}
		return s.refreshProbeCandidate(ctx, probe, checks)
	}); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshProbeFailure, err)
	}
	flow.success(AssignmentEventRefreshProbeSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return s.refreshProbeCheckAssignmentsForCheck(ctx, projectID, checkID, true)
}

func (s *Service) refreshProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.check.refresh", AssignmentActionRefreshCheck)
	defer flow.end()

	projectID, checkID, err := normalizeCheckTarget(projectID, checkID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshCheckFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setCheckID(checkID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetCheck,
			TargetID:  checkID,
		}); err != nil {
			return flow.refreshFailure(AssignmentEventRefreshCheckFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		check, err := s.repo.GetCheckRefreshCandidate(ctx, projectID, checkID)
		if err != nil {
			return err
		}
		probes, err := s.repo.ListSelectorPreviewCandidates(ctx, projectID)
		if err != nil {
			return err
		}
		return s.refreshCheckCandidate(ctx, check, probes)
	}); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshCheckFailure, err)
	}
	flow.success(AssignmentEventRefreshCheckSuccess)

	return nil
}

func (s *Service) RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string) error {
	return s.refreshProbeCheckAssignmentsForLabel(ctx, projectID, labelID, true)
}

func (s *Service) refreshProbeCheckAssignmentsForLabel(ctx context.Context, projectID, labelID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.label.refresh", AssignmentActionRefreshLabel)
	defer flow.end()

	projectID, labelID, err := normalizeLabelTarget(projectID, labelID)
	if err != nil {
		return flow.businessFailure(AssignmentEventRefreshLabelFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setLabelID(labelID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetLabel,
			TargetID:  labelID,
		}); err != nil {
			return flow.refreshFailure(AssignmentEventRefreshLabelFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		probes, err := s.repo.ListProbeRefreshCandidatesForLabel(ctx, projectID, labelID)
		if err != nil {
			return err
		}
		checks, err := s.repo.ListCheckRefreshCandidatesForProject(ctx, projectID)
		if err != nil {
			return err
		}
		for _, probe := range probes {
			if err := s.refreshProbeCandidate(ctx, probe, checks); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return flow.refreshFailure(AssignmentEventRefreshLabelFailure, err)
	}
	flow.success(AssignmentEventRefreshLabelSuccess)

	return nil
}

func (s *Service) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string) error {
	return s.deleteProbeCheckAssignmentsForProbe(ctx, projectID, probeID, true)
}

func (s *Service) deleteProbeCheckAssignmentsForProbe(ctx context.Context, projectID, probeID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.probe.delete", AssignmentActionDeleteProbe)
	defer flow.end()

	projectID, probeID, err := normalizeProbeTarget(projectID, probeID)
	if err != nil {
		return flow.businessFailure(AssignmentEventDeleteProbeFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setProbeID(probeID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetProbe,
			TargetID:  probeID,
		}); err != nil {
			return flow.deleteFailure(AssignmentEventDeleteProbeFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		return s.repo.DeleteProbeCheckAssignmentsForProbe(ctx, projectID, probeID)
	}); err != nil {
		return flow.deleteFailure(AssignmentEventDeleteProbeFailure, err)
	}
	flow.success(AssignmentEventDeleteProbeSuccess)

	return nil
}

func (s *Service) DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string) error {
	return s.deleteProbeCheckAssignmentsForCheck(ctx, projectID, checkID, true)
}

func (s *Service) deleteProbeCheckAssignmentsForCheck(ctx context.Context, projectID, checkID string, enqueue bool) error {
	ctx, flow := s.startAssignmentFlow(ctx, "assignment.check.delete", AssignmentActionDeleteCheck)
	defer flow.end()

	projectID, checkID, err := normalizeCheckTarget(projectID, checkID)
	if err != nil {
		return flow.businessFailure(AssignmentEventDeleteCheckFailure, AssignmentReasonInvalidInput, err)
	}
	flow.setProjectID(projectID)
	flow.setCheckID(checkID)

	if enqueue {
		if err := s.enqueueRefreshJob(ctx, domainassignment.RefreshTarget{
			ProjectID: projectID,
			Type:      domainassignment.RefreshTargetCheck,
			TargetID:  checkID,
		}); err != nil {
			return flow.deleteFailure(AssignmentEventDeleteCheckFailure, err)
		}
	}
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		return s.repo.DeleteProbeCheckAssignmentsForCheck(ctx, projectID, checkID)
	}); err != nil {
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

	candidates, err := s.repo.ListSelectorPreviewCandidates(ctx, project.ID)
	if err != nil {
		return SelectorPreviewOutput{}, flow.technicalFailure(AssignmentEventPreviewFailure, AssignmentReasonPreviewFailed, err)
	}
	probes := matchingPreviewProbes(normalized.selector, candidates)

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

func (s *Service) refreshProbeCandidate(ctx context.Context, probe ProbeAssignmentCandidate, checks []CheckAssignmentCandidate) error {
	if !probe.Enabled {
		return s.repo.DeleteProbeCheckAssignmentsForProbe(ctx, probe.ProjectID, probe.ProbeID)
	}

	keepCheckIDs := make([]string, 0, len(checks))
	for _, check := range checks {
		if !check.Selector.Matches(probe.Labels) {
			continue
		}
		keepCheckIDs = append(keepCheckIDs, check.Check.ID)
		if err := s.repo.UpsertProbeCheckAssignment(ctx, AssignmentWrite{
			ProjectID:       probe.ProjectID,
			ProbeID:         probe.ProbeID,
			CheckID:         check.Check.ID,
			CheckVersion:    check.CheckVersion,
			SelectorVersion: check.SelectorVersion,
		}); err != nil {
			return err
		}
	}

	return s.repo.DeleteStaleAssignmentsForProbe(ctx, probe.ProjectID, probe.ProbeID, keepCheckIDs)
}

func (s *Service) refreshCheckCandidate(ctx context.Context, check CheckAssignmentCandidate, probes []ProbeAssignmentCandidate) error {
	keepProbeIDs := make([]string, 0, len(probes))
	for _, probe := range probes {
		if !probe.Enabled || !check.Selector.Matches(probe.Labels) {
			continue
		}
		keepProbeIDs = append(keepProbeIDs, probe.ProbeID)
		if err := s.repo.UpsertProbeCheckAssignment(ctx, AssignmentWrite{
			ProjectID:       check.Check.ProjectID,
			ProbeID:         probe.ProbeID,
			CheckID:         check.Check.ID,
			CheckVersion:    check.CheckVersion,
			SelectorVersion: check.SelectorVersion,
		}); err != nil {
			return err
		}
	}

	return s.repo.DeleteStaleAssignmentsForCheck(ctx, check.Check.ProjectID, check.Check.ID, check.CheckVersion, check.SelectorVersion, keepProbeIDs)
}

func matchingPreviewProbes(selector domainselector.Selector, probes []ProbeAssignmentCandidate) []domainprobe.Probe {
	matches := make([]domainprobe.Probe, 0, len(probes))
	for _, probe := range probes {
		if !probe.Enabled || !selector.Matches(probe.Labels) {
			continue
		}
		matches = append(matches, domainprobe.Probe{
			ID:        probe.ProbeID,
			ProjectID: probe.ProjectID,
			Name:      probe.Name,
			Enabled:   probe.Enabled,
			Labels:    append([]domainlabel.Label(nil), probe.Labels...),
		})
	}
	return matches
}
