package label

import appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"

const (
	maxLabelKeyRunes     = 100
	maxLabelValueRunes   = 100
	maxLabelProjectRunes = 100
)

type normalizedCreateLabelInput struct {
	projectRef string
	key        string
	value      string
}

type normalizedUpdateLabelInput struct {
	projectRef string
	labelID    string
	key        *string
	value      *string
}

func normalizeCreateLabelInput(input CreateLabelInput) (normalizedCreateLabelInput, error) {
	projectRef, err := normalizeLabelProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedCreateLabelInput{}, err
	}
	key, err := appvalidation.RequiredString(ErrInvalidInput, "key", input.Key, maxLabelKeyRunes)
	if err != nil {
		return normalizedCreateLabelInput{}, err
	}
	value, err := appvalidation.RequiredString(ErrInvalidInput, "value", input.Value, maxLabelValueRunes)
	if err != nil {
		return normalizedCreateLabelInput{}, err
	}

	return normalizedCreateLabelInput{projectRef: projectRef, key: key, value: value}, nil
}

func normalizeUpdateLabelInput(input UpdateLabelInput) (normalizedUpdateLabelInput, error) {
	projectRef, err := normalizeLabelProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedUpdateLabelInput{}, err
	}
	labelID, err := normalizeLabelID(input.LabelID)
	if err != nil {
		return normalizedUpdateLabelInput{}, err
	}
	key, err := appvalidation.OptionalString(ErrInvalidInput, "key", input.Key, maxLabelKeyRunes)
	if err != nil {
		return normalizedUpdateLabelInput{}, err
	}
	value, err := appvalidation.OptionalString(ErrInvalidInput, "value", input.Value, maxLabelValueRunes)
	if err != nil {
		return normalizedUpdateLabelInput{}, err
	}

	return normalizedUpdateLabelInput{projectRef: projectRef, labelID: labelID, key: key, value: value}, nil
}

func normalizeLabelProjectRef(value string) (string, error) {
	return appvalidation.RequiredString(ErrInvalidInput, "projectRef", value, maxLabelProjectRunes)
}

func normalizeLabelID(value string) (string, error) {
	return appvalidation.CanonicalUUID(ErrInvalidInput, "labelId", value)
}

func normalizeResolveLabelIDs(values []string) ([]string, error) {
	return appvalidation.CanonicalUUIDSet(ErrInvalidInput, "labelIds", values)
}

func invalidLabelField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
