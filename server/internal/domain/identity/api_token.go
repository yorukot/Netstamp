package identity

import "time"

type APITokenScope string

const (
	ScopeProjectsRead     APITokenScope = "projects:read"
	ScopeProjectsWrite    APITokenScope = "projects:write"
	ScopeProbesRead       APITokenScope = "probes:read"
	ScopeProbesWrite      APITokenScope = "probes:write"
	ScopeChecksRead       APITokenScope = "checks:read"
	ScopeChecksWrite      APITokenScope = "checks:write"
	ScopeLabelsRead       APITokenScope = "labels:read"
	ScopeLabelsWrite      APITokenScope = "labels:write"
	ScopeAssignmentsRead  APITokenScope = "assignments:read"
	ScopeResultsRead      APITokenScope = "results:read"
	ScopeAlertsRead       APITokenScope = "alerts:read"
	ScopeAlertsWrite      APITokenScope = "alerts:write"
	ScopeStatusPagesRead  APITokenScope = "status_pages:read"
	ScopeStatusPagesWrite APITokenScope = "status_pages:write"
)

var AllAPITokenScopes = []APITokenScope{
	ScopeProjectsRead, ScopeProjectsWrite,
	ScopeProbesRead, ScopeProbesWrite,
	ScopeChecksRead, ScopeChecksWrite,
	ScopeLabelsRead, ScopeLabelsWrite,
	ScopeAssignmentsRead, ScopeResultsRead,
	ScopeAlertsRead, ScopeAlertsWrite,
	ScopeStatusPagesRead, ScopeStatusPagesWrite,
}

func ValidAPITokenScope(scope APITokenScope) bool {
	for _, candidate := range AllAPITokenScopes {
		if scope == candidate {
			return true
		}
	}
	return false
}

type APIToken struct {
	ID            string
	UserID        string
	Name          string
	TokenHash     []byte
	TokenHint     string
	Scopes        []APITokenScope
	CreatedAt     time.Time
	LastUsedAt    *time.Time
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	RevokedReason *string
}

func (t APIToken) HasScope(scope APITokenScope) bool {
	for _, candidate := range t.Scopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

type CreatedAPIToken struct {
	Token    APIToken
	RawToken string
}
