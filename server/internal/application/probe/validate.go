package probe

import appvalidation "github.com/yorukot/netstamp/internal/application/validation"

const (
	maxProbeNameRunes       = 100
	maxProbeCityRunes       = 100
	maxProbeProjectRefRunes = 100
)

type normalizedCreateProbeInput struct {
	projectRef string
	name       string
	enabled    bool
	city       *string
	latitude   *float64
	longitude  *float64
	labelIDs   []string
}

func normalizeCreateProbeInput(input CreateProbeInput) (normalizedCreateProbeInput, error) {
	projectRef, err := appvalidation.RequiredString(ErrInvalidInput, "projectRef", input.ProjectRef, maxProbeProjectRefRunes)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	name, err := appvalidation.RequiredString(ErrInvalidInput, "name", input.Name, maxProbeNameRunes)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	city, err := appvalidation.OptionalString(ErrInvalidInput, "city", input.City, maxProbeCityRunes)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	latitude, longitude, err := normalizeCoordinates(input.Latitude, input.Longitude)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}
	labelIDs, err := appvalidation.CanonicalUUIDSet(ErrInvalidInput, "labelIds", input.LabelIDs)
	if err != nil {
		return normalizedCreateProbeInput{}, err
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	return normalizedCreateProbeInput{
		projectRef: projectRef,
		name:       name,
		enabled:    enabled,
		city:       city,
		latitude:   latitude,
		longitude:  longitude,
		labelIDs:   labelIDs,
	}, nil
}

func normalizeCoordinates(latitude, longitude *float64) (*float64, *float64, error) {
	if latitude != nil && longitude == nil {
		return nil, nil, invalidProbeField("longitude", "must be provided with latitude", nil)
	}
	if latitude == nil && longitude != nil {
		return nil, nil, invalidProbeField("latitude", "must be provided with longitude", nil)
	}
	if latitude == nil {
		return nil, nil, nil
	}

	lat := *latitude
	lon := *longitude
	return &lat, &lon, nil
}

func invalidProbeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
