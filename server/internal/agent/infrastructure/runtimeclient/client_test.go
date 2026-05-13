package runtimeclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	agentruntime "github.com/yorukot/netstamp/internal/agent/runtime"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func TestRuntimeClientHelloUsesProbeAuthAndEmptyBody(t *testing.T) {
	t.Parallel()

	const probeID = "11111111-1111-1111-1111-111111111111"
	client := New(agentconfig.Config{
		ControllerURL: "http://controller.test",
		APIVersion:    agentconfig.DefaultAPIVersion,
		ProbeID:       probeID,
		ProbeSecret:   "secret-value",
		HTTPTimeout:   time.Second,
	})
	client.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/runtime/probes/"+probeID+"/hello" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Probe secret-value" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if len(body) != 0 {
			t.Fatalf("expected empty hello body, got %q", string(body))
		}

		var response bytes.Buffer
		_ = json.NewEncoder(&response).Encode(agentruntime.HelloResponse{
			ServerTime:                   time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC),
			MinimumSupportedAgentVersion: agentruntime.Version,
			Config:                       domainprobe.DefaultRuntimeConfig(),
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(&response),
			Header:     make(http.Header),
		}, nil
	})}

	output, err := client.Hello(t.Context())
	if err != nil {
		t.Fatalf("hello failed: %v", err)
	}
	if output.MinimumSupportedAgentVersion != agentruntime.Version {
		t.Fatalf("unexpected minimum supported version %q", output.MinimumSupportedAgentVersion)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
