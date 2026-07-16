package security

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

type OIDCClientConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
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

func (c *OIDCClient) AuthorizationURL(ctx context.Context, state, nonce, pkceVerifier string, forceReauthentication bool) (string, error) {
	config, err := c.config(ctx)
	if err != nil {
		return "", err
	}
	options := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(pkceVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	if forceReauthentication {
		options = append(options, oauth2.SetAuthURLParam("prompt", "login"), oauth2.SetAuthURLParam("max_age", "0"))
	}
	return config.AuthCodeURL(state, options...), nil
}

func (c *OIDCClient) Exchange(ctx context.Context, code, pkceVerifier, nonce string) (appauth.OIDCClaims, error) {
	config, err := c.config(ctx)
	if err != nil {
		return appauth.OIDCClaims{}, err
	}
	token, err := config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", pkceVerifier))
	if err != nil {
		return appauth.OIDCClaims{}, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return appauth.OIDCClaims{}, errors.New("oidc response missing id token")
	}

	c.mu.Lock()
	provider := c.provider
	c.mu.Unlock()
	idToken, err := provider.Verifier(&oidc.Config{ClientID: c.cfg.ClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return appauth.OIDCClaims{}, err
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
	}
	if err := idToken.Claims(&claims); err != nil {
		return appauth.OIDCClaims{}, err
	}
	if claims.Nonce != nonce {
		return appauth.OIDCClaims{}, errors.New("oidc nonce mismatch")
	}
	displayName := claims.Name
	if displayName == "" {
		displayName = claims.PreferredName
	}
	var authTime time.Time
	if claims.AuthTime > 0 {
		authTime = time.Unix(claims.AuthTime, 0).UTC()
	}
	return appauth.OIDCClaims{
		Issuer: idToken.Issuer, Subject: idToken.Subject, Email: claims.Email,
		EmailVerified: claims.EmailVerified, DisplayName: displayName, AuthTime: authTime,
	}, nil
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
