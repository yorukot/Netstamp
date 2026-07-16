package apitoken

import (
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type CreateInput struct {
	CurrentUserID string
	Name          string
	Scopes        []string
	ExpiresAt     time.Time
}

type CreateOutput struct {
	Token    identity.APIToken
	RawToken string
}

type ListInput struct {
	CurrentUserID string
}

type RevokeInput struct {
	CurrentUserID string
	TokenID       string
}

type Principal struct {
	TokenID string
	UserID  string
	Scopes  []identity.APITokenScope
}

func (p Principal) HasScope(scope identity.APITokenScope) bool {
	for _, candidate := range p.Scopes {
		if candidate == scope {
			return true
		}
	}
	return false
}
