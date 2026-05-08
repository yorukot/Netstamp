package label

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	applabel "github.com/yorukot/netstamp/internal/application/label"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testLabelID   = "33333333-3333-3333-3333-333333333333"
)

func TestListLabelsReturnsProjectLabels(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerLabelRepository{
		labels: []domainlabel.Label{newHandlerLabel("region", "tokyo")},
	}
	NewHandler(applabel.NewService(repo, &handlerProjectAccess{role: domainproject.RoleViewer}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/labels", "Authorization: Bearer valid-token")
	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body listLabelsOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Labels) != 1 || body.Labels[0].ID != testLabelID {
		t.Fatalf("expected labels response, got %#v", body.Labels)
	}
}

func TestCreateLabelReturnsCreatedLabel(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerLabelRepository{}
	NewHandler(applabel.NewService(repo, &handlerProjectAccess{role: domainproject.RoleEditor}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/labels", map[string]any{
		"key":   " region ",
		"value": " tokyo ",
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body labelOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Label.Key != "region" || body.Label.Value != "tokyo" {
		t.Fatalf("expected normalized label response, got %#v", body.Label)
	}
}

func TestLabelRoutesRequireBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(applabel.NewService(&handlerLabelRepository{}, &handlerProjectAccess{}), &handlerTokenVerifier{}).RegisterRoutes(api)

	res := api.Get("/projects/engineering/labels")
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestCreateLabelMapsForbiddenToForbidden(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(applabel.NewService(&handlerLabelRepository{}, &handlerProjectAccess{role: domainproject.RoleViewer}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/labels", map[string]any{
		"key":   "region",
		"value": "tokyo",
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}
}

func TestCreateLabelMapsDuplicateToConflict(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(applabel.NewService(&handlerLabelRepository{createErr: applabel.ErrLabelAlreadyExists}, &handlerProjectAccess{role: domainproject.RoleOwner}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/labels", map[string]any{
		"key":   "region",
		"value": "tokyo",
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", res.Code)
	}
}

func TestUpdateLabelMapsEmptyPatchToUnprocessableEntity(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(applabel.NewService(&handlerLabelRepository{}, &handlerProjectAccess{role: domainproject.RoleAdmin}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Patch("/projects/engineering/labels/"+testLabelID, map[string]any{}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestUpdateLabelReturnsUpdatedLabel(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerLabelRepository{
		label: newHandlerLabel("region", "tokyo"),
	}
	NewHandler(applabel.NewService(repo, &handlerProjectAccess{role: domainproject.RoleAdmin}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Patch("/projects/engineering/labels/"+testLabelID, map[string]any{
		"value": "osaka",
	}, "Authorization: Bearer valid-token")
	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body labelOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Label.Key != "region" || body.Label.Value != "osaka" {
		t.Fatalf("expected partial update response, got %#v", body.Label)
	}
}

func TestDeleteLabelReturnsNoContent(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerLabelRepository{}
	NewHandler(applabel.NewService(repo, &handlerProjectAccess{role: domainproject.RoleAdmin}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Delete("/projects/engineering/labels/"+testLabelID, "Authorization: Bearer valid-token")
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if repo.gotDeleteLabelID != testLabelID {
		t.Fatalf("expected deleted label id, got %q", repo.gotDeleteLabelID)
	}
}

func TestDeleteLabelMapsMissingLabelToNotFound(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(applabel.NewService(&handlerLabelRepository{deleteErr: applabel.ErrLabelNotFound}, &handlerProjectAccess{role: domainproject.RoleOwner}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Delete("/projects/engineering/labels/"+testLabelID, "Authorization: Bearer valid-token")
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
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

type handlerLabelRepository struct {
	labels           []domainlabel.Label
	label            domainlabel.Label
	createErr        error
	deleteErr        error
	gotDeleteLabelID string
}

func (r *handlerLabelRepository) ListLabels(context.Context, string) ([]domainlabel.Label, error) {
	return r.labels, nil
}

func (r *handlerLabelRepository) GetLabel(context.Context, string, string) (domainlabel.Label, error) {
	if r.label.ID != "" {
		return r.label, nil
	}
	return newHandlerLabel("region", "tokyo"), nil
}

func (r *handlerLabelRepository) CreateLabel(_ context.Context, input domainlabel.CreateLabelStorageInput) (domainlabel.Label, error) {
	if r.createErr != nil {
		return domainlabel.Label{}, r.createErr
	}
	return newHandlerLabel(input.Key, input.Value), nil
}

func (r *handlerLabelRepository) UpdateLabel(_ context.Context, input domainlabel.UpdateLabelStorageInput) (domainlabel.Label, error) {
	return newHandlerLabel(input.Key, input.Value), nil
}

func (r *handlerLabelRepository) SoftDeleteLabel(_ context.Context, _ string, labelID string) error {
	r.gotDeleteLabelID = labelID
	return r.deleteErr
}

func (r *handlerLabelRepository) GetActiveLabelsByIDsForProject(context.Context, string, []string) ([]domainlabel.Label, error) {
	return r.labels, nil
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

func newHandlerLabel(key string, value string) domainlabel.Label {
	return domainlabel.Label{
		ID:        testLabelID,
		ProjectID: testProjectID,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
