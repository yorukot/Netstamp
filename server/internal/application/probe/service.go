package probe

import (
	"context"
	"errors"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
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
	flow.setProjectRef(input.ProjectRef)

	normalized, err := normalizeCreateProbeInput(input)
	if err != nil {
		return CreateProbeOutput{}, flow.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	}

	project, err := s.projectAccess.GetProjectForUser(ctx, input.ProjectRef, input.CurrentUserID)
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

type normalizedCreateProbeInput struct {
	name      string
	enabled   bool
	city      *string
	latitude  *float64
	longitude *float64
	labelIDs  []string
}

func normalizeCreateProbeInput(input CreateProbeInput) (normalizedCreateProbeInput, error) {
	name, err := normalize.RequiredString(input.Name, ErrInvalidInput)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	city, err := normalize.OptionalString(input.City, ErrInvalidInput)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	latitude, longitude, err := normalizeCoordinates(input.Latitude, input.Longitude)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	labelIDs, err := normalize.CanonicalUUIDSet(input.LabelIDs, ErrInvalidInput)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	return normalizedCreateProbeInput{
		name:      name,
		enabled:   enabled,
		city:      city,
		latitude:  latitude,
		longitude: longitude,
		labelIDs:  labelIDs,
	}, nil
}

func normalizeCoordinates(latitude *float64, longitude *float64) (*float64, *float64, error) {
	if (latitude == nil) != (longitude == nil) {
		return nil, nil, ErrInvalidInput
	}
	if latitude == nil {
		return nil, nil, nil
	}

	lat := *latitude
	lon := *longitude
	return &lat, &lon, nil
}
