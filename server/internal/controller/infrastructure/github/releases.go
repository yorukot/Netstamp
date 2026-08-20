package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	appupdatecheck "github.com/yorukot/netstamp/internal/controller/application/updatecheck"
	appversion "github.com/yorukot/netstamp/internal/platform/version"
)

const (
	defaultAPIBaseURL = "https://api.github.com"
	maxReleaseBody    = 1 << 20
	requestTimeout    = 10 * time.Second
)

type ReleaseClient struct {
	httpClient *http.Client
	apiBaseURL string
	owner      string
	repo       string
}

func NewReleaseClient() *ReleaseClient {
	return &ReleaseClient{
		httpClient: &http.Client{
			Timeout:   requestTimeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		apiBaseURL: defaultAPIBaseURL,
		owner:      appversion.GitHubOwner,
		repo:       appversion.GitHubRepo,
	}
}

func (c *ReleaseClient) LatestRelease(ctx context.Context) (appupdatecheck.Release, error) {
	if c == nil || c.httpClient == nil {
		return appupdatecheck.Release{}, errors.New("github release client is unavailable")
	}
	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/releases/latest",
		strings.TrimRight(c.apiBaseURL, "/"),
		url.PathEscape(c.owner),
		url.PathEscape(c.repo),
	)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return appupdatecheck.Release{}, fmt.Errorf("create GitHub release request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "netstamp")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return appupdatecheck.Release{}, fmt.Errorf("request GitHub release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return appupdatecheck.Release{}, fmt.Errorf("GitHub release API returned status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxReleaseBody+1))
	if err != nil {
		return appupdatecheck.Release{}, fmt.Errorf("read GitHub release response: %w", err)
	}
	if len(body) > maxReleaseBody {
		return appupdatecheck.Release{}, errors.New("GitHub release response exceeds size limit")
	}
	var payload struct {
		TagName     string    `json:"tag_name"`
		HTMLURL     string    `json:"html_url"`
		PublishedAt time.Time `json:"published_at"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return appupdatecheck.Release{}, fmt.Errorf("decode GitHub release response: %w", err)
	}
	payload.TagName = strings.TrimSpace(payload.TagName)
	payload.HTMLURL = strings.TrimSpace(payload.HTMLURL)
	if payload.TagName == "" {
		return appupdatecheck.Release{}, errors.New("GitHub release response is missing tag_name")
	}
	if !validReleaseURL(payload.HTMLURL) {
		return appupdatecheck.Release{}, errors.New("GitHub release response has an invalid html_url")
	}
	if payload.PublishedAt.IsZero() {
		return appupdatecheck.Release{}, errors.New("GitHub release response is missing published_at")
	}
	return appupdatecheck.Release{
		TagName: payload.TagName, URL: payload.HTMLURL, PublishedAt: payload.PublishedAt.UTC(),
	}, nil
}

func validReleaseURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.IsAbs() && parsed.Hostname() != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.User == nil
}
