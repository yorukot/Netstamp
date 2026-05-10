package proberegistry

import appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"

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

type normalizedProjectProbeInput struct {
	projectRef string
	probeID    string
}

type normalizedUpdateProbeInput struct {
	projectRef    string
	probeID       string
	name          *string
	enabled       *bool
	city          *string
	latitude      *float64
	longitude     *float64
	replaceLabels bool
	labelIDs      []string
}

func normalizeCreateProbeInput(input CreateProbeInput) (normalizedCreateProbeInput, error) {
	projectRef, err := normalizeProbeProjectRef(input.ProjectRef)
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

func normalizeListProbesInput(input ListProbesInput) (string, error) {
	return normalizeProbeProjectRef(input.ProjectRef)
}

func normalizeGetProbeInput(input GetProbeInput) (normalizedProjectProbeInput, error) {
	return normalizeProjectProbeInput(input.ProjectRef, input.ProbeID)
}

func normalizeDeleteProbeInput(input DeleteProbeInput) (normalizedProjectProbeInput, error) {
	return normalizeProjectProbeInput(input.ProjectRef, input.ProbeID)
}

func normalizeRotateProbeSecretInput(input RotateProbeSecretInput) (normalizedProjectProbeInput, error) {
	return normalizeProjectProbeInput(input.ProjectRef, input.ProbeID)
}

func normalizeUpdateProbeInput(input UpdateProbeInput) (normalizedUpdateProbeInput, error) {
	ids, err := normalizeProjectProbeInput(input.ProjectRef, input.ProbeID)
	if err != nil {
		return normalizedUpdateProbeInput{}, err
	}
	if !hasUpdateProbeChanges(input) {
		return normalizedUpdateProbeInput{}, invalidProbeField("", "at least one field must be provided", nil)
	}

	name, err := appvalidation.OptionalString(ErrInvalidInput, "name", input.Name, maxProbeNameRunes)
	if err != nil {
		return normalizedUpdateProbeInput{}, err
	}
	city, err := appvalidation.OptionalString(ErrInvalidInput, "city", input.City, maxProbeCityRunes)
	if err != nil {
		return normalizedUpdateProbeInput{}, err
	}
	latitude, longitude, err := normalizeCoordinates(input.Latitude, input.Longitude)
	if err != nil {
		return normalizedUpdateProbeInput{}, err
	}
	replaceLabels, labelIDs, err := normalizeOptionalLabelIDs(input.LabelIDs)
	if err != nil {
		return normalizedUpdateProbeInput{}, err
	}

	return normalizedUpdateProbeInput{
		projectRef:    ids.projectRef,
		probeID:       ids.probeID,
		name:          name,
		enabled:       normalizeOptionalBool(input.Enabled),
		city:          city,
		latitude:      latitude,
		longitude:     longitude,
		replaceLabels: replaceLabels,
		labelIDs:      labelIDs,
	}, nil
}

func hasUpdateProbeChanges(input UpdateProbeInput) bool {
	return input.Name != nil ||
		input.Enabled != nil ||
		input.City != nil ||
		input.Latitude != nil ||
		input.Longitude != nil ||
		input.LabelIDs != nil
}

func normalizeProjectProbeInput(projectRefValue, probeIDValue string) (normalizedProjectProbeInput, error) {
	projectRef, err := normalizeProbeProjectRef(projectRefValue)
	if err != nil {
		return normalizedProjectProbeInput{}, err
	}
	probeID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "probeId", probeIDValue)
	if err != nil {
		return normalizedProjectProbeInput{}, err
	}

	return normalizedProjectProbeInput{projectRef: projectRef, probeID: probeID}, nil
}

func normalizeProbeProjectRef(value string) (string, error) {
	return appvalidation.RequiredString(ErrInvalidInput, "projectRef", value, maxProbeProjectRefRunes)
}

func normalizeOptionalBool(value *bool) *bool {
	if value == nil {
		return nil
	}

	normalized := *value
	return &normalized
}

func normalizeOptionalLabelIDs(value *[]string) (bool, []string, error) {
	if value == nil {
		return false, nil, nil
	}

	labelIDs, err := appvalidation.CanonicalUUIDSet(ErrInvalidInput, "labelIds", *value)
	if err != nil {
		return false, nil, err
	}

	return true, labelIDs, nil
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

func chooseString(current string, next *string) string {
	if next == nil {
		return current
	}

	return *next
}

func chooseBool(current bool, next *bool) bool {
	if next == nil {
		return current
	}

	return *next
}

func chooseOptionalString(current, next *string) *string {
	if next == nil {
		return current
	}

	return next
}

func chooseOptionalFloat64(current, next *float64) *float64 {
	if next == nil {
		return current
	}

	return next
}

func invalidProbeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
