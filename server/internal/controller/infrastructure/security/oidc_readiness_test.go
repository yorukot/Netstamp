package security

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOIDCReadinessCheckerAcceptsMatchingDiscoveryIssuer(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "https://issuer.example.com",
		"authorization_endpoint": "https://issuer.example.com/authorize",
		"token_endpoint": "https://issuer.example.com/token",
		"jwks_uri": "https://issuer.example.com/keys"
	}`)

	if err := checker.Check(context.Background(), "https://issuer.example.com"); err != nil {
		t.Fatalf("check OIDC readiness: %v", err)
	}
}

func TestOIDCReadinessCheckerAcceptsLoopbackHTTPDiscoveryEndpoints(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "http://127.0.0.1:5556",
		"authorization_endpoint": "http://127.0.0.1:5556/authorize",
		"token_endpoint": "http://localhost:5556/token",
		"jwks_uri": "http://[::1]:5556/keys"
	}`)

	if err := checker.Check(context.Background(), "http://127.0.0.1:5556"); err != nil {
		t.Fatalf("check loopback OIDC readiness: %v", err)
	}
}

func TestOIDCReadinessCheckerMarksMismatchedIssuerAsInvalidMetadata(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "https://different.example.com",
		"authorization_endpoint": "https://different.example.com/authorize",
		"token_endpoint": "https://different.example.com/token",
		"jwks_uri": "https://different.example.com/keys"
	}`)

	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
}

func TestOIDCReadinessCheckerMarksMissingRequiredEndpointAsInvalidMetadata(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "https://issuer.example.com",
		"authorization_endpoint": "https://issuer.example.com/authorize",
		"jwks_uri": "https://issuer.example.com/keys"
	}`)

	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
}

func TestOIDCReadinessCheckerRejectsEndpointWithoutHostname(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "https://issuer.example.com",
		"authorization_endpoint": "https://issuer.example.com/authorize",
		"token_endpoint": "https://:443/token",
		"jwks_uri": "https://issuer.example.com/keys"
	}`)

	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
}

func TestOIDCReadinessCheckerRejectsPublicHTTPDiscoveryEndpoint(t *testing.T) {
	checker := discoveryChecker(http.StatusOK, `{
		"issuer": "https://issuer.example.com",
		"authorization_endpoint": "http://issuer.example.com/authorize",
		"token_endpoint": "https://issuer.example.com/token",
		"jwks_uri": "https://issuer.example.com/keys"
	}`)

	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
}

func TestOIDCReadinessCheckerRejectsRedirectDowngradeToInternalHTTP(t *testing.T) {
	requests := 0
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requests++
			return &http.Response{
				StatusCode: http.StatusFound,
				Header: http.Header{
					"Location": []string{"http://169.254.169.254/latest/meta-data/"},
				},
				Body: io.NopCloser(strings.NewReader("redirecting")),
			}, nil
		}),
	}
	checker := OIDCReadinessChecker{client: client}

	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
	if requests != 1 {
		t.Fatalf("unsafe redirect reached its target: requests=%d", requests)
	}
	if client.CheckRedirect != nil {
		t.Fatal("readiness check mutated the injected HTTP client")
	}
}

func TestOIDCReadinessCheckerClassifiesDiscoveryHTTPFailures(t *testing.T) {
	for _, status := range []int{
		http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			checker := discoveryChecker(status, "temporarily unavailable")

			err := checker.Check(context.Background(), "https://issuer.example.com")
			if err == nil {
				t.Fatal("expected discovery failure")
			}
			var metadataErr interface{ InvalidOIDCMetadata() }
			if errors.As(err, &metadataErr) {
				t.Fatalf("expected upstream failure not to be marked invalid metadata: %v", err)
			}
		})
	}

	checker := discoveryChecker(http.StatusNotFound, "not found")
	err := checker.Check(context.Background(), "https://issuer.example.com")
	assertInvalidOIDCMetadata(t, err)
}

func discoveryChecker(status int, body string) OIDCReadinessChecker {
	return OIDCReadinessChecker{
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		},
	}
}

func assertInvalidOIDCMetadata(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected invalid discovery metadata")
	}
	var metadataErr interface{ InvalidOIDCMetadata() }
	if !errors.As(err, &metadataErr) {
		t.Fatalf("expected invalid metadata marker, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
