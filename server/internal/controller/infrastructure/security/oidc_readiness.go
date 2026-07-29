package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	maxOIDCDiscoveryResponseBytes = 1 << 20
	maxOIDCDiscoveryRedirects     = 10
)

// OIDCDiscoveryMetadataError marks deterministic discovery configuration
// failures. Callers can distinguish these from transient transport and upstream
// availability errors without importing this infrastructure package.
type OIDCDiscoveryMetadataError struct {
	cause error
}

func (e *OIDCDiscoveryMetadataError) Error() string {
	return e.cause.Error()
}

func (e *OIDCDiscoveryMetadataError) Unwrap() error {
	return e.cause
}

func (*OIDCDiscoveryMetadataError) InvalidOIDCMetadata() {}

// OIDCReadinessChecker verifies that an issuer exposes usable OIDC discovery
// metadata. Request deadlines are owned by the application use case.
type OIDCReadinessChecker struct {
	client *http.Client
}

func NewOIDCReadinessChecker() OIDCReadinessChecker {
	return OIDCReadinessChecker{client: http.DefaultClient}
}

func (c OIDCReadinessChecker) Check(ctx context.Context, issuerURL string) error {
	metadata, err := c.discoveryMetadata(ctx, issuerURL)
	if err != nil {
		return err
	}
	if metadata.Issuer != issuerURL {
		return invalidOIDCMetadataError(fmt.Errorf(
			"discovery issuer %q does not match configured issuer %q",
			metadata.Issuer,
			issuerURL,
		))
	}
	for name, value := range map[string]string{
		"authorization_endpoint": metadata.AuthorizationEndpoint,
		"token_endpoint":         metadata.TokenEndpoint,
		"jwks_uri":               metadata.JWKSURL,
	} {
		if endpointErr := validateOIDCDiscoveryEndpoint(value); endpointErr != nil {
			return invalidOIDCMetadataError(fmt.Errorf("OIDC discovery %s: %w", name, endpointErr))
		}
	}
	return nil
}

type oidcDiscoveryMetadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURL               string `json:"jwks_uri"`
}

func (c OIDCReadinessChecker) discoveryMetadata(ctx context.Context, issuerURL string) (oidcDiscoveryMetadata, error) {
	discoveryURL := strings.TrimSuffix(strings.TrimSpace(issuerURL), "/") + "/.well-known/openid-configuration"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, http.NoBody)
	if err != nil {
		return oidcDiscoveryMetadata{}, invalidOIDCMetadataError(fmt.Errorf("build OIDC discovery request: %w", err))
	}

	client := c.redirectSafeClient()
	response, err := client.Do(request)
	if err != nil {
		return oidcDiscoveryMetadata{}, fmt.Errorf("request OIDC discovery: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusInternalServerError ||
		response.StatusCode == http.StatusTooManyRequests ||
		response.StatusCode == http.StatusRequestTimeout {
		return oidcDiscoveryMetadata{}, fmt.Errorf("OIDC discovery service returned status %d", response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		return oidcDiscoveryMetadata{}, invalidOIDCMetadataError(
			fmt.Errorf("OIDC discovery endpoint returned status %d", response.StatusCode),
		)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxOIDCDiscoveryResponseBytes+1))
	if err != nil {
		return oidcDiscoveryMetadata{}, fmt.Errorf("read OIDC discovery response: %w", err)
	}
	if len(body) > maxOIDCDiscoveryResponseBytes {
		return oidcDiscoveryMetadata{}, invalidOIDCMetadataError(errors.New("OIDC discovery response is too large"))
	}

	var metadata oidcDiscoveryMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return oidcDiscoveryMetadata{}, invalidOIDCMetadataError(fmt.Errorf("decode OIDC discovery metadata: %w", err))
	}
	return metadata, nil
}

func (c OIDCReadinessChecker) redirectSafeClient() *http.Client {
	base := c.client
	if base == nil {
		base = http.DefaultClient
	}
	client := *base
	previousCheck := client.CheckRedirect
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= maxOIDCDiscoveryRedirects {
			return invalidOIDCMetadataError(errors.New("OIDC discovery exceeded the redirect limit"))
		}
		if endpointErr := validateOIDCDiscoveryEndpoint(request.URL.String()); endpointErr != nil {
			return invalidOIDCMetadataError(fmt.Errorf("OIDC discovery redirect: %w", endpointErr))
		}
		if previousCheck != nil {
			return previousCheck(request, via)
		}
		return nil
	}
	return &client
}

func invalidOIDCMetadataError(cause error) error {
	return &OIDCDiscoveryMetadataError{cause: cause}
}

func validateOIDCDiscoveryEndpoint(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return errors.New("endpoint must be an absolute HTTPS URL")
	}
	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if isOIDCDiscoveryLoopbackHost(parsed.Hostname()) {
			return nil
		}
	}
	return errors.New("endpoint must use HTTPS except for a loopback development host")
}

func isOIDCDiscoveryLoopbackHost(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
