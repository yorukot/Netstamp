package probe

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
		return CreateProbeOutput{}, flow.projectLookupFailure(err)
	}
	flow.setProjectID(project.ID)

	role, err := s.projectAccess.GetMemberRole(ctx, project.ID, input.CurrentUserID)
	if err != nil {
		return CreateProbeOutput{}, flow.roleLookupFailure(err)
	}
	if !domainproject.Can(role, domainproject.ActionCreateProbe) {
		return CreateProbeOutput{}, flow.businessFailure(ProbeEventCreateFailure, ProbeReasonForbidden, ErrForbidden)
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, normalized.labelIDs)
	if err != nil {
		return CreateProbeOutput{}, flow.labelLookupFailure(err)
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
	probe.Labels = labels

	return CreateProbeOutput{
		Probe:  probe,
		Secret: plaintextSecret,
	}, nil
}
