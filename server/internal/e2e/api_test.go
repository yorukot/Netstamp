//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestAPIAuthProjectAndProbeRuntimeFlow(t *testing.T) {
	suite := newAPISuite(t)
	email := "e2e-" + randomHex(t, 4) + "@example.com"
	password := "correct-horse-battery-staple"

	var registered authResponse
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

	var login authResponse
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

	var currentUser meResponse
	suite.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, authHeaders(login.AccessToken), http.StatusOK, &currentUser)
	if !currentUser.Authenticated || currentUser.User.ID != registered.User.ID {
		t.Fatalf("expected current user from token, got %#v", currentUser)
	}

	suite.doJSON(t, http.MethodGet, "/api/v1/projects", nil, nil, http.StatusUnauthorized, nil)

	var createdProject projectResponse
	suite.doJSON(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"name": "E2E Project",
		"slug": "e2e-" + randomHex(t, 4),
	}, authHeaders(login.AccessToken), http.StatusCreated, &createdProject)
	if createdProject.Project.ID == "" {
		t.Fatal("expected created project id")
	}

	var projects listProjectsResponse
	suite.doJSON(t, http.MethodGet, "/api/v1/projects", nil, authHeaders(login.AccessToken), http.StatusOK, &projects)
	if len(projects.Projects) != 1 || projects.Projects[0].ID != createdProject.Project.ID {
		t.Fatalf("expected created project in project list, got %#v", projects.Projects)
	}

	var fetchedProject projectResponse
	suite.doJSON(t, http.MethodGet, "/api/v1/projects/"+createdProject.Project.Slug, nil, authHeaders(login.AccessToken), http.StatusOK, &fetchedProject)
	if fetchedProject.Project.ID != createdProject.Project.ID {
		t.Fatalf("expected fetched project id %q, got %q", createdProject.Project.ID, fetchedProject.Project.ID)
	}

	var createdProbe createProbeResponse
	suite.doJSON(t, http.MethodPost, "/api/v1/projects/"+createdProject.Project.Slug+"/probes", map[string]any{
		"name":    "e2e-probe",
		"enabled": true,
		"city":    "JP-13",
	}, authHeaders(login.AccessToken), http.StatusCreated, &createdProbe)
	if createdProbe.Probe.ID == "" {
		t.Fatal("expected created probe id")
	}
	if createdProbe.Secret == "" {
		t.Fatal("expected one-time probe secret")
	}

	suite.doJSON(t, http.MethodPost, "/api/v1/probes/"+createdProbe.Probe.ID+"/runtime/hello", map[string]any{
		"agentVersion": "netstamp-e2e/0.1.0",
	}, probeHeaders("wrong-secret"), http.StatusUnauthorized, nil)

	var hello helloResponse
	suite.doJSON(t, http.MethodPost, "/api/v1/probes/"+createdProbe.Probe.ID+"/runtime/hello", map[string]any{
		"agentVersion": "netstamp-e2e/0.1.0",
		"publicV4":     "203.0.113.10",
		"addrs":        []string{"10.0.0.10"},
	}, probeHeaders(createdProbe.Secret), http.StatusOK, &hello)
	if hello.HeartbeatIntervalSeconds <= 0 || hello.AssignmentPollIntervalSeconds <= 0 {
		t.Fatalf("expected runtime intervals in hello response, got %#v", hello)
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
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
}

type helloResponse struct {
	HeartbeatIntervalSeconds      int32 `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32 `json:"assignmentPollIntervalSeconds"`
}
