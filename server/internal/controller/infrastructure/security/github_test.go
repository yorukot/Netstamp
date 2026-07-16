package security

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

func TestGitHubAuthorizationURLUsesPKCEAndMinimalScopes(t *testing.T) {
	client := NewGitHubOAuthClient(GitHubOAuthClientConfig{
		ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://netstamp.example.com/callback", AllowSignup: false,
	})
	authorizationURL, err := client.AuthorizationURL(context.Background(), "state", "unused", "verifier", appauth.ExternalAuthIntentLogin)
	if err != nil {
		t.Fatalf("create GitHub authorization URL: %v", err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatalf("parse GitHub authorization URL: %v", err)
	}
	query := parsed.Query()
	if query.Get("state") != "state" || query.Get("code_challenge") == "" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("expected state and PKCE S256 parameters: %s", parsed.RawQuery)
	}
	if query.Get("allow_signup") != "false" {
		t.Fatalf("expected sign-up policy in authorization URL, got %q", query.Get("allow_signup"))
	}
	if scope := query.Get("scope"); !strings.Contains(scope, "read:user") || !strings.Contains(scope, "user:email") {
		t.Fatalf("expected minimal profile and email scopes, got %q", scope)
	}
}

func TestGitHubExchangeUsesStableIDAndPrimaryVerifiedEmail(t *testing.T) {
	roundTripper := &githubRoundTripper{}
	client := NewGitHubOAuthClient(GitHubOAuthClientConfig{ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://netstamp.example.com/callback"})
	client.httpClient = &http.Client{Transport: roundTripper}

	claims, err := client.Exchange(context.Background(), "authorization-code", "verifier", "")
	if err != nil {
		t.Fatalf("exchange GitHub callback: %v", err)
	}
	if claims.Issuer != githubIssuer || claims.Subject != "1234567" || claims.Username != "octocat" {
		t.Fatalf("unexpected GitHub identity: %#v", claims)
	}
	if claims.Email != "primary@example.com" || !claims.EmailVerified || claims.DisplayName != "The Octocat" {
		t.Fatalf("unexpected GitHub profile claims: %#v", claims)
	}
	if !roundTripper.sawVerifier {
		t.Fatal("expected token exchange to send the PKCE verifier")
	}
	if roundTripper.apiCalls != 2 {
		t.Fatalf("expected user and email API calls, got %d", roundTripper.apiCalls)
	}
}

func TestGitHubExchangeDoesNotTrustUnverifiedEmail(t *testing.T) {
	roundTripper := &githubRoundTripper{emailsResponse: `[{"email":"primary@example.com","primary":true,"verified":false}]`}
	client := NewGitHubOAuthClient(GitHubOAuthClientConfig{ClientID: "client-id", ClientSecret: "client-secret"})
	client.httpClient = &http.Client{Transport: roundTripper}

	claims, err := client.Exchange(context.Background(), "authorization-code", "verifier", "")
	if err != nil {
		t.Fatalf("exchange GitHub callback: %v", err)
	}
	if claims.Email != "" || claims.EmailVerified {
		t.Fatalf("expected unverified GitHub email to be discarded: %#v", claims)
	}
}

type githubRoundTripper struct {
	sawVerifier    bool
	apiCalls       int
	emailsResponse string
}

func (r *githubRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Request: request}
	response.Header.Set("Content-Type", "application/json")
	switch request.URL.Host + request.URL.Path {
	case "github.com/login/oauth/access_token":
		body, _ := io.ReadAll(request.Body)
		values, _ := url.ParseQuery(string(body))
		r.sawVerifier = values.Get("code_verifier") == "verifier"
		response.Body = io.NopCloser(strings.NewReader(`{"access_token":"callback-only-token","token_type":"bearer","scope":"read:user,user:email"}`))
	case "api.github.com/user":
		r.apiCalls++
		if request.Header.Get("Authorization") != "Bearer callback-only-token" || request.Header.Get(githubVersionHeader) != githubAPIVersion {
			response.StatusCode = http.StatusUnauthorized
		}
		response.Body = io.NopCloser(strings.NewReader(`{"id":1234567,"login":"octocat","name":"The Octocat","avatar_url":"https://avatars.githubusercontent.com/u/1234567"}`))
	case "api.github.com/user/emails":
		r.apiCalls++
		emailsResponse := r.emailsResponse
		if emailsResponse == "" {
			emailsResponse = `[{"email":"other@example.com","primary":false,"verified":true},{"email":"primary@example.com","primary":true,"verified":true}]`
		}
		response.Body = io.NopCloser(strings.NewReader(emailsResponse))
	default:
		response.StatusCode = http.StatusNotFound
		response.Body = io.NopCloser(strings.NewReader(`{"message":"not found"}`))
	}
	return response, nil
}
