package probe

import (
	"context"
	"fmt"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/normalize"
)

type Service struct {
	repo            Repository
	secretGenerator SecretGenerator
}

func NewService(repo Repository, secretGenerator SecretGenerator) *Service {
	return &Service{
		repo:            repo,
		secretGenerator: secretGenerator,
	}
}

func (s *Service) CreateProbe(ctx context.Context, input CreateProbeInput) (CreateProbeOutput, error) {
	normalized, err := normalizeCreateProbeInput(input)
	if err != nil {
		return CreateProbeOutput{}, err
	}

	projectID, err := s.repo.GetProjectIDForUser(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return CreateProbeOutput{}, err
	}

	if s.secretGenerator == nil {
		return CreateProbeOutput{}, fmt.Errorf("probe secret generator is not configured")
	}
	plaintextSecret, secretHash, err := s.secretGenerator.GenerateProbeSecret()
	if err != nil {
		return CreateProbeOutput{}, err
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
	if err != nil {
		return CreateProbeOutput{}, err
	}

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
