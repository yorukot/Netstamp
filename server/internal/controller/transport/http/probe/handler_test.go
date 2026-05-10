package probe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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
	projectAccess := &handlerProjectAccess{}
	NewHandler(appprobe.NewService(repo, projectAccess, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
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
	if projectAccess.gotProjectRef != "engineering" {
		t.Fatalf("expected project ref, got %q", projectAccess.gotProjectRef)
	}
}

func TestCreateProbeRequiresBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name": "tokyo-vps-1",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestListProbeReturnsProjectProbes(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/probes", "Authorization: Bearer valid-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	var body listProbesOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Probes) != 1 || body.Probes[0].ID != testProbeID {
		t.Fatalf("expected probe list, got %#v", body.Probes)
	}
}

func TestGetProbeReturnsProjectProbe(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/probes/"+testProbeID, "Authorization: Bearer valid-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	var body probeOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Probe.ID != testProbeID {
		t.Fatalf("expected probe id, got %q", body.Probe.ID)
	}
}

func TestUpdateProbeReturnsUpdatedProbe(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{}
	NewHandler(appprobe.NewService(repo, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Patch("/projects/engineering/probes/"+testProbeID, map[string]any{
		"name":    " osaka-vps-1 ",
		"enabled": false,
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if repo.gotUpdateInput.Name != "osaka-vps-1" {
		t.Fatalf("expected update name, got %q", repo.gotUpdateInput.Name)
	}
	if repo.gotUpdateInput.Enabled {
		t.Fatalf("expected enabled false")
	}
}

func TestDeleteProbeReturnsNoContent(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{}
	NewHandler(appprobe.NewService(repo, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Delete("/projects/engineering/probes/"+testProbeID, "Authorization: Bearer valid-token")

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if repo.gotDelete.probeID != testProbeID {
		t.Fatalf("expected deleted probe id, got %#v", repo.gotDelete)
	}
}

func TestRotateProbeSecretReturnsSecretOnce(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{}
	NewHandler(appprobe.NewService(repo, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "new-plain-secret",
		hash:      "new-secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes/"+testProbeID+"/secret-rotations", map[string]any{}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	var body rotateSecretOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Secret != "new-plain-secret" {
		t.Fatalf("expected secret, got %q", body.Secret)
	}
	if repo.gotRotate.SecretHash != "new-secret-hash" {
		t.Fatalf("expected hash, got %q", repo.gotRotate.SecretHash)
	}
}

func TestCreateProbeMapsInvalidInputToUnprocessableEntity(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name": "   ",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertProbeHumaErrorDetail(t, res, "body.name", "must not be blank")
}

func TestCreateProbeMapsInaccessibleProjectToNotFound(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{
		err: appprobe.ErrProjectNotFound,
	}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/missing/probes", map[string]any{
		"name": "tokyo-vps-1",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

func TestCreateProbeMapsForbiddenToForbidden(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{
		role: domainproject.RoleViewer,
	}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name": "tokyo-vps-1",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}
}

func TestCreateProbeMapsMissingLabelToNotFound(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{
		err: appprobe.ErrLabelNotFound,
	}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name":     "tokyo-vps-1",
		"labelIds": []string{testLabelID},
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

func TestCreateProbeReturnsFieldErrorForMissingCoordinatePair(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appprobe.NewService(&handlerProbeRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, handlerProbeEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/probes", map[string]any{
		"name":     "tokyo-vps-1",
		"latitude": 35.6762,
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertProbeHumaErrorDetail(t, res, "body.longitude", "must be provided with latitude")
}

func assertProbeHumaErrorDetail(t *testing.T, res *httptest.ResponseRecorder, wantLocation, wantMessage string) {
	t.Helper()

	var body huma.ErrorModel
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	for _, detail := range body.Errors {
		if detail.Location == wantLocation && detail.Message == wantMessage {
			return
		}
	}

	t.Fatalf("expected error detail %q/%q, got %#v", wantLocation, wantMessage, body.Errors)
}

type handlerTokenVerifier struct {
	claims identity.AccessTokenClaims
	err    error
}

type handlerProbeEventRecorder struct{}

func (handlerProbeEventRecorder) RecordProbeEvent(context.Context, appprobe.ProbeEvent) {}

func (v *handlerTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	if v.err != nil {
		return identity.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

type handlerProbeRepository struct {
	gotCreateInput domainprobe.CreateProbeStorageInput
	createErr      error
	gotUpdateInput domainprobe.UpdateProbeStorageInput
	gotDelete      struct {
		projectID string
		probeID   string
	}
	gotRotate domainprobe.RotateProbeSecretStorageInput
}

func (r *handlerProbeRepository) CreateProbe(_ context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error) {
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *handlerProbeRepository) ListProbesForProject(context.Context, string) ([]domainprobe.Probe, error) {
	return []domainprobe.Probe{handlerProbe()}, nil
}

func (r *handlerProbeRepository) GetProbeForProject(context.Context, string, string) (domainprobe.Probe, error) {
	return handlerProbe(), nil
}

func (r *handlerProbeRepository) UpdateProbe(_ context.Context, input domainprobe.UpdateProbeStorageInput) (domainprobe.Probe, error) {
	r.gotUpdateInput = input
	probe := handlerProbe()
	probe.Name = input.Name
	probe.Enabled = input.Enabled
	probe.City = input.City
	probe.Latitude = input.Latitude
	probe.Longitude = input.Longitude
	return probe, nil
}

func (r *handlerProbeRepository) SoftDeleteProbe(_ context.Context, projectID, probeID string) error {
	r.gotDelete.projectID = projectID
	r.gotDelete.probeID = probeID
	return nil
}

func (r *handlerProbeRepository) RotateProbeSecret(_ context.Context, input domainprobe.RotateProbeSecretStorageInput) error {
	r.gotRotate = input
	return nil
}

func handlerProbe() domainprobe.Probe {
	now := time.Now()
	return domainprobe.Probe{
		ID:        testProbeID,
		ProjectID: testProjectID,
		Name:      "tokyo-vps-1",
		Enabled:   true,
		Labels: []domainlabel.Label{{
			ID:        testLabelID,
			ProjectID: testProjectID,
			Key:       "region",
			Value:     "tokyo",
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type handlerProjectAccess struct {
	gotProjectRef string
	role          domainproject.Role
	err           error
	roleErr       error
}

func (r *handlerProjectAccess) GetProjectForUser(_ context.Context, projectRef, _ string) (domainproject.Project, error) {
	r.gotProjectRef = projectRef
	if r.err != nil {
		return domainproject.Project{}, r.err
	}
	return domainproject.Project{ID: testProjectID, Slug: "engineering"}, nil
}

func (r *handlerProjectAccess) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	if r.roleErr != nil {
		return "", r.roleErr
	}
	if r.role != "" {
		return r.role, nil
	}
	return domainproject.RoleOwner, nil
}

type handlerLabelAccess struct {
	err error
}

func (r *handlerLabelAccess) GetActiveLabelsByIDsForProject(_ context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(labelIDs) == 0 {
		return []domainlabel.Label{}, nil
	}
	return []domainlabel.Label{{
		ID:        testLabelID,
		ProjectID: projectID,
		Key:       "region",
		Value:     "tokyo",
	}}, nil
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
