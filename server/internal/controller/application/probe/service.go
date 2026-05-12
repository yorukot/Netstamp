package probe

import (
	"context"
	"errors"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo                Repository
	projectAccess       ProjectAccess
	labelAccess         LabelAccess
	assignmentRefresher AssignmentRefresher
	secretGenerator     SecretGenerator
	events              EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, assignmentRefresher AssignmentRefresher, secretGenerator SecretGenerator, events EventRecorder) *Service {
	return &Service{
		repo:                repo,
		projectAccess:       projectAccess,
		labelAccess:         labelAccess,
		assignmentRefresher: assignmentRefresher,
		secretGenerator:     secretGenerator,
		events:              events,
	}
}

func (s *Service) CreateProbe(ctx context.Context, input CreateProbeInput) (CreateProbeOutput, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.create", ProbeActionCreate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeCreateProbeInput(input)
	if err != nil {
		return CreateProbeOutput{}, flow.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return CreateProbeOutput{}, flow.projectLookupFailure(ProbeEventCreateFailure, err)
	}
	flow.setProjectID(project.ID)

	role, err := s.projectAccess.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return CreateProbeOutput{}, flow.roleLookupFailure(ProbeEventCreateFailure, err)
	}
	if !domainproject.Can(role, domainproject.ActionManageProbes) {
		return CreateProbeOutput{}, flow.businessFailure(ProbeEventCreateFailure, ProbeReasonForbidden, ErrForbidden)
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, input.LabelIDs)
	if err != nil {
		return CreateProbeOutput{}, flow.labelLookupFailure(ProbeEventCreateFailure, err)
	}

	if s.secretGenerator == nil {
		return CreateProbeOutput{}, flow.technicalFailure(
			ProbeEventCreateFailure,
			ProbeReasonSecretGeneratorMissing,
			errors.New("probe secret generator is not configured"),
		)
	}
	plaintextSecret, secretHash, err := s.secretGenerator.GenerateProbeSecret()
	if err != nil {
		return CreateProbeOutput{}, flow.technicalFailure(ProbeEventCreateFailure, ProbeReasonSecretGenerateFailed, err)
	}

	probe, err := s.repo.CreateProbe(ctx, domainprobe.Probe{
		ProjectID:       project.ID,
		Name:            input.Name,
		Enabled:         *input.Enabled,
		SubdivisionCode: input.SubdivisionCode,
		Latitude:        input.Latitude,
		Longitude:       input.Longitude,
		Labels:          labels,
	}, secretHash)
	if err != nil {
		return CreateProbeOutput{}, flow.createFailure(err)
	}
	flow.setProbeID(probe.ID)
	probe.Labels = labels

	if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForProbe(ctx, project.ID, probe.ID); err != nil {
		return CreateProbeOutput{}, flow.assignmentRefreshFailure(ProbeEventCreateFailure, err)
	}

	return CreateProbeOutput{
		Probe:  probe,
		Secret: plaintextSecret,
	}, nil
}

func (s *Service) ListProbes(ctx context.Context, input ListProbesInput) ([]domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.list", ProbeActionList, input.CurrentUserID)
	defer flow.end()

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		inputErr := invalidProbeField("projectRef", err.Error(), input.ProjectRef)
		return nil, flow.businessFailure(ProbeEventListFailure, ProbeReasonInvalidInput, inputErr)
	}
	flow.setProjectRef(projectRef)

	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return nil, flow.projectLookupFailure(ProbeEventListFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventListFailure, domainproject.ActionReadProject); actionErr != nil {
		return nil, actionErr
	}

	probes, err := s.repo.ListProbesForProject(ctx, project.ID)
	if err != nil {
		return nil, flow.probeListFailure(err)
	}

	return probes, nil
}

