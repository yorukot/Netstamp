package security

import "strings"

const googleIssuer = "https://accounts.google.com"

type GoogleOIDCClientConfig struct {
	ClientID             string
	ClientSecret         string
	RedirectURL          string
	AllowedHostedDomains []string
}

func NewGoogleOIDCClient(cfg GoogleOIDCClientConfig) *OIDCClient {
	return NewOIDCClient(OIDCClientConfig{
		IssuerURL: googleIssuer, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, RedirectURL: cfg.RedirectURL,
		Google: true, AcceptedIssuers: []string{googleIssuer, "accounts.google.com"}, CanonicalIssuer: googleIssuer,
		AllowedHostedDomains: normalizeHostedDomains(cfg.AllowedHostedDomains),
	})
}

func normalizeHostedDomains(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}
