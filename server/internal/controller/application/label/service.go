package label

import (
	"context"
	"errors"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo                Repository
	projectAccess       ProjectAccess
	events              EventRecorder
	assignmentRefresher AssignmentRefresher
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, assignmentRefresher AssignmentRefresher) *Service {

	return &Service{
		repo:                repo,
		projectAccess:       projectAccess,
		events:              events,
		assignmentRefresher: assignmentRefresher,
	}
}

func (s *Service) ListLabels(ctx context.Context, input ListLabelsInput) ([]domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.list", LabelActionList, input.CurrentUserID)
	defer flow.end()
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
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
	input, err := normalizeCreateLabelInput(input)
	if err != nil {
		return domainlabel.Label{}, flow.businessFailure(LabelEventCreateFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, LabelEventCreateFailure)
	if err != nil {
		return domainlabel.Label{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, LabelEventCreateFailure, domainproject.ActionManageLabels)
	if err != nil {
		return domainlabel.Label{}, err
	}

	label, err := s.repo.CreateLabel(ctx, domainlabel.Label{
		ProjectID: project.ID,
		Key:       input.Key,
		Value:     input.Value,
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
	input, err := normalizeUpdateLabelInput(input)
	if err != nil {
		return domainlabel.Label{}, flow.businessFailure(LabelEventUpdateFailure, LabelReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setLabelID(input.LabelID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, LabelEventUpdateFailure)
	if err != nil {
		return domainlabel.Label{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, LabelEventUpdateFailure, domainproject.ActionManageLabels)
	if err != nil {
		return domainlabel.Label{}, err
	}
	current, err := s.repo.GetLabel(ctx, project.ID, input.LabelID)
	if err != nil {
		return domainlabel.Label{}, flow.lookupFailure(LabelEventUpdateFailure, err)
	}

	key := current.Key
	if input.Key != nil {
		key = *input.Key
	}
	value := current.Value
	if input.Value != nil {
		value = *input.Value
	}

	label, err := s.repo.UpdateLabel(ctx, domainlabel.Label{
		ProjectID: project.ID,
		ID:        input.LabelID,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		return domainlabel.Label{}, flow.writeFailure(LabelEventUpdateFailure, LabelReasonLabelUpdateFailed, err)
	}

	// TODO: Once we update the label we need to update all the assigment at the probe and check.

	flow.setLabelID(label.ID)
	flow.success(LabelEventUpdateSuccess)

	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, input DeleteLabelInput) error {
	ctx, flow := s.startLabelFlow(ctx, "label.delete", LabelActionDelete, input.CurrentUserID)
	defer flow.end()
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		err = invalidLabelField("projectRef", err.Error(), input.ProjectRef)
		return flow.businessFailure(LabelEventDeleteFailure, LabelReasonInvalidInput, err)
	}
	labelID, err := domainlabel.VNLabelID(input.LabelID)
	if err != nil {
		err = invalidLabelField("labelId", err.Error(), input.LabelID)
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
	if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForLabel(ctx, project.ID, labelID); err != nil {
		return flow.technicalFailure(LabelEventDeleteFailure, LabelReasonAssignmentRefreshFailed, err)
	}

	// TODO: Once we update the label we need to update all the assigment at the probe and check.

	flow.success(LabelEventDeleteSuccess)

	return nil
}

func (s *Service) GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	ctx, flow := s.startLabelFlow(ctx, "label.resolve", LabelActionResolve, "")
	defer flow.end()
	flow.setProjectID(projectID)

	var normalizedLabelIDs []string
	for _, labelID := range labelIDs {
		normalizedLabelID, err := domainlabel.VNLabelID(labelID)
		if err != nil {
			return nil, flow.resolveFailure(err)
		}
		normalizedLabelIDs = append(normalizedLabelIDs, normalizedLabelID)
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
