package probe

import (
	"context"
	"errors"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/normalize"
)

type Service struct {
	repo            Repository
	secretGenerator SecretGenerator
	events          EventRecorder
}

func NewService(repo Repository, secretGenerator SecretGenerator, events EventRecorder) *Service {
	return &Service{
		repo:            repo,
		secretGenerator: secretGenerator,
		events:          events,
	}
}

func (s *Service) CreateProbe(ctx context.Context, input CreateProbeInput) (CreateProbeOutput, error) {
	ctx, flow := s.startProbeFlow(ctx, "probe.create", ProbeActionCreate, input.CurrentUserID)
	defer flow.End()
	flow.SetProjectRef(input.ProjectRef)

	normalized, err := normalizeCreateProbeInput(input)
	if err != nil {
		return CreateProbeOutput{}, flow.BusinessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	}

	projectID, err := s.repo.GetProjectIDForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if errors.Is(err, ErrProjectNotFound) {
		return CreateProbeOutput{}, flow.BusinessFailure(ProbeEventCreateFailure, ProbeReasonProjectNotFound, err)
	}
	if err != nil {
		return CreateProbeOutput{}, flow.TechnicalFailure(ProbeEventCreateFailure, ProbeReasonProjectLookupFailed, err)
	}
	flow.SetProjectID(projectID)

	if s.secretGenerator == nil {
		return CreateProbeOutput{}, flow.TechnicalFailure(
			ProbeEventCreateFailure,
			ProbeReasonSecretGeneratorMissing,
			errors.New("probe secret generator is not configured"),
		)
	}
	plaintextSecret, secretHash, err := s.secretGenerator.GenerateProbeSecret()
	if err != nil {
		return CreateProbeOutput{}, flow.TechnicalFailure(ProbeEventCreateFailure, ProbeReasonSecretGenerateFailed, err)
	}

	probe, err := s.repo.CreateProbe(ctx, domainprobe.CreateProbeStorageInput{
		ProjectID:  projectID,
		Name:       normalized.name,
		Enabled:    normalized.enabled,
		City:       normalized.city,
		Latitude:   normalized.latitude,
		Longitude:  normalized.longitude,
		LabelIDs:   normalized.labelIDs,
		SecretHash: secretHash,
	})
	if errors.Is(err, ErrInvalidInput) {
		return CreateProbeOutput{}, flow.BusinessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	}
	if errors.Is(err, ErrProjectNotFound) {
		return CreateProbeOutput{}, flow.BusinessFailure(ProbeEventCreateFailure, ProbeReasonProjectNotFound, err)
	}
	if errors.Is(err, ErrLabelNotFound) {
		return CreateProbeOutput{}, flow.BusinessFailure(ProbeEventCreateFailure, ProbeReasonLabelNotFound, err)
	}
	if err != nil {
		return CreateProbeOutput{}, flow.TechnicalFailure(ProbeEventCreateFailure, ProbeReasonProbeCreateFailed, err)
	}
	flow.SetProbe(probe)
	flow.Success(ProbeEventCreateSuccess)

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
