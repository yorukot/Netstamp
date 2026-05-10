package normalize

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var projectSlugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func Email(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func RequiredString(value string, invalidErr error) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", invalidErr
	}

	return value, nil
}

func OptionalString(value *string, invalidErr error) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional string.
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, invalidErr
	}

	return &normalized, nil
}

func ProjectSlug(value string, invalidErr error) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !IsProjectSlug(value) {
		return "", invalidErr
	}

	return value, nil
}

func IsProjectSlug(value string) bool {
	return projectSlugPattern.MatchString(value)
}

func CanonicalUUIDSet(values []string, invalidErr error) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(strings.TrimSpace(value))
		if err != nil {
			return nil, invalidErr
		}

		canonical := parsed.String()
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		normalized = append(normalized, canonical)
	}

	return normalized, nil
}
