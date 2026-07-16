package apitoken

import (
	"sort"
	"strings"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	MaxActiveTokens = 20
	MaxTokenTTL     = 365 * 24 * time.Hour
)

func normalizeCreateInput(input CreateInput, now time.Time) (CreateInput, error) {
	input.CurrentUserID = strings.TrimSpace(input.CurrentUserID)
	input.Name = strings.TrimSpace(input.Name)
	if input.CurrentUserID == "" || input.Name == "" || len(input.Name) > 100 || strings.TrimSpace(input.CurrentPassword) == "" {
		return CreateInput{}, ErrInvalidInput
	}
	if !input.ExpiresAt.After(now) || input.ExpiresAt.After(now.Add(MaxTokenTTL)) {
		return CreateInput{}, ErrInvalidInput
	}
	seen := make(map[string]struct{}, len(input.Scopes))
	normalized := make([]string, 0, len(input.Scopes))
	for _, raw := range input.Scopes {
		scope := string(identity.APITokenScope(strings.TrimSpace(raw)))
		if !identity.ValidAPITokenScope(identity.APITokenScope(scope)) {
			return CreateInput{}, ErrInvalidInput
		}
		if _, exists := seen[scope]; exists {
			return CreateInput{}, ErrInvalidInput
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	if len(normalized) == 0 {
		return CreateInput{}, ErrInvalidInput
	}
	sort.Strings(normalized)
	input.Scopes = normalized
	input.ExpiresAt = input.ExpiresAt.UTC()
	return input, nil
}
