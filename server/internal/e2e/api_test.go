//go:build integration

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAPIAuthProjectAndProbeRuntimeFlow(t *testing.T) {
	suite := newAPISuite(t)
	email := "e2e-" + randomHex(t, 4) + "@example.com"
	password := "correct-horse-battery-staple"

	var registered authResponse
	t.Logf("e2e: registering user %s", email)
	suite.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":       " " + email + " ",
		"displayName": " E2E User ",
		"password":    password,
	}, nil, http.StatusCreated, &registered)
	if registered.User.ID == "" {
		t.Fatal("expected registered user id")
	}
	if registered.User.Email != email {
		t.Fatalf("expected normalized registered email %q, got %q", email, registered.User.Email)
	}
	if registered.User.DisplayName == nil || *registered.User.DisplayName != "E2E User" {
		t.Fatalf("expected trimmed registered display name, got %#v", registered.User.DisplayName)
	}
	if registered.AccessToken == "" || registered.TokenType != "Bearer" {
		t.Fatalf("expected bearer token in register response, got %#v", registered)
	}
	t.Logf("e2e: registered user id %s", registered.User.ID)

	var login authResponse
	t.Logf("e2e: logging in user %s", email)
	suite.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	}, nil, http.StatusOK, &login)
	if login.User.ID != registered.User.ID {
		t.Fatalf("expected login user id %q, got %q", registered.User.ID, login.User.ID)
	}
	if login.AccessToken == "" {
		t.Fatal("expected login access token")
	}
	t.Log("e2e: received login access token")

	var currentUser meResponse
	t.Log("e2e: verifying bearer token with /auth/me")
	suite.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, authHeaders(login.AccessToken), http.StatusOK, &currentUser)
	if !currentUser.Authenticated || currentUser.User.ID != registered.User.ID {
		t.Fatalf("expected current user from token, got %#v", currentUser)
	}

	t.Log("e2e: verifying protected project list rejects missing bearer token")
	suite.doJSON(t, http.MethodGet, "/api/v1/projects", nil, nil, http.StatusUnauthorized, nil)

	var createdProject projectResponse
	projectSlug := "e2e-" + randomHex(t, 4)
	t.Logf("e2e: creating project %q", projectSlug)
	suite.doJSON(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "E2E Project",
		"slug": projectSlug,
	}, authHeaders(login.AccessToken), http.StatusCreated, &createdProject)
	if createdProject.Project.ID == "" {
		t.Fatal("expected created project id")
	}
	t.Logf("e2e: created project id %s", createdProject.Project.ID)

	var projects listProjectsResponse
	t.Log("e2e: listing projects for authenticated user")
	suite.doJSON(t, http.MethodGet, "/api/v1/projects", nil, authHeaders(login.AccessToken), http.StatusOK, &projects)
	if len(projects.Projects) != 1 || projects.Projects[0].ID != createdProject.Project.ID {
		t.Fatalf("expected created project in project list, got %#v", projects.Projects)
	}

	var fetchedProject projectResponse
	t.Logf("e2e: fetching project by slug %q", createdProject.Project.Slug)
	suite.doJSON(t, http.MethodGet, "/api/v1/projects/"+createdProject.Project.Slug, nil, authHeaders(login.AccessToken), http.StatusOK, &fetchedProject)
	if fetchedProject.Project.ID != createdProject.Project.ID {
		t.Fatalf("expected fetched project id %q, got %q", createdProject.Project.ID, fetchedProject.Project.ID)
	}

	var createdLabel labelResponse
	t.Logf("e2e: creating label for project %q", createdProject.Project.Slug)
	suite.doJSON(t, http.MethodPost, "/api/v1/projects/"+createdProject.Project.Slug+"/labels", map[string]any{
		"key":   " region ",
		"value": " tokyo ",
	}, authHeaders(login.AccessToken), http.StatusCreated, &createdLabel)
	if createdLabel.Label.ID == "" {
		t.Fatal("expected created label id")
	}
	if createdLabel.Label.Key != "region" || createdLabel.Label.Value != "tokyo" {
		t.Fatalf("expected normalized label, got %#v", createdLabel.Label)
	}
	t.Logf("e2e: created label id %s", createdLabel.Label.ID)

	var labels listLabelsResponse
	t.Log("e2e: listing labels for authenticated user")
	suite.doJSON(t, http.MethodGet, "/api/v1/projects/"+createdProject.Project.Slug+"/labels", nil, authHeaders(login.AccessToken), http.StatusOK, &labels)
	if len(labels.Labels) != 1 || labels.Labels[0].ID != createdLabel.Label.ID {
		t.Fatalf("expected created label in label list, got %#v", labels.Labels)
	}

	var createdProbe createProbeResponse
	t.Logf("e2e: creating labeled probe under project %q", createdProject.Project.Slug)
	suite.doJSON(t, http.MethodPost, "/api/v1/projects/"+createdProject.Project.Slug+"/probes", map[string]any{
		"name":            "e2e-probe",
		"enabled":         true,
		"subdivisionCode": "JP-13",
		"labelIds":        []string{createdLabel.Label.ID},
	}, authHeaders(login.AccessToken), http.StatusCreated, &createdProbe)
	if createdProbe.Probe.ID == "" {
		t.Fatal("expected created probe id")
	}
	if createdProbe.Secret == "" {
		t.Fatal("expected one-time probe secret")
	}
	t.Logf("e2e: created probe id %s and received one-time secret", createdProbe.Probe.ID)

	var plainCheck checkResponse
	t.Log("e2e: creating match-all ping check")
	suite.doJSON(t, http.MethodPost, "/api/v1/projects/"+createdProject.Project.Slug+"/checks", map[string]any{
		"name":            " controller-ping ",
		"type":            "ping",
		"target":          " 1.1.1.1 ",
		"intervalSeconds": 30,
	}, authHeaders(login.AccessToken), http.StatusCreated, &plainCheck)
	if plainCheck.Check.ID == "" {
		t.Fatal("expected plain check id")
	}
	if plainCheck.Check.Name != "controller-ping" || plainCheck.Check.Target != "1.1.1.1" {
		t.Fatalf("expected normalized plain check, got %#v", plainCheck.Check)
	}
	if len(plainCheck.Check.Selector) != 0 {
		t.Fatalf("expected empty selector for match-all check, got %#v", plainCheck.Check.Selector)
	}
	t.Logf("e2e: created match-all check id %s", plainCheck.Check.ID)

	var labeledCheck checkResponse
	t.Log("e2e: creating labeled ping check with label selector")
	suite.doJSON(t, http.MethodPost, "/api/v1/projects/"+createdProject.Project.Slug+"/checks", map[string]any{
		"name":   " tokyo-edge-ping ",
		"type":   "ping",
		"target": " 8.8.8.8 ",
		"selector": map[string]any{
			"label": map[string]any{
				"key":   " region ",
				"op":    " eq ",
				"value": " tokyo ",
			},
		},
		"description":     " Tokyo edge probes ",
		"intervalSeconds": 60,
		"pingConfig": map[string]any{
			"packetCount": 5,
			"timeoutMs":   2000,
			"ipFamily":    "inet",
		},
		"labelIds": []string{createdLabel.Label.ID},
	}, authHeaders(login.AccessToken), http.StatusCreated, &labeledCheck)
	if labeledCheck.Check.ID == "" {
		t.Fatal("expected labeled check id")
	}
	if len(labeledCheck.Check.Labels) != 1 || labeledCheck.Check.Labels[0].ID != createdLabel.Label.ID {
		t.Fatalf("expected label attached to labeled check, got %#v", labeledCheck.Check.Labels)
	}
	assertLabelSelector(t, labeledCheck.Check.Selector, "region", "eq", "tokyo")
	if labeledCheck.Check.PingConfig.PacketCount != 5 || labeledCheck.Check.PingConfig.TimeoutMs != 2000 || labeledCheck.Check.PingConfig.IPFamily == nil || *labeledCheck.Check.PingConfig.IPFamily != "inet" {
		t.Fatalf("expected custom ping config on labeled check, got %#v", labeledCheck.Check)
	}
	t.Logf("e2e: created label-selector check id %s", labeledCheck.Check.ID)

	var checks listChecksResponse
	t.Log("e2e: listing checks for authenticated user")
	suite.doJSON(t, http.MethodGet, "/api/v1/projects/"+createdProject.Project.Slug+"/checks", nil, authHeaders(login.AccessToken), http.StatusOK, &checks)
	if !containsCheck(checks.Checks, plainCheck.Check.ID) || !containsCheck(checks.Checks, labeledCheck.Check.ID) {
		t.Fatalf("expected both created checks in check list, got %#v", checks.Checks)
	}

	var assignments assignmentsResponse
	t.Log("e2e: listing probe runtime assignments after check creation")
	suite.doJSON(t, http.MethodGet, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/assignments", nil, probeHeaders(createdProbe.Secret), http.StatusOK, &assignments)
	if !containsAssignment(assignments.Assignments, plainCheck.Check.ID) || !containsAssignment(assignments.Assignments, labeledCheck.Check.ID) {
		t.Fatalf("expected assignments for both checks, got %#v", assignments.Assignments)
	}
	if assignments.ServerTime.IsZero() || assignments.Config.HeartbeatIntervalSeconds <= 0 || assignments.Config.AssignmentPollIntervalSeconds <= 0 {
		t.Fatalf("expected runtime metadata in assignments response, got %#v", assignments)
	}
	t.Logf("e2e: probe runtime returned %d assignments", len(assignments.Assignments))

	t.Log("e2e: verifying probe runtime rejects an invalid secret")
	suite.doJSON(t, http.MethodPost, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/hello", nil, probeHeaders("wrong-secret"), http.StatusUnauthorized, nil)

	var hello helloResponse
	t.Log("e2e: starting probe runtime session with valid secret")
	suite.doJSON(t, http.MethodPost, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/hello", nil, probeHeaders(createdProbe.Secret), http.StatusOK, &hello)
	if hello.ServerTime.IsZero() || hello.MinimumSupportedAgentVersion == "" || hello.Config.HeartbeatIntervalSeconds <= 0 || hello.Config.AssignmentPollIntervalSeconds <= 0 {
		t.Fatalf("expected runtime metadata in hello response, got %#v", hello)
	}
	t.Logf(
		"e2e: probe runtime hello returned minimumSupportedAgentVersion=%s heartbeat=%ds assignmentPoll=%ds",
		hello.MinimumSupportedAgentVersion,
		hello.Config.HeartbeatIntervalSeconds,
		hello.Config.AssignmentPollIntervalSeconds,
	)

	var submitted submitResultsResponse
	submitPayload := map[string]any{
		"results": []map[string]any{{
			"checkId": plainCheck.Check.ID,
			"type":    "ping",
			"ping": []map[string]any{{
				"startedAt":     "2026-05-13T10:00:00Z",
				"finishedAt":    "2026-05-13T10:00:01Z",
				"durationMs":    1000,
				"status":        "successful",
				"sentCount":     4,
				"receivedCount": 4,
				"lossPercent":   0,
				"rttMinMs":      10.1,
				"rttAvgMs":      12.3,
				"rttMedianMs":   12.0,
				"rttMaxMs":      15.6,
				"rttStddevMs":   1.7,
				"rttSamplesMs":  []float64{10.1, 11.5, 12.0, 15.6},
				"resolvedIp":    "1.1.1.1",
				"ipFamily":      "inet",
			}},
		}},
	}
	t.Log("e2e: submitting probe runtime ping results")
	suite.doJSON(t, http.MethodPost, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/results", submitPayload, probeHeaders(createdProbe.Secret), http.StatusOK, &submitted)
	if submitted.Accepted != 1 {
		t.Fatalf("expected one accepted result, got %#v", submitted)
	}
	t.Log("e2e: submitting duplicate probe runtime ping results")
	suite.doJSON(t, http.MethodPost, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/results", submitPayload, probeHeaders(createdProbe.Secret), http.StatusOK, &submitted)
	if submitted.Accepted != 1 {
		t.Fatalf("expected duplicate submit to remain accepted, got %#v", submitted)
	}

	assertExtensionNotEnabled(t, suite, "timescaledb_toolkit")

	seriesStart := time.Date(2026, 5, 13, 10, 5, 0, 0, time.UTC)
	seriesPayload := map[string]any{
		"results": []map[string]any{{
			"checkId": plainCheck.Check.ID,
			"type":    "ping",
			"ping": []map[string]any{
				pingResultPayload(seriesStart, 11.0),
				pingResultPayload(seriesStart.Add(time.Minute), 21.0),
				pingResultPayload(seriesStart.Add(2*time.Minute), 31.0),
			},
		}},
	}
	t.Log("e2e: submitting multiple ping results for bucketed series query")
	suite.doJSON(t, http.MethodPost, "/api/v1/runtime/probes/"+createdProbe.Probe.ID+"/results", seriesPayload, probeHeaders(createdProbe.Secret), http.StatusOK, &submitted)
	if submitted.Accepted != 3 {
		t.Fatalf("expected three accepted series results, got %#v", submitted)
	}

	var series pingSeriesResponse
	seriesPath := fmt.Sprintf(
		"/api/v1/projects/%s/results/ping/series?probeId=%s&checkId=%s&from=%d&to=%d&metric=rttAvgMs&maxDataPoints=2",
		createdProject.Project.Slug,
		createdProbe.Probe.ID,
		plainCheck.Check.ID,
		seriesStart.Add(-time.Minute).UnixMilli(),
		seriesStart.Add(3*time.Minute).UnixMilli(),
	)
	t.Log("e2e: querying ping result series with maxDataPoints forcing bucket resolution")
	suite.doJSON(t, http.MethodGet, seriesPath, nil, authHeaders(login.AccessToken), http.StatusOK, &series)
	if series.Query.Resolution != "bucket" {
		t.Fatalf("expected bucket resolution without TimescaleDB Toolkit LTTB, got %#v", series.Query)
	}
	if series.Query.TotalPoints != 3 {
		t.Fatalf("expected three source points in series metadata, got %#v", series.Query)
	}
	if len(series.Series) != 1 || series.Series[0].Name != "rttAvgMs" || len(series.Series[0].Points) == 0 {
		t.Fatalf("expected rttAvgMs series with bucketed points, got %#v", series.Series)
	}
	if len(series.Series[0].Points) > 2 {
		t.Fatalf("expected bucketed series to honor maxDataPoints target, got %d points: %#v", len(series.Series[0].Points), series.Series[0].Points)
	}
}

func assertExtensionNotEnabled(t *testing.T, suite *apiSuite, extension string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var enabled bool
	if err := suite.pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = $1)", extension).Scan(&enabled); err != nil {
		t.Fatalf("query extension %q state: %v", extension, err)
	}
	if enabled {
		t.Fatalf("expected extension %q to be absent from migrated database", extension)
	}
}

func pingResultPayload(startedAt time.Time, rttAvgMs float64) map[string]any {
	return map[string]any{
		"startedAt":     startedAt.Format(time.RFC3339),
		"finishedAt":    startedAt.Add(time.Second).Format(time.RFC3339),
		"durationMs":    1000,
		"status":        "successful",
		"sentCount":     4,
		"receivedCount": 4,
		"lossPercent":   0,
		"rttMinMs":      rttAvgMs - 1,
		"rttAvgMs":      rttAvgMs,
		"rttMedianMs":   rttAvgMs,
		"rttMaxMs":      rttAvgMs + 1,
		"rttStddevMs":   1,
		"rttSamplesMs":  []float64{rttAvgMs - 1, rttAvgMs, rttAvgMs, rttAvgMs + 1},
		"resolvedIp":    "1.1.1.1",
		"ipFamily":      "inet",
	}
}

type authResponse struct {
	User        userResponse `json:"user"`
	TokenType   string       `json:"tokenType"`
	AccessToken string       `json:"accessToken"`
	ExpiresIn   int          `json:"expiresIn"`
}

type meResponse struct {
	Authenticated bool         `json:"authenticated"`
	User          userResponse `json:"user"`
}

type userResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"displayName"`
}

type projectResponse struct {
	Project projectBody `json:"project"`
}

type listProjectsResponse struct {
	Projects []projectBody `json:"projects"`
}

type projectBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type createProbeResponse struct {
	Probe  probeBody `json:"probe"`
	Secret string    `json:"secret"`
}

type probeBody struct {
	ID        string      `json:"id"`
	ProjectID string      `json:"projectId"`
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Labels    []labelBody `json:"labels"`
}

type helloResponse struct {
	ServerTime                   time.Time         `json:"serverTime"`
	MinimumSupportedAgentVersion string            `json:"minimumSupportedAgentVersion"`
	Config                       runtimeConfigBody `json:"config"`
}

type submitResultsResponse struct {
	Accepted int `json:"accepted"`
}

type pingSeriesResponse struct {
	Series []seriesBody        `json:"series"`
	Query  seriesQueryMetadata `json:"query"`
}

type seriesBody struct {
	Name   string             `json:"name"`
	Labels map[string]string  `json:"labels"`
	Unit   string             `json:"unit"`
	Points []seriesPointTuple `json:"points"`
}

type seriesPointTuple [2]float64

type seriesQueryMetadata struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Resolution    string `json:"resolution"`
	TotalPoints   int64  `json:"totalPoints"`
}

type labelResponse struct {
	Label labelBody `json:"label"`
}

type listLabelsResponse struct {
	Labels []labelBody `json:"labels"`
}

type labelBody struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

type checkResponse struct {
	Check checkBody `json:"check"`
}

type listChecksResponse struct {
	Checks []checkBody `json:"checks"`
}

type checkBody struct {
	ID              string               `json:"id"`
	ProjectID       string               `json:"projectId"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	Target          string               `json:"target"`
	Selector        map[string]any       `json:"selector"`
	Description     *string              `json:"description"`
	IntervalSeconds int32                `json:"intervalSeconds"`
	PingConfig      pingConfigBody       `json:"pingConfig"`
	Labels          []checkLabelResponse `json:"labels"`
}

type pingConfigBody struct {
	PacketCount     int32   `json:"packetCount"`
	PacketSizeBytes int32   `json:"packetSizeBytes"`
	TimeoutMs       int32   `json:"timeoutMs"`
	IPFamily        *string `json:"ipFamily"`
}

type checkLabelResponse struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type assignmentsResponse struct {
	ServerTime  time.Time         `json:"serverTime"`
	Config      runtimeConfigBody `json:"config"`
	Assignments []assignmentBody  `json:"assignments"`
}

type runtimeConfigBody struct {
	HeartbeatIntervalSeconds      int32 `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32 `json:"assignmentPollIntervalSeconds"`
	MaxConcurrentWorkers          int32 `json:"maxConcurrentWorkers"`
	InitialBackoffSeconds         int32 `json:"initialBackoffSeconds"`
	MaxBackoffSeconds             int32 `json:"maxBackoffSeconds"`
	MaxAttempts                   int32 `json:"maxAttempts"`
}

type assignmentBody struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	CheckVersion    string `json:"checkVersion"`
	SelectorVersion string `json:"selectorVersion"`
	Check           struct {
		ID              string `json:"id"`
		Type            string `json:"type"`
		Target          string `json:"target"`
		IntervalSeconds int32  `json:"intervalSeconds"`
	} `json:"check"`
}

func containsCheck(checks []checkBody, checkID string) bool {
	for _, check := range checks {
		if check.ID == checkID {
			return true
		}
	}
	return false
}

func containsAssignment(assignments []assignmentBody, checkID string) bool {
	for _, assignment := range assignments {
		if assignment.Check.ID == checkID {
			return true
		}
	}
	return false
}

func assertLabelSelector(t *testing.T, selector map[string]any, key, op, value string) {
	t.Helper()

	labelSelector, ok := selector["label"].(map[string]any)
	if !ok {
		t.Fatalf("expected label selector object, got %#v", selector)
	}
	if labelSelector["key"] != key || labelSelector["op"] != op || labelSelector["value"] != value {
		t.Fatalf("expected label selector %s %s %s, got %#v", key, op, value, labelSelector)
	}
}
