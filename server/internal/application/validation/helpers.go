package validation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/yorukot/netstamp/internal/normalize"
)

func RequiredString(base error, field, value string, maxRunes int) (string, error) {
	return RequiredStringRange(base, field, value, 0, maxRunes)
}

func RequiredStringRange(base error, field, value string, minRunes, maxRunes int) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", New(base, field, "must not be blank", value)
	}
	if minRunes > 0 && utf8.RuneCountInString(normalized) < minRunes {
		return "", New(base, field, minRunesMessage(minRunes), value)
	}
	if maxRunes > 0 && utf8.RuneCountInString(normalized) > maxRunes {
		return "", New(base, field, maxRunesMessage(maxRunes), value)
	}

	return normalized, nil
}

func OptionalString(base error, field string, value *string, maxRunes int) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	normalized, err := RequiredString(base, field, *value, maxRunes)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func PositiveInt32(base error, field string, value int32) (int32, error) {
	if value <= 0 {
		return 0, New(base, field, "must be greater than 0", value)
	}

	return value, nil
}

func OptionalPositiveInt32(base error, field string, value *int32) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	normalized, err := PositiveInt32(base, field, *value)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func OptionalInt32Range(base error, field string, value *int32, minValue, maxValue int32) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}
	if *value < minValue || *value > maxValue {
		return nil, New(base, field, fmt.Sprintf("must be between %d and %d", minValue, maxValue), *value)
	}

	normalized := *value
	return &normalized, nil
}

func CanonicalUUIDSet(base error, field string, values []string) ([]string, error) {
	normalized, err := normalize.CanonicalUUIDSet(values, base)
	if err != nil {
		return nil, New(base, field, "must contain valid UUIDs", values)
	}

	return normalized, nil
}

func maxRunesMessage(maxRunes int) string {
	return fmt.Sprintf("must be at most %d characters", maxRunes)
}

func minRunesMessage(minRunes int) string {
	return fmt.Sprintf("must be at least %d characters", minRunes)
}
