package check

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	labelAccess   LabelAccess
	events        EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, events EventRecorder) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		labelAccess:   labelAccess,
		events:        events,
	}
}

func (s *Service) ListChecks(ctx context.Context, input ListChecksInput) ([]domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.list", CheckActionList, input.CurrentUserID)
	defer flow.end()
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
	flow.setProjectRef(input.ProjectRef)

	normalized, err := normalizeCreateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventCreateFailure, CheckReasonInvalidInput, err)
	}

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventCreateFailure)
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

	check, err := s.repo.CreateCheck(ctx, domaincheck.CreateCheckStorageInput{
		ProjectID:       project.ID,
		Name:            normalized.name,
		Type:            normalized.checkType,
		Target:          normalized.target,
		Selector:        normalized.selector,
		CheckVersion:    domaincheck.CheckVersion(normalized.executionSpec()),
		SelectorVersion: domaincheck.SelectorVersion(normalized.selector),
		Description:     normalized.description,
		IntervalSeconds: normalized.intervalSeconds,
		PingConfig:      normalized.pingConfig,
		LabelIDs:        normalized.labelIDs,
	})
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventCreateFailure, CheckReasonCheckCreateFailed, err)
	}
	check.Labels = labels
	flow.setCheckID(check.ID)
	flow.success(CheckEventCreateSuccess)

	return check, nil
}

func (s *Service) UpdateCheck(ctx context.Context, input UpdateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.update", CheckActionUpdate, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	normalized, err := normalizeUpdateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	current, err := s.repo.GetCheck(ctx, project.ID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventUpdateFailure, err)
	}

	labelIDs := []string(nil)
	resolvedLabels := current.Labels
	if normalized.replaceLabels {
		labelIDs = normalized.labelIDs
		resolvedLabels, err = s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, labelIDs)
		if err != nil {
			return domaincheck.Check{}, flow.labelLookupFailure(CheckEventUpdateFailure, err)
		}
	}

	updatedPingConfig := domainping.Config{
		PacketCount:     chooseInt32(current.PingConfig.PacketCount, normalized.packetCount),
		PacketSizeBytes: chooseInt32(current.PingConfig.PacketSizeBytes, normalized.packetSizeBytes),
		TimeoutMs:       chooseInt32(current.PingConfig.TimeoutMs, normalized.timeoutMs),
		IPFamily:        chooseIPFamily(current.PingConfig.IPFamily, normalized.ipFamily),
	}
	updatedSelector, err := canonicalizeSelector(chooseRawMessage(current.Selector, normalized.selector))
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	updated := domaincheck.UpdateCheckStorageInput{
		ProjectID:       project.ID,
		CheckID:         input.CheckID,
		Name:            chooseString(current.Name, normalized.name),
		Type:            chooseCheckType(current.Type, normalized.checkType),
		Target:          chooseString(current.Target, normalized.target),
		Selector:        updatedSelector,
		Description:     chooseOptionalString(current.Description, normalized.description),
		IntervalSeconds: chooseInt32(current.IntervalSeconds, normalized.intervalSeconds),
		PingConfig:      updatedPingConfig,
		ReplaceLabels:   normalized.replaceLabels,
		LabelIDs:        labelIDs,
	}
	updated.CheckVersion = domaincheck.CheckVersion(domaincheck.ExecutionSpec{
		Type:            updated.Type,
		Target:          updated.Target,
		IntervalSeconds: updated.IntervalSeconds,
		PingConfig:      updated.PingConfig,
	})
	updated.SelectorVersion = domaincheck.SelectorVersion(updated.Selector)

	check, err := s.repo.UpdateCheck(ctx, updated)
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventUpdateFailure, CheckReasonCheckUpdateFailed, err)
	}
	if normalized.replaceLabels {
		check.Labels = resolvedLabels
	}
	flow.success(CheckEventUpdateSuccess)

	return check, nil
}

func (s *Service) DeleteCheck(ctx context.Context, input GetCheckInput) error {
	ctx, flow := s.startCheckFlow(ctx, "check.delete", CheckActionDelete, input.CurrentUserID)
	defer flow.end()
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
