package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

const (
	githubIssuer        = "https://github.com"
	githubAPIBaseURL    = "https://api.github.com"
	githubAPIVersion    = "2026-03-10"
	githubMaxBody       = 1 << 20
	githubAuthURL       = "https://github.com/login/oauth/authorize"
	githubTokenURL      = "https://github.com/login/oauth/access_token" //nolint:gosec // Public OAuth endpoint, not credential material.
	githubVersionHeader = "X-Github-Api-Version"
)

type GitHubOAuthClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AllowSignup  bool
}

type GitHubOAuthClient struct {
	oauth2Config *oauth2.Config
	httpClient   *http.Client
	apiBaseURL   string
	allowSignup  bool
}

func NewGitHubOAuthClient(cfg GitHubOAuthClientConfig) *GitHubOAuthClient {
	return &GitHubOAuthClient{
		oauth2Config: &oauth2.Config{
			ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, RedirectURL: cfg.RedirectURL,
			Scopes: []string{"read:user", "user:email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  githubAuthURL,
				TokenURL: githubTokenURL,
			},
		},
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		apiBaseURL:  githubAPIBaseURL,
		allowSignup: cfg.AllowSignup,
	}
}

func (c *GitHubOAuthClient) AuthorizationURL(_ context.Context, state, _, pkceVerifier, intent string) (string, error) {
	options := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(pkceVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("allow_signup", strconv.FormatBool(c.allowSignup)),
	}
	if intent != appauth.ExternalAuthIntentLogin {
		options = append(options, oauth2.SetAuthURLParam("prompt", "select_account"))
	}
	return c.oauth2Config.AuthCodeURL(state, options...), nil
}

func (c *GitHubOAuthClient) Exchange(ctx context.Context, code, pkceVerifier, _ string) (appauth.ExternalIdentityClaims, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(pkceVerifier) == "" {
		return appauth.ExternalIdentityClaims{}, errors.New("github callback is incomplete")
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
	token, err := c.oauth2Config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", pkceVerifier))
	if err != nil || token.AccessToken == "" {
		return appauth.ExternalIdentityClaims{}, errors.New("github token exchange failed")
	}

	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.getJSON(ctx, token.AccessToken, "/user", &user); err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	if user.ID <= 0 {
		return appauth.ExternalIdentityClaims{}, errors.New("github user response is missing a stable id")
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := c.getJSON(ctx, token.AccessToken, "/user/emails", &emails); err != nil {
		return appauth.ExternalIdentityClaims{}, err
	}
	email := ""
	for _, candidate := range emails {
		if candidate.Primary && candidate.Verified {
			email = strings.TrimSpace(candidate.Email)
			break
		}
	}
	displayName := strings.TrimSpace(user.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(user.Login)
	}
	return appauth.ExternalIdentityClaims{
		Issuer: githubIssuer, Subject: strconv.FormatInt(user.ID, 10), Email: email, EmailVerified: email != "",
		DisplayName: displayName, Username: strings.TrimSpace(user.Login), AvatarURL: strings.TrimSpace(user.AvatarURL),
	}, nil
}

func (c *GitHubOAuthClient) getJSON(ctx context.Context, accessToken, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.apiBaseURL, "/")+path, http.NoBody)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set(githubVersionHeader, githubAPIVersion)
	request.Header.Set("User-Agent", "netstamp")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("github API request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("github API returned status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, githubMaxBody))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode github API response: %w", err)
	}
	return nil
}
