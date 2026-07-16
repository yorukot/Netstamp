package security

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

type OIDCClientConfig struct {
	IssuerURL            string
	ClientID             string
	ClientSecret         string
	RedirectURL          string
	Google               bool
	AcceptedIssuers      []string
	CanonicalIssuer      string
	AllowedHostedDomains []string
}

type OIDCClient struct {
	cfg       OIDCClientConfig
	mu        sync.Mutex
	provider  *oidc.Provider
	oauth2Cfg *oauth2.Config
}

func NewOIDCClient(cfg OIDCClientConfig) *OIDCClient {
	return &OIDCClient{cfg: cfg}
}

func (c *OIDCClient) AuthorizationURL(ctx context.Context, state, nonce, pkceVerifier, intent string) (string, error) {
	config, err := c.config(ctx)
	if err != nil {
		return "", err
	}
	options := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(pkceVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	if c.cfg.Google {
		if len(c.cfg.AllowedHostedDomains) > 0 {
			options = append(options, oauth2.SetAuthURLParam("hd", c.cfg.AllowedHostedDomains[0]))
		}
		if intent != appauth.ExternalAuthIntentLogin {
			options = append(options, oauth2.SetAuthURLParam("prompt", "select_account"))
		}
		if intent == appauth.ExternalAuthIntentSudo {
			options = append(options, oauth2.SetAuthURLParam("claims", `{"id_token":{"auth_time":{"essential":true}}}`))
		}
	} else if intent != appauth.ExternalAuthIntentLogin {
		options = append(options, oauth2.SetAuthURLParam("prompt", "login"), oauth2.SetAuthURLParam("max_age", "0"))
	}
	return config.AuthCodeURL(state, options...), nil
}

func (c *OIDCClient) Exchange(ctx context.Context, code, pkceVerifier, nonce string) (appauth.ExternalIdentityClaims, error) {
	config, err := c.config(ctx)
	if err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	token, err := config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", pkceVerifier))
	if err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return appauth.ExternalIdentityClaims{}, errors.New("oidc response missing id token")
	}

	c.mu.Lock()
	provider := c.provider
	c.mu.Unlock()
	idToken, err := provider.Verifier(&oidc.Config{ClientID: c.cfg.ClientID, SkipIssuerCheck: len(c.cfg.AcceptedIssuers) > 0}).Verify(ctx, rawIDToken)
	if err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	var claims struct {
		Issuer        string `json:"iss"`
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		PreferredName string `json:"preferred_username"`
		Nonce         string `json:"nonce"`
		AuthTime      int64  `json:"auth_time"`
		HostedDomain  string `json:"hd"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	if claims.Nonce != nonce {
		return appauth.ExternalIdentityClaims{}, errors.New("oidc nonce mismatch")
	}
	issuer, err := c.validatedIssuer(idToken.Issuer)
	if err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	if !allowedHostedDomain(claims.HostedDomain, c.cfg.AllowedHostedDomains) {
		return appauth.ExternalIdentityClaims{}, errors.New("oidc hosted domain is not allowed")
	}
	displayName := claims.Name
	if displayName == "" {
		displayName = claims.PreferredName
	}
	var authTime time.Time
	if claims.AuthTime > 0 {
		authTime = time.Unix(claims.AuthTime, 0).UTC()
	}
	return appauth.ExternalIdentityClaims{
		Issuer: issuer, Subject: idToken.Subject, Email: claims.Email,
		EmailVerified: claims.EmailVerified, DisplayName: displayName, AuthTime: authTime,
	}, nil
}

func (c *OIDCClient) validatedIssuer(issuer string) (string, error) {
	if len(c.cfg.AcceptedIssuers) == 0 {
		return issuer, nil
	}
	for _, accepted := range c.cfg.AcceptedIssuers {
		if issuer == accepted {
			if c.cfg.CanonicalIssuer != "" {
				return c.cfg.CanonicalIssuer, nil
			}
			return issuer, nil
		}
	}
	return "", errors.New("oidc issuer is not allowed")
}

func allowedHostedDomain(hostedDomain string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	hostedDomain = strings.ToLower(strings.TrimSpace(hostedDomain))
	for _, domain := range allowed {
		if hostedDomain == strings.ToLower(strings.TrimSpace(domain)) {
			return true
		}
	}
	return false
}

func (c *OIDCClient) config(ctx context.Context) (*oauth2.Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.oauth2Cfg != nil {
		return c.oauth2Cfg, nil
	}
	provider, err := oidc.NewProvider(ctx, c.cfg.IssuerURL)
	if err != nil {
		return nil, err
	}
	c.provider = provider
	c.oauth2Cfg = &oauth2.Config{
		ClientID: c.cfg.ClientID, ClientSecret: c.cfg.ClientSecret, RedirectURL: c.cfg.RedirectURL,
		Endpoint: provider.Endpoint(), Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return c.oauth2Cfg, nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
