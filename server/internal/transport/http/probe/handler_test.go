package probe

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	appauth "github.com/yorukot/netstamp/internal/application/auth"
	appprobe "github.com/yorukot/netstamp/internal/application/probe"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testLabelID   = "44444444-4444-4444-4444-444444444444"
)

func TestCreateProbeReturnsCreatedProbeAndSecret(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{}
	NewHandler(appprobe.NewService(repo, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}), &handlerTokenVerifier{
		claims: appauth.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name":      " tokyo-vps-1 ",
		"enabled":   true,
		"city":      " JP-13 ",
		"latitude":  35.6762,
		"longitude": 139.6503,
		"labelIds":  []string{testLabelID},
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body createProbeOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Secret != "plain-secret" {
		t.Fatalf("expected top-level secret, got %q", body.Secret)
	}
	if body.Probe.ID != testProbeID {
		t.Fatalf("expected probe id, got %q", body.Probe.ID)
	}
	if body.Probe.City == nil || *body.Probe.City != "JP-13" {
		t.Fatalf("expected trimmed city, got %#v", body.Probe.City)
	}
	if len(body.Probe.Labels) != 1 || body.Probe.Labels[0].ID != testLabelID {
		t.Fatalf("expected full labels, got %#v", body.Probe.Labels)
	}
	if repo.gotProjectRef != "engineering" {
		t.Fatalf("expected project ref, got %q", repo.gotProjectRef)
	}
}

func TestCreateProbeRequiresBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}), &handlerTokenVerifier{}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name": "tokyo-vps-1",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestCreateProbeMapsInvalidInputToUnprocessableEntity(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}), &handlerTokenVerifier{
		claims: appauth.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name": "   ",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestCreateProbeMapsInaccessibleProjectToNotFound(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{
		projectErr: appprobe.ErrProjectNotFound,
	}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}), &handlerTokenVerifier{
		claims: appauth.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/missing/probes", map[string]any{
		"name": "tokyo-vps-1",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

func TestCreateProbeMapsMissingLabelToNotFound(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{
		createErr: appprobe.ErrLabelNotFound,
	}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}), &handlerTokenVerifier{
		claims: appauth.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name":     "tokyo-vps-1",
		"labelIds": []string{testLabelID},
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

type handlerTokenVerifier struct {
	claims appauth.AccessTokenClaims
	err    error
}

func (v *handlerTokenVerifier) VerifyAccessToken(context.Context, string) (appauth.AccessTokenClaims, error) {
	if v.err != nil {
		return appauth.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

type handlerProbeRepository struct {
	gotProjectRef  string
	projectErr     error
	gotCreateInput appprobe.CreateProbeStorageInput
	createErr      error
}

func (r *handlerProbeRepository) GetProjectIDForUser(_ context.Context, projectRef string, _ string) (string, error) {
	r.gotProjectRef = projectRef
	if r.projectErr != nil {
		return "", r.projectErr
	}
	return testProjectID, nil
}

func (r *handlerProbeRepository) CreateProbe(_ context.Context, input appprobe.CreateProbeStorageInput) (domainprobe.Probe, error) {
	r.gotCreateInput = input
	if r.createErr != nil {
		return domainprobe.Probe{}, r.createErr
	}
	return domainprobe.Probe{
		ID:        testProbeID,
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Enabled:   input.Enabled,
		City:      input.City,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		Labels: []domainprobe.Label{{
			ID:        testLabelID,
			ProjectID: input.ProjectID,
			Key:       "region",
			Value:     "tokyo",
		}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type handlerSecretGenerator struct {
	plaintext string
	hash      string
	err       error
}

func (g handlerSecretGenerator) GenerateProbeSecret() (string, string, error) {
	if g.err != nil {
		return "", "", g.err
	}
	return g.plaintext, g.hash, nil
}
