package security

import (
	"context"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

func TestGoogleAuthorizationURLRequestsRecentAuthenticationForSudo(t *testing.T) {
	cfg := OIDCClientConfig{ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://netstamp.example.com/callback", Google: true, AllowedHostedDomains: []string{"example.com"}}
	client := &OIDCClient{cfg: cfg, oauth2Cfg: &oauth2.Config{
		ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, RedirectURL: cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/v2/auth"},
	}}
	authorizationURL, err := client.AuthorizationURL(context.Background(), "state", "nonce", "verifier", appauth.ExternalAuthIntentSudo)
	if err != nil {
		t.Fatalf("create Google authorization URL: %v", err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatalf("parse authorization URL: %v", err)
	}
	query := parsed.Query()
	if query.Get("prompt") != "select_account" {
		t.Fatalf("expected Google account picker, got %q", query.Get("prompt"))
	}
	if query.Get("claims") != `{"id_token":{"auth_time":{"essential":true}}}` {
		t.Fatalf("expected essential auth_time claim, got %q", query.Get("claims"))
	}
	if query.Get("hd") != "example.com" || query.Get("nonce") != "nonce" {
		t.Fatalf("unexpected Google authorization hints: %s", parsed.RawQuery)
	}
	if query.Get("code_challenge") == "" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("expected PKCE S256 parameters: %s", parsed.RawQuery)
	}
}

func TestGoogleIssuerAndHostedDomainValidation(t *testing.T) {
	client := NewGoogleOIDCClient(GoogleOIDCClientConfig{AllowedHostedDomains: []string{"Example.COM", "example.com"}})
	if got, err := client.validatedIssuer("accounts.google.com"); err != nil || got != googleIssuer {
		t.Fatalf("expected legacy Google issuer to normalize, got %q, %v", got, err)
	}
	if _, err := client.validatedIssuer("https://attacker.example.com"); err == nil {
		t.Fatal("expected an untrusted issuer to be rejected")
	}
	if !allowedHostedDomain("EXAMPLE.COM", client.cfg.AllowedHostedDomains) {
		t.Fatal("expected configured hosted domain to be accepted case-insensitively")
	}
	if allowedHostedDomain("other.example.com", client.cfg.AllowedHostedDomains) {
		t.Fatal("expected an unconfigured hosted domain to be rejected")
	}
	if len(client.cfg.AllowedHostedDomains) != 1 {
		t.Fatalf("expected normalized hosted domains, got %#v", client.cfg.AllowedHostedDomains)
	}
}
