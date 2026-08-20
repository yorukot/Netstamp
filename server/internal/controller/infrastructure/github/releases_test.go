package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReleaseClientGetsLatestReleaseWithoutAuthentication(t *testing.T) {
	publishedAt := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	client := NewReleaseClient()
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/repos/yorukot/netstamp/releases/latest" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Accept") != "application/vnd.github+json" || request.Header.Get("User-Agent") != "netstamp" {
			t.Errorf("unexpected headers: %#v", request.Header)
		}
		if request.Header.Get("Authorization") != "" {
			t.Errorf("unexpected authorization header: %q", request.Header.Get("Authorization"))
		}
		body := fmt.Sprintf(`{"tag_name":"v1.2.3","html_url":"https://github.com/yorukot/netstamp/releases/tag/v1.2.3","published_at":%q}`, publishedAt.Format(time.RFC3339))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})}

	release, err := client.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("get latest release: %v", err)
	}
	if release.TagName != "v1.2.3" || release.URL != "https://github.com/yorukot/netstamp/releases/tag/v1.2.3" || release.PublishedAt != publishedAt {
		t.Fatalf("unexpected release: %#v", release)
	}
}

func TestReleaseClientRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "failure status", statusCode: http.StatusForbidden, body: `{}`},
		{name: "invalid json", statusCode: http.StatusOK, body: `{`},
		{name: "missing tag", statusCode: http.StatusOK, body: `{"html_url":"https://example.com/release","published_at":"2026-08-20T09:00:00Z"}`},
		{name: "invalid url", statusCode: http.StatusOK, body: `{"tag_name":"v1.0.0","html_url":"javascript:alert(1)","published_at":"2026-08-20T09:00:00Z"}`},
		{name: "missing publication time", statusCode: http.StatusOK, body: `{"tag_name":"v1.0.0","html_url":"https://example.com/release"}`},
		{name: "oversized response", statusCode: http.StatusOK, body: strings.Repeat("x", maxReleaseBody+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewReleaseClient()
			client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.statusCode,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(test.body)),
					Request:    request,
				}, nil
			})}
			if _, err := client.LatestRelease(context.Background()); err == nil {
				t.Fatal("expected invalid response error")
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