func (s *Service) GetProbe(ctx context.Context, input TargetProbeInput) (domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.get", ProbeActionGet, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetProbeInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setProbeID(input.ProbeID)
		return domainprobe.Probe{}, flow.businessFailure(ProbeEventGetFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setProbeID(input.ProbeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainprobe.Probe{}, flow.projectLookupFailure(ProbeEventGetFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventGetFailure, domainproject.ActionReadProject); actionErr != nil {
		return domainprobe.Probe{}, actionErr
	}

	probe, err := s.repo.GetProbeForProject(ctx, project.ID, input.ProbeID)
	if err != nil {
		return domainprobe.Probe{}, flow.probeLookupFailure(ProbeEventGetFailure, err)
	}

	return probe, nil
}

func (s *Service) UpdateProbe(ctx context.Context, input UpdateProbeInput) (domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.update", ProbeActionUpdate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeUpdateProbeInput(input)
	if err != nil {
		return domainprobe.Probe{}, flow.businessFailure(ProbeEventUpdateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setProbeID(input.ProbeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainprobe.Probe{}, flow.projectLookupFailure(ProbeEventUpdateFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventUpdateFailure, domainproject.ActionManageProbes); actionErr != nil {
		return domainprobe.Probe{}, actionErr
	}

	current, err := s.repo.GetProbeForProject(ctx, project.ID, input.ProbeID)
	if err != nil {
		return domainprobe.Probe{}, flow.probeLookupFailure(ProbeEventUpdateFailure, err)
	}

	refreshAssignments := input.Enabled != nil || input.LabelIDs != nil
	syncProbe, err := s.syncUpdateProbeInput(ctx, input, current)
	if err != nil {
		return domainprobe.Probe{}, flow.updateFailure(err)
	}

	updated, err := s.repo.UpdateProbe(ctx, syncProbe)
	if err != nil {
		return domainprobe.Probe{}, flow.updateFailure(err)
	}

	if refreshAssignments {
		if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForProbe(ctx, project.ID, updated.ID); err != nil {
			return domainprobe.Probe{}, flow.assignmentRefreshFailure(ProbeEventUpdateFailure, err)
		}
	}

	flow.success(ProbeEventUpdateSuccess)

	return updated, nil
}

func (s *Service) syncUpdateProbeInput(ctx context.Context, input UpdateProbeInput, current domainprobe.Probe) (output domainprobe.Probe, err error) {
	probe := current
	if input.Name != nil {
		probe.Name = *input.Name
	}
	if input.Enabled != nil {
		probe.Enabled = *input.Enabled
	}
	if input.SubdivisionCode != nil {
		probe.SubdivisionCode = input.SubdivisionCode
	}
	if input.Latitude != nil {
		probe.Latitude = input.Latitude
	}
	if input.Longitude != nil {
		probe.Longitude = input.Longitude
	}
	if input.LabelIDs != nil {
		outlabels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, current.ProjectID, *input.LabelIDs)
		if err != nil {
			return domainprobe.Probe{}, err
		}
		probe.Labels = outlabels
	}

	return probe, nil
}

func (s *Service) DeleteProbe(ctx context.Context, input TargetProbeInput) error {
	ctx, flow := s.startProbeFlow(ctx, "probe.delete", ProbeActionDelete, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetProbeInput(input)
	if err != nil {
		return flow.businessFailure(ProbeEventDeleteFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setProbeID(input.ProbeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return flow.projectLookupFailure(ProbeEventDeleteFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventDeleteFailure, domainproject.ActionManageProbes); actionErr != nil {
		return actionErr
	}

	if err := s.repo.SoftDeleteProbe(ctx, project.ID, input.ProbeID); err != nil {
		return flow.deleteFailure(err)
	}
	if err := s.assignmentRefresher.DeleteProbeCheckAssignmentsForProbe(ctx, project.ID, input.ProbeID); err != nil {
		return flow.assignmentDeleteFailure(ProbeEventDeleteFailure, err)
	}
	flow.success(ProbeEventDeleteSuccess)

	return nil
}

func (s *Service) RotateProbeSecret(ctx context.Context, input TargetProbeInput) (RotateProbeSecretOutput, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.secret.rotate", ProbeActionSecretRotate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetProbeInput(input)
	if err != nil {
		return RotateProbeSecretOutput{}, flow.businessFailure(ProbeEventSecretRotateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setProbeID(input.ProbeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return RotateProbeSecretOutput{}, flow.projectLookupFailure(ProbeEventSecretRotateFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventSecretRotateFailure, domainproject.ActionManageProbes); actionErr != nil {
		return RotateProbeSecretOutput{}, actionErr
	}

	probe, err := s.repo.GetProbeForProject(ctx, project.ID, input.ProbeID)
	if err != nil {
		return RotateProbeSecretOutput{}, flow.probeLookupFailure(ProbeEventSecretRotateFailure, err)
	}

	if s.secretGenerator == nil {
		return RotateProbeSecretOutput{}, flow.technicalFailure(
			ProbeEventSecretRotateFailure,
			ProbeReasonSecretGeneratorMissing,
			errors.New("probe secret generator is not configured"),
		)
	}
	plaintextSecret, secretHash, err := s.secretGenerator.GenerateProbeSecret()
	if err != nil {
		return RotateProbeSecretOutput{}, flow.technicalFailure(ProbeEventSecretRotateFailure, ProbeReasonSecretGenerateFailed, err)
	}

	if err := s.repo.RotateProbeSecret(ctx, domainprobe.Probe{
		ID:        input.ProbeID,
		ProjectID: project.ID,
	}, secretHash); err != nil {
		return RotateProbeSecretOutput{}, flow.rotateFailure(err)
	}
	flow.success(ProbeEventSecretRotateSuccess)

	return RotateProbeSecretOutput{
		Probe:  probe,
		Secret: plaintextSecret,
	}, nil
}

func (s *Service) requireProjectAction(
	ctx context.Context,
	flow *probeFlow,
	projectID string,
	userID string,
	event ProbeEventName,
	action domainproject.Action,
) (domainproject.Role, error) {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return "", flow.roleLookupFailure(event, err)
	}
	if !domainproject.Can(role, action) {
		return "", flow.businessFailure(event, ProbeReasonForbidden, ErrForbidden)
	}

	return role, nil
}
