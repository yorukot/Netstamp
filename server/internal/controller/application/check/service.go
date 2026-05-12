package check

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo                Repository
	projectAccess       ProjectAccess
	labelAccess         LabelAccess
	assignmentRefresher AssignmentRefresher
	events              EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, assignmentRefresher AssignmentRefresher, events EventRecorder) *Service {
	return &Service{
		repo:                repo,
		projectAccess:       projectAccess,
		labelAccess:         labelAccess,
		assignmentRefresher: assignmentRefresher,
		events:              events,
	}
}

func (s *Service) ListChecks(ctx context.Context, input ListChecksInput) ([]domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.list", CheckActionList, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeListChecksInput(input)
	if err != nil {
		return nil, flow.businessFailure(CheckEventListFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventListFailure)
	if err != nil {
		return nil, err
	}

	checks, err := s.repo.ListChecks(ctx, project.ID)
	if err != nil {
		return nil, flow.technicalFailure(CheckEventListFailure, CheckReasonCheckListFailed, err)
	}

	return checks, nil
}

func (s *Service) GetCheck(ctx context.Context, input GetCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.get", CheckActionGet, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventGetFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventGetFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	check, err := s.repo.GetCheck(ctx, project.ID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventGetFailure, err)
	}

	return check, nil
}

func (s *Service) CreateCheck(ctx context.Context, input CreateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.create", CheckActionCreate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeCreateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventCreateFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, CheckEventCreateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventCreateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, normalized.labelIDs)
	if err != nil {
		return domaincheck.Check{}, flow.labelLookupFailure(CheckEventCreateFailure, err)
	}

	checkInput := domaincheck.Check{
		ProjectID:       project.ID,
		Name:            normalized.name,
		Type:            normalized.checkType,
		Target:          normalized.target,
		Selector:        normalized.selector,
		Description:     normalized.description,
		IntervalSeconds: normalized.intervalSeconds,
		PingConfig:      &normalized.pingConfig,
	}
	check, err := s.repo.CreateCheck(ctx, checkInput, normalized.labelIDs)
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventCreateFailure, CheckReasonCheckCreateFailed, err)
	}
	check.Labels = labels
	flow.setCheckID(check.ID)
	if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForCheck(ctx, project.ID, check.ID); err != nil {
		return domaincheck.Check{}, flow.assignmentRefreshFailure(CheckEventCreateFailure, err)
	}
	flow.success(CheckEventCreateSuccess)

	return check, nil
}

func (s *Service) UpdateCheck(ctx context.Context, input UpdateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.update", CheckActionUpdate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeUpdateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setCheckID(normalized.checkID)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	current, err := s.repo.GetCheck(ctx, project.ID, normalized.checkID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventUpdateFailure, err)
	}

	labelIDs, resolvedLabels, err := s.resolveUpdateLabels(ctx, project.ID, current.Labels, normalized)
	if err != nil {
		return domaincheck.Check{}, flow.labelLookupFailure(CheckEventUpdateFailure, err)
	}

	updated, err := buildUpdatedCheck(project.ID, current, normalized)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	check, err := s.repo.UpdateCheck(ctx, updated, normalized.replaceLabels, labelIDs)
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventUpdateFailure, CheckReasonCheckUpdateFailed, err)
	}
	if normalized.replaceLabels {
		check.Labels = resolvedLabels
	}
	if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForCheck(ctx, project.ID, check.ID); err != nil {
		return domaincheck.Check{}, flow.assignmentRefreshFailure(CheckEventUpdateFailure, err)
	}
	flow.success(CheckEventUpdateSuccess)

	return check, nil
}

func (s *Service) resolveUpdateLabels(ctx context.Context, projectID string, currentLabels []domainlabel.Label, normalized normalizedUpdateCheckInput) ([]string, []domainlabel.Label, error) {
	if !normalized.replaceLabels {
		return nil, currentLabels, nil
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, projectID, normalized.labelIDs)
	if err != nil {
		return nil, nil, err
	}

	return normalized.labelIDs, labels, nil
}

func buildUpdatedCheck(projectID string, current domaincheck.Check, normalized normalizedUpdateCheckInput) (domaincheck.Check, error) {
	selector := current.Selector
	if normalized.selector != nil {
		selector = normalized.selector
	}
	updatedSelector, err := canonicalizeSelector(selector)
	if err != nil {
		return domaincheck.Check{}, err
	}

	return domaincheck.Check{
		ProjectID:       projectID,
		ID:              normalized.checkID,
		Name:            mergedString(current.Name, normalized.name),
		Type:            mergedCheckType(current.Type, normalized.checkType),
		Target:          mergedString(current.Target, normalized.target),
		Selector:        updatedSelector,
		Description:     mergedDescription(current.Description, normalized.description),
		IntervalSeconds: mergedInt32(current.IntervalSeconds, normalized.intervalSeconds),
		PingConfig:      mergedPingConfig(current.PingConfig, normalized),
	}, nil
}

func mergedPingConfig(current *domainping.Config, normalized normalizedUpdateCheckInput) *domainping.Config {
	config := domainping.DefaultConfig()
	if current != nil {
		config = *current
	}
	if normalized.packetCount != nil {
		config.PacketCount = *normalized.packetCount
	}
	if normalized.packetSizeBytes != nil {
		config.PacketSizeBytes = *normalized.packetSizeBytes
	}
	if normalized.timeoutMs != nil {
		config.TimeoutMs = *normalized.timeoutMs
	}
	if normalized.ipFamily != nil {
		config.IPFamily = normalized.ipFamily
	}
	return &config
}

func mergedString(current string, next *string) string {
	if next == nil {
		return current
	}
	return *next
}

func mergedCheckType(current domaincheck.Type, next *domaincheck.Type) domaincheck.Type {
	if next == nil {
		return current
	}
	return *next
}

func mergedDescription(current, next *string) *string {
	if next == nil {
		return current
	}
	return next
}

func mergedInt32(current int32, next *int32) int32 {
	if next == nil {
		return current
	}
	return *next
}

func (s *Service) DeleteCheck(ctx context.Context, input GetCheckInput) error {
	ctx, flow := s.startCheckFlow(ctx, "check.delete", CheckActionDelete, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetCheckInput(input)
	if err != nil {
		return flow.businessFailure(CheckEventDeleteFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventDeleteFailure)
	if err != nil {
		return err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventDeleteFailure); err != nil {
		return err
	}

	if err := s.repo.SoftDeleteCheck(ctx, project.ID, input.CheckID); err != nil {
		return flow.writeFailure(CheckEventDeleteFailure, CheckReasonCheckDeleteFailed, err)
	}
	if err := s.assignmentRefresher.DeleteProbeCheckAssignmentsForCheck(ctx, project.ID, input.CheckID); err != nil {
		return flow.assignmentDeleteFailure(CheckEventDeleteFailure, err)
	}
	flow.success(CheckEventDeleteSuccess)

	return nil
}

func (s *Service) loadProject(ctx context.Context, flow *checkFlow, projectRef, userID string, failureEvent CheckEventName) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(failureEvent, err)
	}
	flow.setProject(project)

	return project, nil
}

func (s *Service) requireAction(ctx context.Context, flow *checkFlow, projectID, userID string, failureEvent CheckEventName) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return flow.roleLookupFailure(failureEvent, err)
	}
	if !domainproject.Can(role, domainproject.ActionManageChecks) {
		return flow.businessFailure(failureEvent, CheckReasonForbidden, ErrForbidden)
	}

	return nil
}
