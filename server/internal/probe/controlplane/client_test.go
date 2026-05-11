package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func TestClientHelloSendsRuntimeStatus(t *testing.T) {
	client := testClient(t, func(r *http.Request) *http.Response {
		if r.URL.Path != "/api/v1/probes/probe-1/runtime/hello" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Probe secret" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var body statusInputBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.AgentVersion == nil || *body.AgentVersion != "probe-test/1.0.0" {
			t.Fatalf("unexpected agent version: %#v", body.AgentVersion)
		}

		return jsonResponse(http.StatusOK, `{"serverTime":"2026-05-10T00:00:00Z","heartbeatIntervalSeconds":15,"assignmentPollIntervalSeconds":20}`)
	})
	output, err := client.Hello(context.Background(), StatusInput{AgentVersion: "probe-test/1.0.0"})
	if err != nil {
		t.Fatalf("hello: %v", err)
	}
	if output.HeartbeatIntervalSeconds != 15 || output.AssignmentPollIntervalSeconds != 20 {
		t.Fatalf("unexpected hello output: %#v", output)
	}
}

func TestClientPollAssignmentsMapsResponse(t *testing.T) {
	client := testClient(t, func(r *http.Request) *http.Response {
		if r.URL.Path != "/api/v1/probes/probe-1/runtime/assignments" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		return jsonResponse(http.StatusOK, `{
			"assignments": [{
				"assignmentId": "assignment-1",
				"projectId": "project-1",
				"probeId": "probe-1",
				"checkId": "check-1",
				"checkVersion": "check-version",
				"selectorVersion": "selector-version",
				"type": "ping",
				"target": "127.0.0.1",
				"intervalSeconds": 30,
				"pingConfig": {"packetCount": 2, "packetSizeBytes": 56, "timeoutMs": 1000, "ipFamily": "inet"}
			}]
		}`)
	})
	output, err := client.PollAssignments(context.Background())
	if err != nil {
		t.Fatalf("poll assignments: %v", err)
	}
	if len(output.Assignments) != 1 {
		t.Fatalf("expected one assignment, got %#v", output.Assignments)
	}
	assignment := output.Assignments[0]
	if assignment.ID != "assignment-1" || assignment.Type != domaincheck.TypePing || assignment.Target != "127.0.0.1" {
		t.Fatalf("unexpected assignment: %#v", assignment)
	}
	if assignment.PingConfig.IPFamily == nil || *assignment.PingConfig.IPFamily != domainnetwork.IPFamilyInet {
		t.Fatalf("unexpected ping config: %#v", assignment.PingConfig)
	}
}

func TestClientSubmitResultsGroupsByCheckID(t *testing.T) {
	client := testClient(t, func(r *http.Request) *http.Response {
		if r.URL.Path != "/api/v1/probes/probe-1/runtime/results" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]resultGroupInputBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		group := body["check-1"]
		if group.Type != domaincheck.TypePing {
			t.Fatalf("unexpected group type: %#v", group.Type)
		}
		if group.Detail.AssignmentID != "assignment-1" || group.Detail.CheckVersion != "check-version" || group.Detail.SelectorVersion != "selector-version" {
			t.Fatalf("unexpected detail: %#v", group.Detail)
		}
		if len(group.Results) != 1 || group.Results[0].Status != domainping.StatusSuccessful {
			t.Fatalf("unexpected results: %#v", group.Results)
		}

		return jsonResponse(http.StatusOK, `{"accepted":true,"resyncNeeded":false,"staleChecks":[],"assignments":[]}`)
	})
	output, err := client.SubmitResults(context.Background(), domainprobe.ResultBatch{
		ProbeID: "probe-1",
		Results: []domainprobe.Result{{
			AssignmentID:    "assignment-1",
			CheckID:         "check-1",
			CheckVersion:    "check-version",
			SelectorVersion: "selector-version",
			Type:            domaincheck.TypePing,
			Ping: domainping.Result{
				StartedAt:     time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				FinishedAt:    time.Date(2026, 5, 10, 0, 0, 1, 0, time.UTC),
				Status:        domainping.StatusSuccessful,
				SentCount:     1,
				ReceivedCount: 1,
			},
		}},
	})
	if err != nil {
		t.Fatalf("submit results: %v", err)
	}
	if !output.Accepted {
		t.Fatalf("expected accepted output, got %#v", output)
	}
}

func TestClientReturnsStatusError(t *testing.T) {
	client := testClient(t, func(*http.Request) *http.Response {
		return textResponse(http.StatusUnauthorized, "nope")
	})
	_, err := client.PollAssignments(context.Background())
	var statusErr StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected status error, got %T %v", err, err)
	}
	if statusErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: %d", statusErr.StatusCode)
	}
}

func testClient(t *testing.T, roundTrip func(*http.Request) *http.Response) *Client {
	t.Helper()

	return &Client{
		baseURL: "http://controller/api/v1",
		probeID: "probe-1",
		secret:  "secret",
		httpClient: &http.Client{
			Transport: roundTripFunc(roundTrip),
		},
	}
}

type roundTripFunc func(*http.Request) *http.Response

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req), nil
}

func jsonResponse(status int, body string) *http.Response {
	res := textResponse(status, body)
	res.Header.Set("Content-Type", "application/json")
	return res
}

func textResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
