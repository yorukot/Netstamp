package label

import (
	"context"
	"errors"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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

func (s *Service) ListLabels(ctx context.Context, input ListLabelsInput) ([]domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.list", LabelActionList, input.CurrentUserID)
	defer flow.end()
	projectRef, err := normalizeLabelProjectRef(input.ProjectRef)
	if err != nil {
		return nil, flow.businessFailure(LabelEventListFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)

	project, err := s.loadProject(ctx, flow, projectRef, input.CurrentUserID, LabelEventListFailure)
	if err != nil {
		return nil, err
	}

	labels, err := s.repo.ListLabels(ctx, project.ID)
	if err != nil {
		return nil, flow.technicalFailure(LabelEventListFailure, LabelReasonLabelListFailed, err)
	}

	return labels, nil
}

func (s *Service) CreateLabel(ctx context.Context, input CreateLabelInput) (domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.create", LabelActionCreate, input.CurrentUserID)
	defer flow.end()
	normalized, err := normalizeCreateLabelInput(input)
	if err != nil {
		return domainlabel.Label{}, flow.businessFailure(LabelEventCreateFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, LabelEventCreateFailure)
	if err != nil {
		return domainlabel.Label{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, LabelEventCreateFailure, domainproject.ActionManageLabels)
	if err != nil {
		return domainlabel.Label{}, err
	}

	label, err := s.repo.CreateLabel(ctx, domainlabel.CreateLabelStorageInput{
		ProjectID: project.ID,
		Key:       normalized.key,
		Value:     normalized.value,
	})
	if err != nil {
		return domainlabel.Label{}, flow.writeFailure(LabelEventCreateFailure, LabelReasonLabelCreateFailed, err)
	}
	flow.setLabelID(label.ID)
	flow.success(LabelEventCreateSuccess)

	return label, nil
}

func (s *Service) UpdateLabel(ctx context.Context, input UpdateLabelInput) (domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.update", LabelActionUpdate, input.CurrentUserID)
	defer flow.end()
	normalized, err := normalizeUpdateLabelInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setLabelID(input.LabelID)
		return domainlabel.Label{}, flow.businessFailure(LabelEventUpdateFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setLabelID(normalized.labelID)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, LabelEventUpdateFailure)
	if err != nil {
		return domainlabel.Label{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, LabelEventUpdateFailure, domainproject.ActionManageLabels)
	if err != nil {
		return domainlabel.Label{}, err
	}
	if normalized.key == nil && normalized.value == nil {
		return domainlabel.Label{}, flow.businessFailure(
			LabelEventUpdateFailure,
			LabelReasonInvalidInput,
			invalidLabelField("", "at least one field must be provided", nil),
		)
	}
	current, err := s.repo.GetLabel(ctx, project.ID, normalized.labelID)
	if err != nil {
		return domainlabel.Label{}, flow.lookupFailure(LabelEventUpdateFailure, err)
	}

	key := current.Key
	if normalized.key != nil {
		key = *normalized.key
	}
	value := current.Value
	if normalized.value != nil {
		value = *normalized.value
	}

	label, err := s.repo.UpdateLabel(ctx, domainlabel.UpdateLabelStorageInput{
		ProjectID: project.ID,
		LabelID:   normalized.labelID,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		return domainlabel.Label{}, flow.writeFailure(LabelEventUpdateFailure, LabelReasonLabelUpdateFailed, err)
	}
	flow.setLabelID(label.ID)
	flow.success(LabelEventUpdateSuccess)

	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, input DeleteLabelInput) error {
	ctx, flow := s.startLabelFlow(ctx, "label.delete", LabelActionDelete, input.CurrentUserID)
	defer flow.end()
	projectRef, err := normalizeLabelProjectRef(input.ProjectRef)
	if err != nil {
		return flow.businessFailure(LabelEventDeleteFailure, LabelReasonInvalidInput, err)
	}
	labelID, err := normalizeLabelID(input.LabelID)
	if err != nil {
		return flow.businessFailure(LabelEventDeleteFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)
	flow.setLabelID(labelID)

	project, err := s.loadProject(ctx, flow, projectRef, input.CurrentUserID, LabelEventDeleteFailure)
	if err != nil {
		return err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, LabelEventDeleteFailure, domainproject.ActionManageLabels); err != nil {
		return err
	}

	if err := s.repo.SoftDeleteLabel(ctx, project.ID, labelID); err != nil {
		return flow.writeFailure(LabelEventDeleteFailure, LabelReasonLabelDeleteFailed, err)
	}
	flow.success(LabelEventDeleteSuccess)

	return nil
}

func (s *Service) GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.resolve", LabelActionResolve, "")
	defer flow.end()
	flow.setProjectID(projectID)

	normalizedLabelIDs, err := normalizeResolveLabelIDs(labelIDs)
	if err != nil {
		return nil, flow.resolveFailure(err)
	}

	labels, err := s.repo.GetActiveLabelsByIDsForProject(ctx, projectID, normalizedLabelIDs)
	if err != nil {
		return nil, flow.resolveFailure(err)
	}

	return labels, nil
}

func (s *Service) loadProject(ctx context.Context, flow *labelFlow, projectRef, userID string, failureEvent LabelEventName) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if errors.Is(err, ErrProjectNotFound) {
		return domainproject.Project{}, flow.businessFailure(failureEvent, LabelReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return domainproject.Project{}, flow.businessFailure(failureEvent, LabelReasonUserNotFound, err)
	}
	if err != nil {
		return domainproject.Project{}, flow.technicalFailure(failureEvent, LabelReasonProjectLookupFailed, err)
	}
	flow.setProject(project)

	return project, nil
}

func (s *Service) requireAction(ctx context.Context, flow *labelFlow, projectID, userID string, failureEvent LabelEventName, action domainproject.Action) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if errors.Is(err, ErrProjectNotFound) {
		return flow.businessFailure(failureEvent, LabelReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrUserNotFound) {
		return flow.businessFailure(failureEvent, LabelReasonUserNotFound, err)
	}
	if err != nil {
		return flow.technicalFailure(failureEvent, LabelReasonRoleLookupFailed, err)
	}
	if !domainproject.Can(role, action) {
		return flow.businessFailure(failureEvent, LabelReasonForbidden, ErrForbidden)
	}

	return nil
}
