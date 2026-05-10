package proberegistry

import (
	"context"
	"errors"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Service struct {
	repo            Repository
	projectAccess   ProjectAccess
	labelAccess     LabelAccess
	secretGenerator SecretGenerator
	events          EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, secretGenerator SecretGenerator, events EventRecorder) *Service {
	return &Service{
		repo:            repo,
		projectAccess:   projectAccess,
		labelAccess:     labelAccess,
		secretGenerator: secretGenerator,
		events:          events,
	}
}

func (s *Service) CreateProbe(ctx context.Context, input CreateProbeInput) (CreateProbeOutput, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.create", ProbeActionCreate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeCreateProbeInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		return CreateProbeOutput{}, flow.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, input.CurrentUserID)
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

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, normalized.labelIDs)
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

	probe, err := s.repo.CreateProbe(ctx, domainprobe.CreateProbeStorageInput{
		ProjectID:  project.ID,
		Name:       normalized.name,
		Enabled:    normalized.enabled,
		City:       normalized.city,
		Latitude:   normalized.latitude,
		Longitude:  normalized.longitude,
		LabelIDs:   normalized.labelIDs,
		SecretHash: secretHash,
	})
	if err != nil {
		return CreateProbeOutput{}, flow.createFailure(err)
	}
	flow.setProbeID(probe.ID)
	probe.Labels = labels

	return CreateProbeOutput{
		Probe:  probe,
		Secret: plaintextSecret,
	}, nil
}

func (s *Service) ListProbes(ctx context.Context, input ListProbesInput) ([]domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.list", ProbeActionList, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeListProbesInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		return nil, flow.businessFailure(ProbeEventListFailure, ProbeReasonInvalidInput, err)
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

func (s *Service) GetProbe(ctx context.Context, input GetProbeInput) (domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.get", ProbeActionGet, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeGetProbeInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setProbeID(input.ProbeID)
		return domainprobe.Probe{}, flow.businessFailure(ProbeEventGetFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setProbeID(normalized.probeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, input.CurrentUserID)
	if err != nil {
		return domainprobe.Probe{}, flow.projectLookupFailure(ProbeEventGetFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventGetFailure, domainproject.ActionReadProject); actionErr != nil {
		return domainprobe.Probe{}, actionErr
	}

	probe, err := s.repo.GetProbeForProject(ctx, project.ID, normalized.probeID)
	if err != nil {
		return domainprobe.Probe{}, flow.probeLookupFailure(ProbeEventGetFailure, err)
	}

	return probe, nil
}

func (s *Service) UpdateProbe(ctx context.Context, input UpdateProbeInput) (domainprobe.Probe, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.update", ProbeActionUpdate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeUpdateProbeInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setProbeID(input.ProbeID)
		return domainprobe.Probe{}, flow.businessFailure(ProbeEventUpdateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setProbeID(normalized.probeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, input.CurrentUserID)
	if err != nil {
		return domainprobe.Probe{}, flow.projectLookupFailure(ProbeEventUpdateFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventUpdateFailure, domainproject.ActionManageProbes); actionErr != nil {
		return domainprobe.Probe{}, actionErr
	}

	current, err := s.repo.GetProbeForProject(ctx, project.ID, normalized.probeID)
	if err != nil {
		return domainprobe.Probe{}, flow.probeLookupFailure(ProbeEventUpdateFailure, err)
	}

	if normalized.replaceLabels {
		if _, labelErr := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, normalized.labelIDs); labelErr != nil {
			return domainprobe.Probe{}, flow.labelLookupFailure(ProbeEventUpdateFailure, labelErr)
		}
	}

	updated, err := s.repo.UpdateProbe(ctx, domainprobe.UpdateProbeStorageInput{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		Name:          chooseString(current.Name, normalized.name),
		Enabled:       chooseBool(current.Enabled, normalized.enabled),
		City:          chooseOptionalString(current.City, normalized.city),
		Latitude:      chooseOptionalFloat64(current.Latitude, normalized.latitude),
		Longitude:     chooseOptionalFloat64(current.Longitude, normalized.longitude),
		ReplaceLabels: normalized.replaceLabels,
		LabelIDs:      normalized.labelIDs,
	})
	if err != nil {
		return domainprobe.Probe{}, flow.updateFailure(err)
	}
	flow.success(ProbeEventUpdateSuccess)

	return updated, nil
}

func (s *Service) DeleteProbe(ctx context.Context, input DeleteProbeInput) error {
	ctx, flow := s.startProbeFlow(ctx, "probe.delete", ProbeActionDelete, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeDeleteProbeInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setProbeID(input.ProbeID)
		return flow.businessFailure(ProbeEventDeleteFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setProbeID(normalized.probeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, input.CurrentUserID)
	if err != nil {
		return flow.projectLookupFailure(ProbeEventDeleteFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventDeleteFailure, domainproject.ActionManageProbes); actionErr != nil {
		return actionErr
	}

	if err := s.repo.SoftDeleteProbe(ctx, project.ID, normalized.probeID); err != nil {
		return flow.deleteFailure(err)
	}
	flow.success(ProbeEventDeleteSuccess)

	return nil
}

func (s *Service) RotateProbeSecret(ctx context.Context, input RotateProbeSecretInput) (RotateProbeSecretOutput, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.secret.rotate", ProbeActionSecretRotate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeRotateProbeSecretInput(input)
	if err != nil {
		flow.setProjectRef(input.ProjectRef)
		flow.setProbeID(input.ProbeID)
		return RotateProbeSecretOutput{}, flow.businessFailure(ProbeEventSecretRotateFailure, ProbeReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setProbeID(normalized.probeID)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, input.CurrentUserID)
	if err != nil {
		return RotateProbeSecretOutput{}, flow.projectLookupFailure(ProbeEventSecretRotateFailure, err)
	}
	flow.setProjectID(project.ID)

	if _, actionErr := s.requireProjectAction(ctx, flow, project.ID, input.CurrentUserID, ProbeEventSecretRotateFailure, domainproject.ActionManageProbes); actionErr != nil {
		return RotateProbeSecretOutput{}, actionErr
	}

	probe, err := s.repo.GetProbeForProject(ctx, project.ID, normalized.probeID)
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

	if err := s.repo.RotateProbeSecret(ctx, domainprobe.RotateProbeSecretStorageInput{
		ProjectID:  project.ID,
		ProbeID:    normalized.probeID,
		SecretHash: secretHash,
	}); err != nil {
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
