package check

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testCheckID   = "33333333-3333-3333-3333-333333333333"
	testLabelID   = "44444444-4444-4444-4444-444444444444"
)

func TestListChecksReturnsProjectChecks(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerCheckRepository{
		checks: []domaincheck.Check{newHandlerCheck("api-latency")},
	}
	NewHandler(appcheck.NewService(repo, &handlerProjectAccess{role: domainproject.RoleViewer}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/checks", "Authorization: Bearer valid-token")
	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body listChecksOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Checks) != 1 || body.Checks[0].ID != testCheckID {
		t.Fatalf("expected checks response, got %#v", body.Checks)
	}
}

func TestCreateCheckReturnsCreatedCheck(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerCheckRepository{}
	NewHandler(appcheck.NewService(repo, &handlerProjectAccess{role: domainproject.RoleEditor}, &handlerLabelAccess{
		labels: []domainlabel.Label{newHandlerLabel()},
	}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/checks", map[string]any{
		"name":            " api-latency ",
		"type":            "ping",
		"target":          " api.netstamp.io ",
		"intervalSeconds": 30,
		"labelIds":        []string{testLabelID},
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body checkOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Check.Name != "api-latency" || body.Check.Target != "api.netstamp.io" {
		t.Fatalf("expected normalized check response, got %#v", body.Check)
	}
	if len(body.Check.Labels) != 1 || body.Check.Labels[0].ID != testLabelID {
		t.Fatalf("expected labels response, got %#v", body.Check.Labels)
	}
}

func TestCheckRoutesRequireBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/checks")
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestCreateCheckMapsForbiddenToForbidden(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{role: domainproject.RoleViewer}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/checks", map[string]any{
		"name":            "api-latency",
		"type":            "ping",
		"target":          "api.netstamp.io",
		"intervalSeconds": 30,
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}
}

func TestCreateCheckReturnsFieldErrorForBlankName(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{role: domainproject.RoleEditor}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/checks", map[string]any{
		"name":            "   ",
		"type":            "ping",
		"target":          "api.netstamp.io",
		"intervalSeconds": 30,
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertHumaErrorDetail(t, res, "body.name", "must not be blank")
}

func TestCreateCheckReturnsFieldErrorForInvalidSelector(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{role: domainproject.RoleEditor}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/checks", map[string]any{
		"name":            "api-latency",
		"type":            "ping",
		"target":          "api.netstamp.io",
		"selector":        map[string]any{"label": "edge"},
		"intervalSeconds": 30,
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertHumaErrorDetail(t, res, "body.selector", "must be a valid selector")
}

func TestCreateCheckReturnsFieldErrorForInvalidLabelIDs(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{role: domainproject.RoleEditor}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/checks", map[string]any{
		"name":            "api-latency",
		"type":            "ping",
		"target":          "api.netstamp.io",
		"intervalSeconds": 30,
		"labelIds":        []string{"not-a-uuid"},
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertHumaErrorDetail(t, res, "body.labelIds", "must contain valid UUIDs")
}

func TestUpdateCheckMapsEmptyPatchToUnprocessableEntity(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appcheck.NewService(&handlerCheckRepository{}, &handlerProjectAccess{role: domainproject.RoleAdmin}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Patch("/projects/engineering/checks/"+testCheckID, map[string]any{}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertHumaErrorDetail(t, res, "body", "at least one field must be provided")
}

func TestDeleteCheckReturnsNoContent(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerCheckRepository{}
	NewHandler(appcheck.NewService(repo, &handlerProjectAccess{role: domainproject.RoleAdmin}, &handlerLabelAccess{}, handlerCheckEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Delete("/projects/engineering/checks/"+testCheckID, "Authorization: Bearer valid-token")
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if repo.gotDeleteCheckID != testCheckID {
		t.Fatalf("expected deleted check id, got %q", repo.gotDeleteCheckID)
	}
}

func assertHumaErrorDetail(t *testing.T, res *httptest.ResponseRecorder, wantLocation, wantMessage string) {
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

func (v *handlerTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	if v.err != nil {
		return identity.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

type handlerCheckEventRecorder struct{}

func (handlerCheckEventRecorder) RecordCheckEvent(context.Context, appcheck.CheckEvent) {}

type handlerCheckRepository struct {
	checks           []domaincheck.Check
	check            domaincheck.Check
	createErr        error
	deleteErr        error
	gotDeleteCheckID string
}

func (r *handlerCheckRepository) ListChecks(context.Context, string) ([]domaincheck.Check, error) {
	return r.checks, nil
}

func (r *handlerCheckRepository) GetCheck(context.Context, string, string) (domaincheck.Check, error) {
	if r.check.ID != "" {
		return r.check, nil
	}
	return newHandlerCheck("api-latency"), nil
}

func (r *handlerCheckRepository) CreateCheck(_ context.Context, input domaincheck.CreateCheckStorageInput) (domaincheck.Check, error) {
	if r.createErr != nil {
		return domaincheck.Check{}, r.createErr
	}
	check := newHandlerCheck(input.Name)
	check.Target = input.Target
	check.Selector = input.Selector
	check.PingConfig = input.PingConfig
	return check, nil
}

func (r *handlerCheckRepository) UpdateCheck(_ context.Context, input domaincheck.UpdateCheckStorageInput) (domaincheck.Check, error) {
	check := newHandlerCheck(input.Name)
	check.Target = input.Target
	check.Selector = input.Selector
	check.PingConfig = input.PingConfig
	return check, nil
}

func (r *handlerCheckRepository) SoftDeleteCheck(_ context.Context, _, checkID string) error {
	r.gotDeleteCheckID = checkID
	return r.deleteErr
}

type handlerProjectAccess struct {
	role domainproject.Role
	err  error
}

func (r *handlerProjectAccess) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	if r.err != nil {
		return domainproject.Project{}, r.err
	}
	return domainproject.Project{ID: testProjectID, Slug: "engineering"}, nil
}

func (r *handlerProjectAccess) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.role != "" {
		return r.role, nil
	}
	return domainproject.RoleOwner, nil
}

type handlerLabelAccess struct {
	labels []domainlabel.Label
	err    error
}

func (r *handlerLabelAccess) GetActiveLabelsByIDsForProject(context.Context, string, []string) ([]domainlabel.Label, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.labels, nil
}

func newHandlerCheck(name string) domaincheck.Check {
	return domaincheck.Check{
		ID:              testCheckID,
		ProjectID:       testProjectID,
		Name:            name,
		Type:            domaincheck.TypePing,
		Target:          "api.netstamp.io",
		Selector:        json.RawMessage(`{}`),
		IntervalSeconds: 30,
		PingConfig: domainping.Config{
			PacketCount:     4,
			PacketSizeBytes: 56,
			TimeoutMs:       3000,
		},
		Labels:    []domainlabel.Label{newHandlerLabel()},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newHandlerLabel() domainlabel.Label {
	return domainlabel.Label{
		ID:        testLabelID,
		ProjectID: testProjectID,
		Key:       "region",
		Value:     "tokyo",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
